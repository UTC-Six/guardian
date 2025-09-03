# Guardian 黄反校验SDK

一个基于Go语言开发的高性能黄反校验SDK，支持Nacos配置中心、热加载、分类管理等特性。

## 特性

- 🚀 **高性能**: 基于AC自动机算法，支持大规模敏感词匹配
- 🔄 **热加载**: 支持从Nacos配置中心动态更新敏感词库
- 📊 **分类管理**: 支持敏感词分类（辱骂、政治、暴力、成人等）
- ⚡ **缓存优化**: 内置LRU缓存，提升重复查询性能
- 🛡️ **白名单机制**: 支持白名单，避免误判
- 🔧 **易于集成**: 提供简洁的API接口
- 📈 **监控统计**: 内置统计信息，便于监控和调优

## 架构设计

### 核心组件

1. **AC自动机算法**: 高效的敏感词匹配算法
2. **Nacos集成**: 配置中心集成，支持动态配置更新
3. **缓存层**: LRU缓存，提升性能
4. **分类管理**: 支持多维度敏感词分类
5. **热加载**: 配置变更自动生效

### 技术栈

- Go 1.21+
- Nacos SDK
- Logrus (日志)
- AC自动机算法

## 快速开始

### 安装

```bash
go get github.com/UTC-Six/guardian
```

### 基本使用

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/UTC-Six/guardian/pkg/guardian"
    "github.com/UTC-Six/guardian/internal/types"
)

func main() {
    // 创建配置
    config := &types.Config{
        NacosConfig: types.NacosConfig{
            ServerConfigs: []types.ServerConfig{
                {IpAddr: "127.0.0.1", Port: 8848},
            },
            ClientConfig: types.ClientConfig{
                NamespaceId: "public",
                TimeoutMs:   5000,
                LogDir:      "./logs",
                CacheDir:    "./cache",
                LogLevel:    "info",
            },
        },
        FilterConfig: types.FilterConfig{
            DataId:         "sensitive_words",
            Group:          "DEFAULT_GROUP",
            ReloadPeriod:   5 * time.Minute,
            EnableCache:    true,
            CacheSize:      10000,
            EnableWhitelist: true,
        },
    }

    // 创建Guardian实例
    g, err := guardian.NewGuardian(config)
    if err != nil {
        log.Fatal(err)
    }
    defer g.Close()

    // 检查文本
    result := g.Check("这是一段包含敏感词的文本")
    fmt.Printf("通过: %v\n", result.Passed)
    if !result.Passed {
        fmt.Printf("匹配词: %v\n", result.Words)
        fmt.Printf("分类: %v\n", result.Categories)
    }
}
```

### 高级使用

```go
// 分类检查
result := g.CheckCategory("辱骂词", []string{"abuse"})

// 级别检查
result := g.CheckLevel("敏感词", 3)

// 自定义选项
options := &types.FilterOptions{
    EnableWhitelist: true,
    Categories:      []string{"abuse", "politics"},
    MinLevel:        3,
    ReplaceMode:     false,
}
result := g.CheckWithOptions("文本", options)

// 批量检查
texts := []string{"文本1", "文本2", "文本3"}
results := g.BatchCheck(texts)
```

## 配置说明

### Nacos配置

```yaml
nacos_config:
  server_configs:
    - ip_addr: "127.0.0.1"
      port: 8848
  client_config:
    namespace_id: "public"
    timeout_ms: 5000
    not_load_cache_at_start: false
    log_dir: "./logs"
    cache_dir: "./cache"
    log_level: "info"
```

### 过滤器配置

```yaml
filter_config:
  data_id: "sensitive_words"
  group: "DEFAULT_GROUP"
  reload_period: "5m"
  enable_cache: true
  cache_size: 10000
  enable_whitelist: true
```

### 敏感词库配置

```json
{
  "version": "1.0.0",
  "update_time": "2024-01-01T00:00:00Z",
  "whitelist": [
    "正常词汇1",
    "正常词汇2"
  ],
  "blacklist": [
    {
      "word": "敏感词1",
      "categories": ["abuse", "politics"],
      "level": 3
    }
  ],
  "categories": {
    "abuse": [
      {
        "word": "辱骂词1",
        "categories": ["abuse"],
        "level": 4
      }
    ],
    "politics": [
      {
        "word": "政治敏感词1",
        "categories": ["politics"],
        "level": 5
      }
    ]
  },
  "replacements": {
    "敏感词1": "***"
  }
}
```

## API接口

### 核心方法

- `Check(text string) *FilterResult`: 基本文本检查
- `CheckWithOptions(text string, options *FilterOptions) *FilterResult`: 带选项检查
- `CheckCategory(text string, categories []string) *FilterResult`: 分类检查
- `CheckLevel(text string, minLevel int) *FilterResult`: 级别检查
- `BatchCheck(texts []string) []*FilterResult`: 批量检查
- `IsSafe(text string) bool`: 简单安全检查

### 管理方法

- `GetStats() map[string]interface{}`: 获取统计信息
- `HealthCheck() error`: 健康检查
- `AddToWhitelist(word string)`: 添加白名单
- `RemoveFromWhitelist(word string)`: 移除白名单
- `UpdateWordDatabase(wordDB *WordDatabase) error`: 更新词库

## 性能优化

### AC自动机算法

- 时间复杂度: O(n + m + z)，其中n是文本长度，m是模式总长度，z是匹配数
- 空间复杂度: O(m)
- 支持多模式匹配，一次扫描找到所有敏感词

### 缓存策略

- LRU缓存，自动淘汰最少使用的项
- 可配置缓存大小和TTL
- 支持缓存统计和监控

### 并发安全

- 读写锁保护，支持高并发访问
- 原子操作更新配置
- 无锁设计的关键路径

## 部署指南

### Docker部署

```bash
# 构建镜像
docker build -t guardian-sdk .

# 运行容器
docker run -p 8080:8080 guardian-sdk
```

### 直接部署

```bash
# 构建
make build

# 运行
./bin/guardian -config=configs/config.yaml -port=8080
```

### HTTP服务

启动后提供以下HTTP接口：

- `POST /check`: 单文本检查
- `POST /check/batch`: 批量检查
- `GET /stats`: 统计信息
- `GET /health`: 健康检查
- `POST /whitelist`: 添加白名单
- `DELETE /whitelist`: 移除白名单

## 监控和运维

### 统计信息

```go
stats := g.GetStats()
fmt.Printf("版本: %v\n", stats["version"])
fmt.Printf("节点数: %v\n", stats["node_count"])
fmt.Printf("白名单大小: %v\n", stats["whitelist_size"])
fmt.Printf("缓存统计: %v\n", stats["cache_stats"])
```

### 健康检查

```go
if err := g.HealthCheck(); err != nil {
    log.Printf("健康检查失败: %v", err)
}
```

### 日志配置

支持结构化日志，可配置日志级别和输出格式。

## 测试

```bash
# 运行测试
make test

# 运行性能测试
make benchmark

# 测试覆盖率
make test-coverage
```

## 贡献指南

1. Fork 项目
2. 创建特性分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 许可证

MIT License

## 联系方式

如有问题或建议，请提交 Issue 或联系维护者。