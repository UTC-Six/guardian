package algorithm

import (
	"testing"
)

func TestACAutomaton(t *testing.T) {
	ac := NewACAutomaton()

	// 添加敏感词
	ac.AddWord("敏感词1", []string{"abuse"}, 3)
	ac.AddWord("敏感词2", []string{"politics"}, 4)
	ac.AddWord("辱骂词", []string{"abuse"}, 5)

	// 构建失败指针
	ac.BuildFailPointers()

	// 测试搜索
	tests := []struct {
		text     string
		expected int
	}{
		{"这是一段正常文本", 0},
		{"包含敏感词1的文本", 1},
		{"敏感词1和敏感词2都存在", 2},
		{"辱骂词不应该出现", 1},
		{"敏感词1敏感词2辱骂词", 3},
	}

	for _, test := range tests {
		results := ac.Search(test.text)
		if len(results) != test.expected {
			t.Errorf("Search(%s) = %d, expected %d", test.text, len(results), test.expected)
		}
	}
}

func TestACAutomatonWithOptions(t *testing.T) {
	ac := NewACAutomaton()

	// 添加敏感词
	ac.AddWord("敏感词1", []string{"abuse"}, 3)
	ac.AddWord("敏感词2", []string{"politics"}, 4)
	ac.AddWord("辱骂词", []string{"abuse"}, 5)

	// 构建失败指针
	ac.BuildFailPointers()

	// 测试带选项的搜索
	options := &SearchOptions{
		Categories: []string{"abuse"},
		MinLevel:   3,
	}

	results := ac.SearchWithOptions("敏感词1和敏感词2", options)
	if len(results) != 1 {
		t.Errorf("SearchWithOptions should return 1 result, got %d", len(results))
	}

	if results[0].Word != "敏感词1" {
		t.Errorf("Expected word '敏感词1', got '%s'", results[0].Word)
	}
}

func TestACAutomatonEmpty(t *testing.T) {
	ac := NewACAutomaton()
	ac.BuildFailPointers()

	results := ac.Search("任何文本")
	if len(results) != 0 {
		t.Errorf("Empty automaton should return no results")
	}
}

func TestACAutomatonClear(t *testing.T) {
	ac := NewACAutomaton()
	ac.AddWord("测试", []string{"test"}, 1)
	ac.BuildFailPointers()

	if ac.GetNodeCount() == 0 {
		t.Errorf("Automaton should have nodes after adding words")
	}

	ac.Clear()
	if ac.GetNodeCount() != 0 {
		t.Errorf("Automaton should be empty after clear")
	}
}

func BenchmarkACAutomatonSearch(b *testing.B) {
	ac := NewACAutomaton()

	// 添加大量敏感词
	words := []string{"敏感词1", "敏感词2", "敏感词3", "辱骂词1", "辱骂词2", "政治词1", "政治词2"}
	for i, word := range words {
		ac.AddWord(word, []string{"test"}, i%3+1)
	}

	ac.BuildFailPointers()

	text := "这是一段包含敏感词1和辱骂词1的测试文本"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ac.Search(text)
	}
}

func BenchmarkACAutomatonSearchWithOptions(b *testing.B) {
	ac := NewACAutomaton()

	// 添加大量敏感词
	words := []string{"敏感词1", "敏感词2", "敏感词3", "辱骂词1", "辱骂词2", "政治词1", "政治词2"}
	for i, word := range words {
		ac.AddWord(word, []string{"test"}, i%3+1)
	}

	ac.BuildFailPointers()

	text := "这是一段包含敏感词1和辱骂词1的测试文本"
	options := &SearchOptions{
		Categories: []string{"test"},
		MinLevel:   2,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ac.SearchWithOptions(text, options)
	}
}
