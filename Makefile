# Guardian 黄反校验SDK Makefile

.PHONY: build test clean run-example run-advanced-example docker-build docker-run

# 构建
build:
	go build -o bin/guardian ./cmd/guardian

# 测试
test:
	go test -v ./...

# 测试覆盖率
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# 清理
clean:
	rm -rf bin/
	rm -rf logs/
	rm -rf cache/
	rm -f coverage.out coverage.html

# 运行基本示例
run-example:
	go run examples/basic_usage.go

# 运行高级示例
run-advanced-example:
	go run examples/advanced_usage.go

# 安装依赖
deps:
	go mod download
	go mod tidy

# 格式化代码
fmt:
	go fmt ./...

# 代码检查
lint:
	golangci-lint run

# 构建Docker镜像
docker-build:
	docker build -t guardian-sdk .

# 运行Docker容器
docker-run:
	docker run -p 8848:8848 guardian-sdk

# 性能测试
benchmark:
	go test -bench=. -benchmem ./...

# 生成文档
docs:
	godoc -http=:6060

# 安装工具
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 检查代码质量
check: fmt lint test

# 发布准备
release: clean check build

# 帮助
help:
	@echo "Available targets:"
	@echo "  build              - Build the application"
	@echo "  test               - Run tests"
	@echo "  test-coverage      - Run tests with coverage"
	@echo "  clean              - Clean build artifacts"
	@echo "  run-example        - Run basic usage example"
	@echo "  run-advanced-example - Run advanced usage example"
	@echo "  deps               - Download and tidy dependencies"
	@echo "  fmt                - Format code"
	@echo "  lint               - Run linter"
	@echo "  docker-build       - Build Docker image"
	@echo "  docker-run         - Run Docker container"
	@echo "  benchmark          - Run benchmarks"
	@echo "  docs               - Generate documentation"
	@echo "  install-tools      - Install development tools"
	@echo "  check              - Run all checks (fmt, lint, test)"
	@echo "  release            - Prepare for release"
	@echo "  help               - Show this help"
