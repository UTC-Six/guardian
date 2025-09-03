package algorithm

import (
	"sync"
	"testing"
)

// BenchmarkACAutomatonConcurrent 并发搜索性能测试
func BenchmarkACAutomatonConcurrent(b *testing.B) {
	ac := NewACAutomaton()

	// 添加大量敏感词
	words := []string{"敏感词1", "敏感词2", "敏感词3", "辱骂词1", "辱骂词2", "政治词1", "政治词2", "色情词1", "色情词2"}
	for i, word := range words {
		ac.AddWord(word, []string{"test"}, i%3+1)
	}
	ac.BuildFailPointers()

	text := "这是一段包含敏感词1和辱骂词1的测试文本，用于测试AC自动机的并发性能"

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ac.Search(text)
		}
	})
}

// BenchmarkACAutomatonSearchWithOptionsConcurrent 带选项搜索并发性能测试
func BenchmarkACAutomatonSearchWithOptionsConcurrent(b *testing.B) {
	ac := NewACAutomaton()

	// 添加大量敏感词
	words := []string{"敏感词1", "敏感词2", "敏感词3", "辱骂词1", "辱骂词2", "政治词1", "政治词2", "色情词1", "色情词2"}
	for i, word := range words {
		ac.AddWord(word, []string{"test"}, i%3+1)
	}
	ac.BuildFailPointers()

	text := "这是一段包含敏感词1和辱骂词1的测试文本，用于测试AC自动机的并发性能"
	options := &SearchOptions{
		Categories: []string{"test"},
		MinLevel:   2,
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ac.SearchWithOptions(text, options)
		}
	})
}

// TestACAutomatonConcurrentSafety 并发安全性测试
func TestACAutomatonConcurrentSafety(t *testing.T) {
	ac := NewACAutomaton()

	// 添加敏感词
	words := []string{"敏感词1", "敏感词2", "敏感词3", "辱骂词1", "辱骂词2"}
	for i, word := range words {
		ac.AddWord(word, []string{"test"}, i%3+1)
	}
	ac.BuildFailPointers()

	var wg sync.WaitGroup
	numGoroutines := 50
	operationsPerGoroutine := 100

	// 启动多个 goroutine 进行并发搜索
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				text := "这是一段包含敏感词1和辱骂词1的测试文本"
				options := &SearchOptions{
					Categories: []string{"test"},
					MinLevel:   (id+j)%3 + 1,
				}

				// 随机进行不同类型的搜索
				if j%2 == 0 {
					ac.Search(text)
				} else {
					ac.SearchWithOptions(text, options)
				}
			}
		}(i)
	}

	wg.Wait()

	// 验证自动机状态的一致性
	nodeCount := ac.GetNodeCount()
	if nodeCount <= 0 {
		t.Errorf("Node count should be positive, got %d", nodeCount)
	}
}

// BenchmarkACAutomatonBuildFailPointers 构建失败指针性能测试
func BenchmarkACAutomatonBuildFailPointers(b *testing.B) {
	ac := NewACAutomaton()

	// 添加大量敏感词
	words := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		words[i] = string(rune('a'+i%26)) + string(rune('a'+(i+1)%26)) + string(rune('a'+(i+2)%26))
		ac.AddWord(words[i], []string{"test"}, i%3+1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ac.BuildFailPointers()
	}
}
