package algorithm

import (
	"sync"
	"unicode/utf8"
)

// ACNode AC自动机节点
type ACNode struct {
	children map[rune]*ACNode // 子节点
	fail     *ACNode          // 失败指针
	output   []*Output        // 输出信息
	isEnd    bool             // 是否为结束节点
}

// Output 输出信息
type Output struct {
	Word       string   // 敏感词
	Categories []string // 分类
	Level      int      // 敏感级别
}

// ACAutomaton AC自动机
type ACAutomaton struct {
	root    *ACNode
	mu      sync.RWMutex
	version string
}

// NewACAutomaton 创建新的AC自动机
func NewACAutomaton() *ACAutomaton {
	return &ACAutomaton{
		root: &ACNode{
			children: make(map[rune]*ACNode),
			output:   make([]*Output, 0),
		},
	}
}

// AddWord 添加敏感词
func (ac *ACAutomaton) AddWord(word string, categories []string, level int) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if word == "" {
		return
	}

	node := ac.root
	for _, char := range word {
		if node.children[char] == nil {
			node.children[char] = &ACNode{
				children: make(map[rune]*ACNode),
				output:   make([]*Output, 0),
			}
		}
		node = node.children[char]
	}

	node.isEnd = true
	output := &Output{
		Word:       word,
		Categories: categories,
		Level:      level,
	}
	node.output = append(node.output, output)
}

// BuildFailPointers 构建失败指针
func (ac *ACAutomaton) BuildFailPointers() {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	queue := make([]*ACNode, 0)
	ac.root.fail = nil
	queue = append(queue, ac.root)

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		for char, child := range node.children {
			queue = append(queue, child)

			if node == ac.root {
				child.fail = ac.root
			} else {
				fail := node.fail
				for fail != nil {
					if fail.children[char] != nil {
						child.fail = fail.children[char]
						break
					}
					fail = fail.fail
				}
				if fail == nil {
					child.fail = ac.root
				}
			}

			// 合并输出
			if child.fail != nil {
				child.output = append(child.output, child.fail.output...)
			}
		}
	}
}

// Search 搜索敏感词
func (ac *ACAutomaton) Search(text string) []*Output {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	results := make([]*Output, 0)
	node := ac.root

	for _, char := range text {
		// 如果当前字符不匹配，沿着失败指针回溯
		for node.children[char] == nil && node != ac.root {
			node = node.fail
		}

		// 如果找到匹配的字符，移动到子节点
		if node.children[char] != nil {
			node = node.children[char]
		}

		// 收集输出
		if len(node.output) > 0 {
			results = append(results, node.output...)
		}
	}

	return results
}

// SearchWithOptions 带选项的搜索
func (ac *ACAutomaton) SearchWithOptions(text string, options *SearchOptions) []*Output {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	results := make([]*Output, 0)
	node := ac.root

	for _, char := range text {
		// 如果当前字符不匹配，沿着失败指针回溯
		for node.children[char] == nil && node != ac.root {
			node = node.fail
		}

		// 如果找到匹配的字符，移动到子节点
		if node.children[char] != nil {
			node = node.children[char]
		}

		// 收集输出
		if len(node.output) > 0 {
			for _, output := range node.output {
				if ac.matchesOptions(output, options) {
					results = append(results, output)
				}
			}
		}
	}

	return results
}

// matchesOptions 检查输出是否匹配选项
func (ac *ACAutomaton) matchesOptions(output *Output, options *SearchOptions) bool {
	// 检查敏感级别
	if output.Level < options.MinLevel {
		return false
	}

	// 检查分类
	if len(options.Categories) > 0 {
		found := false
		for _, category := range options.Categories {
			for _, outputCategory := range output.Categories {
				if category == outputCategory {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// Clear 清空自动机
func (ac *ACAutomaton) Clear() {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.root = &ACNode{
		children: make(map[rune]*ACNode),
		output:   make([]*Output, 0),
	}
	ac.version = ""
}

// GetVersion 获取版本
func (ac *ACAutomaton) GetVersion() string {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.version
}

// SetVersion 设置版本
func (ac *ACAutomaton) SetVersion(version string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.version = version
}

// GetNodeCount 获取节点数量
func (ac *ACAutomaton) GetNodeCount() int {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.countNodes(ac.root)
}

// countNodes 递归计算节点数量
func (ac *ACAutomaton) countNodes(node *ACNode) int {
	count := 0
	for _, child := range node.children {
		count += 1 + ac.countNodes(child)
	}
	return count
}

// SearchOptions 搜索选项
type SearchOptions struct {
	Categories []string // 要检查的分类
	MinLevel   int      // 最小敏感级别
}

// FuzzySearch 模糊搜索（支持拼音、简繁转换等）
func (ac *ACAutomaton) FuzzySearch(text string, options *SearchOptions) []*Output {
	// 这里可以实现更复杂的模糊匹配逻辑
	// 比如拼音匹配、简繁转换、同音字替换等
	return ac.SearchWithOptions(text, options)
}

// GetWordLength 获取字符长度（支持中文）
func GetWordLength(word string) int {
	return utf8.RuneCountInString(word)
}

// NormalizeText 标准化文本
func NormalizeText(text string) string {
	// 这里可以实现文本标准化逻辑
	// 比如去除特殊字符、统一大小写等
	return text
}
