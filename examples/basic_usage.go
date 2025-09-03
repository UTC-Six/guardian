package main

import (
	"encoding/json"
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

	// 基本检查
	fmt.Println("=== 基本检查 ===")
	testTexts := []string{
		"这是一段正常的文本",
		"这段文本包含敏感词1",
		"测试词汇是正常的",
		"辱骂词1不应该出现",
		"政治敏感词1很危险",
	}

	for _, text := range testTexts {
		result := g.Check(text)
		fmt.Printf("文本: %s\n", text)
		fmt.Printf("通过: %v\n", result.Passed)
		if !result.Passed {
			fmt.Printf("匹配词: %v\n", result.Words)
			fmt.Printf("分类: %v\n", result.Categories)
		}
		fmt.Println("---")
	}

	// 分类检查
	fmt.Println("\n=== 分类检查 ===")
	abuseResult := g.CheckCategory("辱骂词1不应该出现", []string{"abuse"})
	fmt.Printf("辱骂分类检查: %v\n", abuseResult.Passed)
	if !abuseResult.Passed {
		fmt.Printf("匹配词: %v\n", abuseResult.Words)
	}

	politicsResult := g.CheckCategory("政治敏感词1很危险", []string{"politics"})
	fmt.Printf("政治分类检查: %v\n", politicsResult.Passed)
	if !politicsResult.Passed {
		fmt.Printf("匹配词: %v\n", politicsResult.Words)
	}

	// 级别检查
	fmt.Println("\n=== 级别检查 ===")
	level3Result := g.CheckLevel("敏感词1", 3)
	fmt.Printf("级别3检查: %v\n", level3Result.Passed)

	level5Result := g.CheckLevel("政治敏感词1", 5)
	fmt.Printf("级别5检查: %v\n", level5Result.Passed)

	// 批量检查
	fmt.Println("\n=== 批量检查 ===")
	batchTexts := []string{
		"正常文本1",
		"敏感词1",
		"正常文本2",
		"辱骂词1",
	}

	batchResults := g.BatchCheck(batchTexts)
	for i, result := range batchResults {
		fmt.Printf("文本%d: %s, 通过: %v\n", i+1, batchTexts[i], result.Passed)
	}

	// 获取统计信息
	fmt.Println("\n=== 统计信息 ===")
	stats := g.GetStats()
	statsJSON, _ := json.MarshalIndent(stats, "", "  ")
	fmt.Printf("统计信息: %s\n", statsJSON)

	// 健康检查
	fmt.Println("\n=== 健康检查 ===")
	if err := g.HealthCheck(); err != nil {
		fmt.Printf("健康检查失败: %v\n", err)
	} else {
		fmt.Println("健康检查通过")
	}
}
