package filter

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/guardian/content-filter/internal/algorithm"
	"github.com/guardian/content-filter/internal/cache"
	"github.com/guardian/content-filter/internal/nacos"
	"github.com/guardian/content-filter/internal/types"
)

// ContentFilter 内容过滤器
type ContentFilter struct {
	automaton    *algorithm.ACAutomaton
	nacosClient  *nacos.Client
	cache        cache.Cache
	config       *types.FilterConfig
	logger       *logrus.Logger
	whitelist    map[string]bool
	mu           sync.RWMutex
	lastUpdate   time.Time
	version      string
	stopChan     chan struct{}
	reloadTicker *time.Ticker
}

// NewContentFilter 创建新的内容过滤器
func NewContentFilter(nacosClient *nacos.Client, config *types.FilterConfig, logger *logrus.Logger) (*ContentFilter, error) {
	filter := &ContentFilter{
		automaton:   algorithm.NewACAutomaton(),
		nacosClient: nacosClient,
		config:      config,
		logger:      logger,
		whitelist:   make(map[string]bool),
		stopChan:    make(chan struct{}),
	}

	// 初始化缓存
	if config.EnableCache {
		filter.cache = cache.NewLRUCache(config.CacheSize, 10*time.Minute)
	}

	// 加载初始配置
	if err := filter.loadWordDatabase(); err != nil {
		return nil, fmt.Errorf("failed to load initial word database: %w", err)
	}

	// 启动配置监听
	if err := filter.startConfigListener(); err != nil {
		return nil, fmt.Errorf("failed to start config listener: %w", err)
	}

	// 启动定期重载
	filter.startPeriodicReload()

	return filter, nil
}

// loadWordDatabase 加载词库
func (f *ContentFilter) loadWordDatabase() error {
	wordDB, err := f.nacosClient.GetWordDatabase(f.config.DataId, f.config.Group)
	if err != nil {
		return fmt.Errorf("failed to get word database from nacos: %w", err)
	}

	return f.updateWordDatabase(wordDB)
}

// updateWordDatabase 更新词库
func (f *ContentFilter) updateWordDatabase(wordDB *types.WordDatabase) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// 清空现有数据
	f.automaton.Clear()
	f.whitelist = make(map[string]bool)

	// 更新白名单
	for _, word := range wordDB.Whitelist {
		f.whitelist[strings.ToLower(word)] = true
	}

	// 更新黑名单
	for _, word := range wordDB.Blacklist {
		f.automaton.AddWord(word.Word, word.Categories, word.Level)
	}

	// 更新分类敏感词
	for _, words := range wordDB.Categories {
		for _, word := range words {
			f.automaton.AddWord(word.Word, word.Categories, word.Level)
		}
	}

	// 构建AC自动机
	f.automaton.BuildFailPointers()
	f.automaton.SetVersion(wordDB.Version)

	// 更新版本和时间
	f.version = wordDB.Version
	f.lastUpdate = wordDB.UpdateTime

	// 清空缓存
	if f.cache != nil {
		f.cache.Clear()
	}

	f.logger.Infof("Word database updated successfully, version: %s, words: %d", 
		wordDB.Version, f.automaton.GetNodeCount())

	return nil
}

// startConfigListener 启动配置监听
func (f *ContentFilter) startConfigListener() error {
	return f.nacosClient.ListenConfig(f.config.DataId, f.config.Group, func(content string) {
		f.logger.Info("Received config change notification")
		
		// 解析新的词库配置
		var wordDB types.WordDatabase
		if err := json.Unmarshal([]byte(content), &wordDB); err != nil {
			f.logger.Errorf("Failed to unmarshal word database: %v", err)
			return
		}

		// 更新词库
		if err := f.updateWordDatabase(&wordDB); err != nil {
			f.logger.Errorf("Failed to update word database: %v", err)
		}
	})
}

// startPeriodicReload 启动定期重载
func (f *ContentFilter) startPeriodicReload() {
	if f.config.ReloadPeriod <= 0 {
		return
	}

	f.reloadTicker = time.NewTicker(f.config.ReloadPeriod)
	go func() {
		for {
			select {
			case <-f.reloadTicker.C:
				if err := f.loadWordDatabase(); err != nil {
					f.logger.Errorf("Failed to reload word database: %v", err)
				}
			case <-f.stopChan:
				return
			}
		}
	}()
}

