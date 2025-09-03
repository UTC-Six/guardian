package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/UTC-Six/guardian/internal/algorithm"
	"github.com/UTC-Six/guardian/internal/types"
)

func TestSimpleDemo(t *testing.T) {
	// 测试AC自动机
	fmt.Println("=== AC自动机测试 ===")

	ac := algorithm.NewACAutomaton()

	// 添加敏感词
	ac.AddWord("敏感词1", []string{"abuse"}, 3)
	ac.AddWord("敏感词2", []string{"politics"}, 4)
	ac.AddWord("辱骂词", []string{"abuse"}, 5)

	// 构建失败指针
	ac.BuildFailPointers()

	// 测试搜索
	testTexts := []string{
		"这是一段正常文本",
		"包含敏感词1的文本",
		"敏感词1和敏感词2都存在",
		"辱骂词不应该出现",
	}

	for _, text := range testTexts {
		results := ac.Search(text)
		fmt.Printf("文本: %s\n", text)
		fmt.Printf("匹配数: %d\n", len(results))
		for _, result := range results {
			fmt.Printf("  - 词: %s, 分类: %v, 级别: %d\n",
				result.Word, result.Categories, result.Level)
		}
		fmt.Println("---")
	}

	// 测试带选项的搜索
	fmt.Println("\n=== 带选项搜索测试 ===")
	options := &algorithm.SearchOptions{
		Categories: []string{"abuse"},
		MinLevel:   3,
	}

	results := ac.SearchWithOptions("敏感词1和敏感词2", options)
	fmt.Printf("只匹配abuse分类且级别>=3的结果数: %d\n", len(results))
	for _, result := range results {
		fmt.Printf("  - 词: %s, 分类: %v, 级别: %d\n",
			result.Word, result.Categories, result.Level)
	}

	// 测试统计信息
	fmt.Println("\n=== 统计信息 ===")
	fmt.Printf("节点数: %d\n", ac.GetNodeCount())
	fmt.Printf("版本: %s\n", ac.GetVersion())

	// 测试词库结构
	fmt.Println("\n=== 词库结构测试 ===")
	wordDB := &types.WordDatabase{
		Version:    "1.0.0",
		UpdateTime: time.Now(),
		Whitelist:  []string{"正常词汇1", "正常词汇2"},
		Blacklist: []types.SensitiveWord{
			{
				Word:       "敏感词1",
				Categories: []string{"abuse"},
				Level:      3,
			},
		},
		Categories: map[string][]types.SensitiveWord{
			"abuse": {
				{
					Word:       "辱骂词1",
					Categories: []string{"abuse"},
					Level:      4,
				},
			},
		},
		Replacements: map[string]string{
			"敏感词1": "***",
		},
	}

	fmt.Printf("词库版本: %s\n", wordDB.Version)
	fmt.Printf("白名单大小: %d\n", len(wordDB.Whitelist))
	fmt.Printf("黑名单大小: %d\n", len(wordDB.Blacklist))
	fmt.Printf("分类数: %d\n", len(wordDB.Categories))

	// 测试过滤结果
	fmt.Println("\n=== 过滤结果测试 ===")
	result := &types.FilterResult{
		Passed:     false,
		Categories: []string{"abuse", "politics"},
		Words:      []string{"敏感词1", "敏感词2"},
		Details: map[string]string{
			"敏感词1": "level:3,categories:abuse",
			"敏感词2": "level:4,categories:politics",
		},
	}

	fmt.Printf("通过: %v\n", result.Passed)
	fmt.Printf("分类: %v\n", result.Categories)
	fmt.Printf("敏感词: %v\n", result.Words)
	fmt.Printf("详情: %v\n", result.Details)

	fmt.Println("\n测试完成！")
}
