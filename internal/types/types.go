package types

import "time"

// FilterResult 过滤结果
type FilterResult struct {
	Passed     bool              `json:"passed"`     // 是否通过
	Categories []string          `json:"categories"` // 匹配的敏感词分类
	Words      []string          `json:"words"`      // 匹配的敏感词
	Details    map[string]string `json:"details"`    // 详细信息
}

// SensitiveWord 敏感词结构
type SensitiveWord struct {
	Word       string   `json:"word"`       // 敏感词
	Categories []string `json:"categories"` // 分类
	Level      int      `json:"level"`      // 敏感级别 1-5
}

// Config 配置结构
type Config struct {
	NacosConfig NacosConfig `json:"nacos_config"`
	FilterConfig FilterConfig `json:"filter_config"`
}

// NacosConfig Nacos配置
type NacosConfig struct {
	ServerConfigs []ServerConfig `json:"server_configs"`
	ClientConfig  ClientConfig   `json:"client_config"`
}

// ServerConfig Nacos服务器配置
type ServerConfig struct {
	IpAddr string `json:"ip_addr"`
	Port   uint64 `json:"port"`
}

// ClientConfig Nacos客户端配置
type ClientConfig struct {
	NamespaceId         string `json:"namespace_id"`
	TimeoutMs           uint64 `json:"timeout_ms"`
	NotLoadCacheAtStart bool   `json:"not_load_cache_at_start"`
	LogDir              string `json:"log_dir"`
	CacheDir            string `json:"cache_dir"`
	LogLevel            string `json:"log_level"`
}

// FilterConfig 过滤器配置
type FilterConfig struct {
	DataId        string        `json:"data_id"`        // 配置ID
	Group         string        `json:"group"`          // 配置组
	ReloadPeriod  time.Duration `json:"reload_period"`  // 重载周期
	EnableCache   bool          `json:"enable_cache"`   // 是否启用缓存
	CacheSize     int           `json:"cache_size"`     // 缓存大小
	EnableWhitelist bool        `json:"enable_whitelist"` // 是否启用白名单
}

// WordDatabase 词库结构
type WordDatabase struct {
	Version      string                    `json:"version"`       // 版本号
	UpdateTime   time.Time                 `json:"update_time"`   // 更新时间
	Whitelist    []string                  `json:"whitelist"`     // 白名单
	Blacklist    []SensitiveWord           `json:"blacklist"`     // 黑名单
	Categories   map[string][]SensitiveWord `json:"categories"`   // 分类敏感词
	Replacements map[string]string         `json:"replacements"`  // 替换词
}

// FilterOptions 过滤选项
type FilterOptions struct {
	EnableWhitelist bool     `json:"enable_whitelist"` // 是否启用白名单
	Categories      []string `json:"categories"`       // 要检查的分类
	MinLevel        int      `json:"min_level"`        // 最小敏感级别
	ReplaceMode     bool     `json:"replace_mode"`     // 是否替换模式
}