// Filter 过滤内容
func (f *ContentFilter) Filter(text string, options *types.FilterOptions) *types.FilterResult {
	// 检查缓存
	if f.cache != nil {
		cacheKey := f.generateCacheKey(text, options)
		if result, found := f.cache.Get(cacheKey); found {
			return result
		}
	}

	// 执行过滤
	result := f.doFilter(text, options)

	// 缓存结果
	if f.cache != nil {
		cacheKey := f.generateCacheKey(text, options)
		f.cache.Set(cacheKey, result)
	}

	return result
}

// doFilter 执行过滤逻辑
func (f *ContentFilter) doFilter(text string, options *types.FilterOptions) *types.FilterResult {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// 检查白名单
	if options != nil && options.EnableWhitelist && f.config.EnableWhitelist {
		if f.isInWhitelist(text) {
			return &types.FilterResult{
				Passed:     true,
				Categories: []string{},
				Words:      []string{},
				Details:    map[string]string{"reason": "whitelist"},
			}
		}
	}

	// 标准化文本
	normalizedText := algorithm.NormalizeText(text)

	// 构建搜索选项
	searchOptions := &algorithm.SearchOptions{
		Categories: options.Categories,
		MinLevel:   options.MinLevel,
	}

	// 搜索敏感词
	outputs := f.automaton.SearchWithOptions(normalizedText, searchOptions)

	if len(outputs) == 0 {
		return &types.FilterResult{
			Passed:     true,
			Categories: []string{},
			Words:      []string{},
			Details:    map[string]string{},
		}
	}

	// 收集结果
	categories := make([]string, 0)
	words := make([]string, 0)
	details := make(map[string]string)

	for _, output := range outputs {
		words = append(words, output.Word)
		categories = append(categories, output.Categories...)
		details[output.Word] = fmt.Sprintf("level:%d,categories:%s", 
			output.Level, strings.Join(output.Categories, ","))
	}

	// 去重
	categories = f.removeDuplicates(categories)
	words = f.removeDuplicates(words)

	return &types.FilterResult{
		Passed:     false,
		Categories: categories,
		Words:      words,
		Details:    details,
	}
}

// isInWhitelist 检查是否在白名单中
func (f *ContentFilter) isInWhitelist(text string) bool {
	normalizedText := strings.ToLower(algorithm.NormalizeText(text))
	
	// 检查完整文本
	if f.whitelist[normalizedText] {
		return true
	}

	// 检查文本片段
	words := strings.Fields(normalizedText)
	for _, word := range words {
		if f.whitelist[word] {
			return true
		}
	}

	return false
}

// removeDuplicates 去重
func (f *ContentFilter) removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	result := make([]string, 0)

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

// generateCacheKey 生成缓存键
func (f *ContentFilter) generateCacheKey(text string, options *types.FilterOptions) string {
	var optionsStr string
	if options != nil {
		optionsStr = fmt.Sprintf("%v", options)
	}
	
	key := fmt.Sprintf("%s:%s", text, optionsStr)
	hash := md5.Sum([]byte(key))
	return fmt.Sprintf("%x", hash)
}

// GetStats 获取统计信息
func (f *ContentFilter) GetStats() map[string]interface{} {
	f.mu.RLock()
	defer f.mu.RUnlock()

	stats := map[string]interface{}{
		"version":        f.version,
		"last_update":    f.lastUpdate,
		"node_count":     f.automaton.GetNodeCount(),
		"whitelist_size": len(f.whitelist),
	}

	if f.cache != nil {
		stats["cache_stats"] = f.cache.Stats()
	}

	return stats
}

// UpdateWordDatabase 手动更新词库
func (f *ContentFilter) UpdateWordDatabase(wordDB *types.WordDatabase) error {
	return f.updateWordDatabase(wordDB)
}

// AddToWhitelist 添加到白名单
func (f *ContentFilter) AddToWhitelist(word string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.whitelist[strings.ToLower(word)] = true
}

// RemoveFromWhitelist 从白名单移除
func (f *ContentFilter) RemoveFromWhitelist(word string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.whitelist, strings.ToLower(word))
}

// Close 关闭过滤器
func (f *ContentFilter) Close() error {
	close(f.stopChan)
	
	if f.reloadTicker != nil {
		f.reloadTicker.Stop()
	}
	
	if f.cache != nil {
		f.cache.Close()
	}
	
	return f.nacosClient.Close()
}

// HealthCheck 健康检查
func (f *ContentFilter) HealthCheck() error {
	// 检查Nacos连接
	if err := f.nacosClient.HealthCheck(); err != nil {
		return fmt.Errorf("nacos health check failed: %w", err)
	}

	// 检查自动机状态
	if f.automaton.GetNodeCount() == 0 {
		return fmt.Errorf("automaton is empty")
	}

	return nil
}
