# mcpcan Backend Makefile
# AI 功能测试相关目标

.PHONY: all build test test-unit test-auth test-integration test-ai test-all coverage clean

# 默认目标
all: build

# 构建
build:
	@echo "🔨 构建项目..."
	go build -o bin/market ./cmd/market/...
	@echo "✅ 构建完成"

# 运行所有测试
test: test-all

# 单元测试
test-unit:
	@echo "📋 运行单元测试..."
	go test -v ./pkg/common/... -count=1
	@echo "✅ 单元测试完成"

# 认证测试
test-auth:
	@echo "🔐 运行认证测试..."
	go test -v ./internal/market/service/... -run "Test.*Auth" -count=1
	@echo "✅ 认证测试完成"

# AI 集成测试
test-integration:
	@echo "🧪 运行 AI 集成测试..."
	go test -v ./internal/market/service/... -run "TestAiIntegration" -timeout 120s -count=1
	@echo "✅ 集成测试完成"

# AI 功能完整测试 (包括认证和集成)
test-ai: test-auth test-integration
	@echo "✅ AI 功能测试全部完成"

# 运行所有测试
test-all: test-unit test-ai
	@echo "✅ 所有测试完成"

# 生成测试覆盖率
coverage:
	@echo "📊 生成覆盖率报告..."
	@mkdir -p coverage
	go test -coverprofile=coverage/coverage.out \
		./internal/market/biz/... \
		./internal/market/service/... \
		./pkg/common/... \
		-count=1 || true
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html 2>/dev/null || true
	@echo "📈 覆盖率摘要:"
	@go tool cover -func=coverage/coverage.out | tail -1
	@echo "📁 报告位置: coverage/coverage.html"

# 清理
clean:
	@echo "🧹 清理构建产物..."
	rm -rf bin/
	rm -rf coverage/
	@echo "✅ 清理完成"

# 帮助
help:
	@echo "可用目标:"
	@echo "  build           - 构建项目"
	@echo "  test-unit       - 运行单元测试"
	@echo "  test-auth       - 运行认证测试"
	@echo "  test-integration- 运行 AI 集成测试"
	@echo "  test-ai         - 运行 AI 功能测试 (认证+集成)"
	@echo "  test-all        - 运行所有测试"
	@echo "  coverage        - 生成覆盖率报告"
	@echo "  clean           - 清理构建产物"
