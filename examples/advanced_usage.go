package main

import (
	"fmt"
	"log"
	"time"

	"github.com/UTC-Six/guardian/internal/types"
	"github.com/UTC-Six/guardian/pkg/guardian"
)

func main() {
	// 创建配置
	config := &types.Config{
		NacosConfig: types.NacosConfig{
			ServerConfigs: []types.ServerConfig{
				{
					IpAddr: "127.0.0.1",
					Port:   8848,
				},
			},
			ClientConfig: types.ClientConfig{
				NamespaceId:         "public",
				TimeoutMs:           5000,
				NotLoadCacheAtStart: false,
				LogDir:              "./logs",
				CacheDir:            "./cache",
				LogLevel:            "info",
			},
		},
		FilterConfig: types.FilterConfig{
			DataId:          "sensitive_words",
			Group:           "DEFAULT_GROUP",
			ReloadPeriod:    5 * time.Minute,
			EnableCache:     true,
			CacheSize:       10000,
			EnableWhitelist: true,
		},
	}

	// 创建Guardian实例
	g, err := guardian.NewGuardian(config)
	if err != nil {
		log.Fatalf("Failed to create Guardian: %v", err)
	}
	defer g.Close()

	// 高级选项检查
	fmt.Println("=== 高级选项检查 ===")

	// 自定义选项
	options := &types.FilterOptions{
		EnableWhitelist: true,
		Categories:      []string{"abuse", "politics"},
		MinLevel:        3,
		ReplaceMode:     false,
	}

	testText := "这段文本包含辱骂词1和政治敏感词1"
	result := g.CheckWithOptions(testText, options)

	fmt.Printf("文本: %s\n", testText)
	fmt.Printf("通过: %v\n", result.Passed)
	if !result.Passed {
		fmt.Printf("匹配词: %v\n", result.Words)
		fmt.Printf("分类: %v\n", result.Categories)
		fmt.Printf("详情: %v\n", result.Details)
	}

	// 动态添加白名单
	fmt.Println("\n=== 动态白名单管理 ===")

	// 添加白名单
	g.AddToWhitelist("特殊词汇")

	// 测试白名单效果
	whitelistResult := g.Check("特殊词汇不应该被过滤")
	fmt.Printf("白名单测试: %v\n", whitelistResult.Passed)

	// 移除白名单
	g.RemoveFromWhitelist("特殊词汇")

	// 再次测试
	whitelistResult2 := g.Check("特殊词汇不应该被过滤")
	fmt.Printf("移除白名单后: %v\n", whitelistResult2.Passed)

	// 动态更新词库
	fmt.Println("\n=== 动态词库更新 ===")

	// 创建新的词库
	newWordDB := &types.WordDatabase{
		Version:    "1.1.0",
		UpdateTime: time.Now(),
		Whitelist:  []string{"正常词汇1", "正常词汇2", "测试词汇"},
		Blacklist: []types.SensitiveWord{
			{
				Word:       "新敏感词1",
				Categories: []string{"abuse"},
				Level:      3,
			},
		},
		Categories: map[string][]types.SensitiveWord{
			"abuse": {
				{
					Word:       "新辱骂词1",
					Categories: []string{"abuse"},
					Level:      4,
				},
			},
		},
		Replacements: map[string]string{
			"新敏感词1": "***",
		},
	}

	// 更新词库
	if err := g.UpdateWordDatabase(newWordDB); err != nil {
		fmt.Printf("更新词库失败: %v\n", err)
	} else {
		fmt.Println("词库更新成功")

		// 测试新词库
		newResult := g.Check("新敏感词1")
		fmt.Printf("新敏感词检查: %v\n", newResult.Passed)
		if !newResult.Passed {
			fmt.Printf("匹配词: %v\n", newResult.Words)
		}
	}

	// 性能测试
	fmt.Println("\n=== 性能测试 ===")

	testTexts := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		testTexts[i] = fmt.Sprintf("测试文本%d", i)
	}

	start := time.Now()
	results := g.BatchCheck(testTexts)
	duration := time.Since(start)

	passedCount := 0
	for _, result := range results {
		if result.Passed {
			passedCount++
		}
	}

	fmt.Printf("处理1000个文本耗时: %v\n", duration)
	fmt.Printf("平均每个文本耗时: %v\n", duration/1000)
	fmt.Printf("通过数量: %d\n", passedCount)
	fmt.Printf("拒绝数量: %d\n", 1000-passedCount)

	// 获取最终统计信息
	fmt.Println("\n=== 最终统计信息 ===")
	stats := g.GetStats()
	fmt.Printf("版本: %v\n", stats["version"])
	fmt.Printf("节点数: %v\n", stats["node_count"])
	fmt.Printf("白名单大小: %v\n", stats["whitelist_size"])
	if cacheStats, ok := stats["cache_stats"]; ok {
		fmt.Printf("缓存统计: %v\n", cacheStats)
	}
}
