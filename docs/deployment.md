# Guardian 黄反校验SDK 部署指南

## 部署方式

### 1. 直接部署

#### 环境要求
- Go 1.21+
- Nacos Server 2.0+
- 内存: 最少512MB，推荐2GB+
- CPU: 最少2核，推荐4核+

#### 安装步骤
```bash
# 1. 克隆项目
git clone https://github.com/guardian/content-filter.git
cd content-filter

# 2. 安装依赖
go mod download

# 3. 构建
make build

# 4. 运行
./bin/guardian -config=configs/config.yaml -port=8080
```

### 2. Docker部署

#### 构建镜像
```bash
# 构建镜像
docker build -t guardian-sdk .

# 运行容器
docker run -d \
  --name guardian-sdk \
  -p 8080:8080 \
  -v $(pwd)/configs:/app/configs \
  -v $(pwd)/logs:/app/logs \
  -v $(pwd)/cache:/app/cache \
  guardian-sdk
```

#### Docker Compose部署
```bash
# 启动完整环境
docker-compose up -d

# 查看日志
docker-compose logs -f guardian

# 停止服务
docker-compose down
```

### 3. Kubernetes部署

#### 创建ConfigMap
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: guardian-config
data:
  config.yaml: |
    nacos_config:
      server_configs:
        - ip_addr: "nacos-server"
          port: 8848
      client_config:
        namespace_id: "public"
        timeout_ms: 5000
        log_dir: "/app/logs"
        cache_dir: "/app/cache"
        log_level: "info"
    filter_config:
      data_id: "sensitive_words"
      group: "DEFAULT_GROUP"
      reload_period: "5m"
      enable_cache: true
      cache_size: 10000
      enable_whitelist: true
```

#### 创建Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: guardian-sdk
spec:
  replicas: 3
  selector:
    matchLabels:
      app: guardian-sdk
  template:
    metadata:
      labels:
        app: guardian-sdk
    spec:
      containers:
      - name: guardian
        image: guardian-sdk:latest
        ports:
        - containerPort: 8080
        volumeMounts:
        - name: config
          mountPath: /app/configs
        - name: logs
          mountPath: /app/logs
        - name: cache
          mountPath: /app/cache
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "2Gi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: guardian-config
      - name: logs
        emptyDir: {}
      - name: cache
        emptyDir: {}
```

#### 创建Service
```yaml
apiVersion: v1
kind: Service
metadata:
  name: guardian-service
spec:
  selector:
    app: guardian-sdk
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP
```

## Nacos配置

### 1. 安装Nacos

#### Docker安装
```bash
docker run -d \
  --name nacos \
  -p 8848:8848 \
  -p 9848:9848 \
  -e MODE=standalone \
  nacos/nacos-server:v2.2.3
```

#### 集群安装
```bash
# 修改配置文件
vim conf/cluster.conf

# 添加节点
192.168.1.10:8848
192.168.1.11:8848
192.168.1.12:8848
```

### 2. 配置敏感词库

#### 创建配置
```bash
# 使用Nacos API创建配置
curl -X POST "http://localhost:8848/nacos/v1/cs/configs" \
  -d "dataId=sensitive_words" \
  -d "group=DEFAULT_GROUP" \
  -d "content=$(cat configs/sensitive_words.json)"
```

#### 配置内容示例
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

## 监控配置

### 1. Prometheus监控

#### 添加监控端点
```go
// 在main.go中添加
import "github.com/prometheus/client_golang/prometheus/promhttp"

http.Handle("/metrics", promhttp.Handler())
```

#### Prometheus配置
```yaml
scrape_configs:
  - job_name: 'guardian-sdk'
    static_configs:
      - targets: ['guardian-sdk:8080']
    metrics_path: '/metrics'
    scrape_interval: 15s
```

### 2. Grafana仪表板

#### 关键指标
- 请求延迟 (P50, P90, P99)
- QPS
- 错误率
- 缓存命中率
- 内存使用量
- CPU使用率

### 3. 日志配置

#### 结构化日志
```go
// 配置日志格式
logger := logrus.New()
logger.SetFormatter(&logrus.JSONFormatter{
    TimestampFormat: "2006-01-02 15:04:05",
})
logger.SetLevel(logrus.InfoLevel)
```

#### 日志轮转
```bash
# 使用logrotate
/var/log/guardian/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 guardian guardian
    postrotate
        kill -USR1 $(cat /var/run/guardian.pid)
    endscript
}
```

## 高可用部署

### 1. 负载均衡

#### Nginx配置
```nginx
upstream guardian_backend {
    server guardian-1:8080;
    server guardian-2:8080;
    server guardian-3:8080;
}

server {
    listen 80;
    server_name guardian.example.com;
    
    location / {
        proxy_pass http://guardian_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

#### HAProxy配置
```
backend guardian_backend
    balance roundrobin
    server guardian-1 guardian-1:8080 check
    server guardian-2 guardian-2:8080 check
    server guardian-3 guardian-3:8080 check
