package algorithm

import (
	"sync"
	"unicode/utf8"
)

// ACNode AC自动机节点
// 设计思路：
// 1. 使用Trie树结构存储敏感词，每个节点代表一个字符
// 2. 失败指针(fail)实现KMP算法的回溯机制，避免重复匹配
// 3. 输出信息(output)存储匹配到的敏感词及其元数据
// 4. isEnd标记是否为敏感词的结束节点
type ACNode struct {
	children map[rune]*ACNode // 子节点：字符 -> 子节点映射
	fail     *ACNode          // 失败指针：KMP算法的回溯指针
	output   []*Output        // 输出信息：匹配到的敏感词列表
	isEnd    bool             // 是否为结束节点：敏感词结束标记
}

// Output 输出信息
// 设计思路：
// 1. 存储敏感词的完整信息和元数据
// 2. 支持多分类和敏感级别，便于精细化过滤
// 3. 提供丰富的上下文信息，支持复杂的过滤策略
type Output struct {
	Word       string   // 敏感词：匹配到的完整词汇
	Categories []string // 分类：敏感词所属的分类列表
	Level      int      // 敏感级别：1-10的敏感程度
}

// ACAutomaton AC自动机
// 设计思路：
// 1. 基于Aho-Corasick算法实现多模式字符串匹配
// 2. 使用读写锁保证并发安全，读操作可以并发执行
// 3. 版本控制支持词库的动态更新和回滚
// 4. 时间复杂度：构建O(Σ|Pi|)，搜索O(n+m+z)，其中n是文本长度，m是模式总长度，z是匹配数
type ACAutomaton struct {
	root    *ACNode      // 根节点：Trie树的根
	mu      sync.RWMutex // 读写锁：保证并发安全
	version string       // 版本号：词库版本控制
}

// NewACAutomaton 创建新的AC自动机
// 设计思路：
// 1. 初始化根节点，根节点不存储字符
// 2. 预分配输出切片，避免频繁的内存分配
// 3. 设置合理的初始容量，提升性能
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
