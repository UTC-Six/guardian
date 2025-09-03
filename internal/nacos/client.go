package nacos

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/sirupsen/logrus"

	"github.com/UTC-Six/guardian/internal/types"
)

// Client Nacos客户端
type Client struct {
	configClient config_client.IConfigClient
	config       *types.NacosConfig
	logger       *logrus.Logger
}

// NewClient 创建新的Nacos客户端
func NewClient(config *types.NacosConfig, logger *logrus.Logger) (*Client, error) {
	// 创建服务器配置
	serverConfigs := make([]constant.ServerConfig, 0, len(config.ServerConfigs))
	for _, serverConfig := range config.ServerConfigs {
		serverConfigs = append(serverConfigs, constant.ServerConfig{
			IpAddr: serverConfig.IpAddr,
			Port:   serverConfig.Port,
		})
	}

	// 创建客户端配置
	clientConfig := constant.ClientConfig{
		NamespaceId:         config.ClientConfig.NamespaceId,
		TimeoutMs:           config.ClientConfig.TimeoutMs,
		NotLoadCacheAtStart: config.ClientConfig.NotLoadCacheAtStart,
		LogDir:              config.ClientConfig.LogDir,
		CacheDir:            config.ClientConfig.CacheDir,
		LogLevel:            config.ClientConfig.LogLevel,
	}

	// 创建配置客户端
	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create nacos config client: %w", err)
	}

	return &Client{
		configClient: configClient,
		config:       config,
		logger:       logger,
	}, nil
}

// GetConfig 获取配置
func (c *Client) GetConfig(dataId, group string) (string, error) {
	content, err := c.configClient.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get config from nacos: %w", err)
	}

	return content, nil
}

// ListenConfig 监听配置变化
func (c *Client) ListenConfig(dataId, group string, callback func(string)) error {
	err := c.configClient.ListenConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
		OnChange: func(namespace, group, dataId, data string) {
			c.logger.Infof("Config changed: namespace=%s, group=%s, dataId=%s", namespace, group, dataId)
			callback(data)
		},
	})
	if err != nil {
		return fmt.Errorf("failed to listen config changes: %w", err)
	}

	return nil
}

// PublishConfig 发布配置
func (c *Client) PublishConfig(dataId, group, content string) error {
	success, err := c.configClient.PublishConfig(vo.ConfigParam{
		DataId:  dataId,
		Group:   group,
		Content: content,
	})
	if err != nil {
		return fmt.Errorf("failed to publish config: %w", err)
	}

	if !success {
		return fmt.Errorf("failed to publish config: operation not successful")
	}

	c.logger.Infof("Config published successfully: dataId=%s, group=%s", dataId, group)
	return nil
}

// GetWordDatabase 获取词库配置
func (c *Client) GetWordDatabase(dataId, group string) (*types.WordDatabase, error) {
	content, err := c.GetConfig(dataId, group)
	if err != nil {
		return nil, err
	}

	var wordDB types.WordDatabase
	if err := json.Unmarshal([]byte(content), &wordDB); err != nil {
		return nil, fmt.Errorf("failed to unmarshal word database: %w", err)
	}

	return &wordDB, nil
}

// PublishWordDatabase 发布词库配置
func (c *Client) PublishWordDatabase(dataId, group string, wordDB *types.WordDatabase) error {
	content, err := json.MarshalIndent(wordDB, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal word database: %w", err)
	}

	return c.PublishConfig(dataId, group, string(content))
}

// Close 关闭客户端
func (c *Client) Close() error {
	// Nacos客户端没有显式的关闭方法
	return nil
}

// HealthCheck 健康检查
func (c *Client) HealthCheck() error {
	// 尝试获取一个测试配置来检查连接
	_, err := c.GetConfig("health_check", "DEFAULT_GROUP")
	if err != nil {
		// 如果配置不存在，这是正常的，说明连接正常
		if err.Error() == "config not found" {
			return nil
		}
		return err
	}
	return nil
}

// GetConfigWithRetry 带重试的获取配置
func (c *Client) GetConfigWithRetry(dataId, group string, maxRetries int) (string, error) {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		content, err := c.GetConfig(dataId, group)
		if err == nil {
			return content, nil
		}
		lastErr = err
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	return "", fmt.Errorf("failed to get config after %d retries: %w", maxRetries, lastErr)
}
