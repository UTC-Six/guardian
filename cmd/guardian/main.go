package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/UTC-Six/guardian/internal/types"
	"github.com/UTC-Six/guardian/pkg/guardian"
)

var (
	configFile = flag.String("config", "configs/config.yaml", "配置文件路径")
	port       = flag.String("port", "8080", "服务端口")
)

func main() {
	flag.Parse()

	// 加载配置
	config, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建Guardian实例
	g, err := guardian.NewGuardian(config)
	if err != nil {
		log.Fatalf("Failed to create Guardian: %v", err)
	}
	defer g.Close()

	// 设置HTTP路由
	http.HandleFunc("/health", healthHandler(g))
	http.HandleFunc("/check", checkHandler(g))
	http.HandleFunc("/check/batch", batchCheckHandler(g))
	http.HandleFunc("/stats", statsHandler(g))
	http.HandleFunc("/whitelist", whitelistHandler(g))

	// 启动HTTP服务器
	log.Printf("Starting server on port %s", *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

// loadConfig 加载配置文件
func loadConfig(filename string) (*types.Config, error) {
	// 这里简化处理，实际项目中应该使用yaml解析器
	// 为了演示，我们使用默认配置
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

	// 如果配置文件存在，尝试读取
	if _, err := os.Stat(filename); err == nil {
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		// 这里应该解析YAML，为了简化使用JSON
		if err := json.Unmarshal(data, config); err != nil {
			log.Printf("Warning: failed to parse config file, using default config: %v", err)
		}
	}

	return config, nil
}

// healthHandler 健康检查处理器
func healthHandler(g *guardian.Guardian) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := g.HealthCheck(); err != nil {
			http.Error(w, fmt.Sprintf("Health check failed: %v", err), http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	}
}

// checkHandler 单文本检查处理器
func checkHandler(g *guardian.Guardian) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Text    string               `json:"text"`
			Options *types.FilterOptions `json:"options,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		var result *types.FilterResult
		if req.Options != nil {
			result = g.CheckWithOptions(req.Text, req.Options)
		} else {
			result = g.Check(req.Text)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

// batchCheckHandler 批量检查处理器
func batchCheckHandler(g *guardian.Guardian) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Texts   []string             `json:"texts"`
			Options *types.FilterOptions `json:"options,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		var results []*types.FilterResult
		if req.Options != nil {
			results = g.BatchCheckWithOptions(req.Texts, req.Options)
		} else {
			results = g.BatchCheck(req.Texts)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}

// statsHandler 统计信息处理器
func statsHandler(g *guardian.Guardian) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := g.GetStats()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}
}

// whitelistHandler 白名单管理处理器
func whitelistHandler(g *guardian.Guardian) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			// 添加到白名单
			var req struct {
				Word string `json:"word"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
				return
			}
			g.AddToWhitelist(req.Word)
			w.WriteHeader(http.StatusOK)

		case http.MethodDelete:
			// 从白名单移除
			var req struct {
				Word string `json:"word"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
				return
			}
			g.RemoveFromWhitelist(req.Word)
			w.WriteHeader(http.StatusOK)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