```

### 2. 故障转移

#### 健康检查
```bash
# 健康检查脚本
#!/bin/bash
curl -f http://localhost:8080/health || exit 1
```

#### 自动重启
```bash
# systemd服务配置
[Unit]
Description=Guardian SDK
After=network.target

[Service]
Type=simple
User=guardian
WorkingDirectory=/opt/guardian
ExecStart=/opt/guardian/bin/guardian
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### 3. 数据备份

#### 配置备份
```bash
# 定期备份Nacos配置
#!/bin/bash
curl -s "http://nacos:8848/nacos/v1/cs/configs?dataId=sensitive_words&group=DEFAULT_GROUP" \
  > backup/sensitive_words_$(date +%Y%m%d_%H%M%S).json
```

#### 缓存备份
```bash
# 备份缓存数据
tar -czf cache_backup_$(date +%Y%m%d_%H%M%S).tar.gz cache/
```

## 性能调优

### 1. 系统调优

#### 内核参数
```bash
# 调整文件描述符限制
echo "* soft nofile 65536" >> /etc/security/limits.conf
echo "* hard nofile 65536" >> /etc/security/limits.conf

# 调整网络参数
echo "net.core.somaxconn = 65536" >> /etc/sysctl.conf
echo "net.ipv4.tcp_max_syn_backlog = 65536" >> /etc/sysctl.conf
sysctl -p
```

#### Go运行时参数
```bash
# 调整GC参数
export GOGC=100
export GOMAXPROCS=8

# 启用pprof
export GODEBUG=pprof=1
```

### 2. 应用调优

#### 缓存配置
```yaml
filter_config:
  enable_cache: true
  cache_size: 50000  # 根据内存调整
  cache_ttl: "10m"   # 根据业务调整
```

#### 并发配置
```yaml
server_config:
  max_concurrent: 10000
  read_timeout: "30s"
  write_timeout: "30s"
```

### 3. 监控调优

#### 指标收集
```yaml
metrics:
  enabled: true
  interval: "10s"
  retention: "7d"
```

#### 告警配置
```yaml
alerts:
  - name: "high_latency"
    condition: "p99_latency > 1ms"
    duration: "5m"
  - name: "high_error_rate"
    condition: "error_rate > 0.1%"
    duration: "2m"
```

## 故障排查

### 1. 常见问题

#### 启动失败
```bash
# 检查配置
./bin/guardian -config=configs/config.yaml -check-config

# 检查端口
netstat -tlnp | grep 8080

# 检查日志
tail -f logs/guardian.log
```

#### 性能问题
```bash
# 检查CPU使用
top -p $(pgrep guardian)

# 检查内存使用
ps aux | grep guardian

# 检查网络连接
ss -tuln | grep 8080
```

#### 配置问题
```bash
# 检查Nacos连接
curl http://nacos:8848/nacos/v1/ns/instance/list?serviceName=nacos

# 检查配置内容
curl "http://nacos:8848/nacos/v1/cs/configs?dataId=sensitive_words&group=DEFAULT_GROUP"
```

### 2. 调试工具

#### pprof分析
```bash
# 启动pprof
go tool pprof http://localhost:8080/debug/pprof/profile

# 内存分析
go tool pprof http://localhost:8080/debug/pprof/heap

# 协程分析
go tool pprof http://localhost:8080/debug/pprof/goroutine
```

#### 性能分析
```bash
# 运行基准测试
go test -bench=. -benchmem ./...

# 运行压力测试
go test -bench=. -benchtime=30s ./...
```

## 安全配置

### 1. 网络安全

#### 防火墙配置
```bash
# 只允许必要端口
iptables -A INPUT -p tcp --dport 8080 -j ACCEPT
iptables -A INPUT -p tcp --dport 8848 -j ACCEPT
```

#### TLS配置
```go
// 启用HTTPS
server := &http.Server{
    Addr:    ":8443",
    Handler: mux,
    TLSConfig: &tls.Config{
        MinVersion: tls.VersionTLS12,
    },
}
server.ListenAndServeTLS("cert.pem", "key.pem")
```

### 2. 访问控制

#### API认证
```go
// 添加API Key认证
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        apiKey := r.Header.Get("X-API-Key")
        if apiKey != "your-secret-key" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

#### 限流配置
```go
// 添加限流
import "golang.org/x/time/rate"

limiter := rate.NewLimiter(1000, 100) // 1000 QPS, 突发100
```

## 总结

Guardian SDK提供了多种部署方式，支持从简单的单机部署到复杂的集群部署。通过合理的配置和监控，可以实现高可用、高性能的黄反校验服务。

关键部署要点：
1. **配置管理**: 使用Nacos进行配置管理
2. **监控告警**: 完善的监控和告警机制
3. **高可用**: 负载均衡和故障转移
4. **性能优化**: 系统级和应用级调优
5. **安全防护**: 网络安全和访问控制
