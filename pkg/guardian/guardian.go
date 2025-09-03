package guardian

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/guardian/content-filter/internal/filter"
	"github.com/guardian/content-filter/internal/nacos"
	"github.com/guardian/content-filter/internal/types"
)

// Guardian 黄反校验SDK主入口
type Guardian struct {
	filter *filter.ContentFilter
	logger *logrus.Logger
}

// NewGuardian 创建新的Guardian实例
func NewGuardian(config *types.Config) (*Guardian, error) {
	// 初始化日志
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// 创建Nacos客户端
	nacosClient, err := nacos.NewClient(&config.NacosConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create nacos client: %w", err)
	}

	// 创建内容过滤器
	contentFilter, err := filter.NewContentFilter(nacosClient, &config.FilterConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create content filter: %w", err)
	}

	return &Guardian{
		filter: contentFilter,
		logger: logger,
	}, nil
}

// NewGuardianWithLogger 使用自定义日志创建Guardian实例
func NewGuardianWithLogger(config *types.Config, logger *logrus.Logger) (*Guardian, error) {
	// 创建Nacos客户端
	nacosClient, err := nacos.NewClient(&config.NacosConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create nacos client: %w", err)
	}

	// 创建内容过滤器
	contentFilter, err := filter.NewContentFilter(nacosClient, &config.FilterConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create content filter: %w", err)
	}

	return &Guardian{
		filter: contentFilter,
		logger: logger,
	}, nil
}

// Check 检查文本内容
func (g *Guardian) Check(text string) *types.FilterResult {
	return g.CheckWithOptions(text, &types.FilterOptions{
		EnableWhitelist: true,
		Categories:      []string{},
		MinLevel:        1,
		ReplaceMode:     false,
	})
}

// CheckWithOptions 带选项检查文本内容
func (g *Guardian) CheckWithOptions(text string, options *types.FilterOptions) *types.FilterResult {
	return g.filter.Filter(text, options)
}

// CheckCategory 检查特定分类的敏感词
func (g *Guardian) CheckCategory(text string, categories []string) *types.FilterResult {
	return g.CheckWithOptions(text, &types.FilterOptions{
		EnableWhitelist: true,
		Categories:      categories,
		MinLevel:        1,
		ReplaceMode:     false,
	})
}

// CheckLevel 检查特定级别的敏感词
func (g *Guardian) CheckLevel(text string, minLevel int) *types.FilterResult {
	return g.CheckWithOptions(text, &types.FilterOptions{
		EnableWhitelist: true,
		Categories:      []string{},
		MinLevel:        minLevel,
		ReplaceMode:     false,
	})
}

// IsSafe 检查文本是否安全
func (g *Guardian) IsSafe(text string) bool {
	result := g.Check(text)
	return result.Passed
}

// GetMatchedWords 获取匹配的敏感词
func (g *Guardian) GetMatchedWords(text string) []string {
	result := g.Check(text)
	return result.Words
}

// GetMatchedCategories 获取匹配的分类
func (g *Guardian) GetMatchedCategories(text string) []string {
	result := g.Check(text)
	return result.Categories
}

// GetStats 获取统计信息
func (g *Guardian) GetStats() map[string]interface{} {
	return g.filter.GetStats()
}

// HealthCheck 健康检查
func (g *Guardian) HealthCheck() error {
	return g.filter.HealthCheck()
}

// Close 关闭Guardian
func (g *Guardian) Close() error {
	return g.filter.Close()
}

// BatchCheck 批量检查
func (g *Guardian) BatchCheck(texts []string) []*types.FilterResult {
	results := make([]*types.FilterResult, len(texts))
	for i, text := range texts {
		results[i] = g.Check(text)
	}
	return results
}

// BatchCheckWithOptions 带选项批量检查
func (g *Guardian) BatchCheckWithOptions(texts []string, options *types.FilterOptions) []*types.FilterResult {
	results := make([]*types.FilterResult, len(texts))
	for i, text := range texts {
		results[i] = g.CheckWithOptions(text, options)
	}
	return results
}

// UpdateWordDatabase 更新词库
func (g *Guardian) UpdateWordDatabase(wordDB *types.WordDatabase) error {
	return g.filter.UpdateWordDatabase(wordDB)
}

// AddToWhitelist 添加到白名单
func (g *Guardian) AddToWhitelist(word string) {
	g.filter.AddToWhitelist(word)
}

// RemoveFromWhitelist 从白名单移除
func (g *Guardian) RemoveFromWhitelist(word string) {
	g.filter.RemoveFromWhitelist(word)
}

// SetLogger 设置日志器
func (g *Guardian) SetLogger(logger *logrus.Logger) {
	g.logger = logger
}

// GetLogger 获取日志器
func (g *Guardian) GetLogger() *logrus.Logger {
	return g.logger
}
