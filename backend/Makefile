# mcpcan Backend Makefile
# AI Integration Test Targets

.PHONY: all build test test-unit test-auth test-integration test-ai test-all coverage clean

# Default target
all: build

# Build
build:
	@echo "🔨 Building project..."
	go build -o bin/market ./cmd/market/...
	@echo "✅ Build complete"

# Run all tests
test: test-all

# Unit tests
test-unit:
	@echo "📋 Running unit tests..."
	go test -v ./pkg/common/... -count=1
	@echo "✅ Unit tests complete"

# Authentication tests
test-auth:
	@echo "🔐 Running authentication tests..."
	go test -v ./internal/market/service/... -run "Test.*Auth" -count=1
	@echo "✅ Authentication tests complete"

# AI integration tests
test-integration:
	@echo "🧪 Running AI integration tests..."
	go test -v ./internal/market/service/... -run "TestAiIntegration" -timeout 120s -count=1
	@echo "✅ Integration tests complete"

# AI full tests (authentication + integration)
test-ai: test-auth test-integration
	@echo "✅ AI tests complete"

# Run all tests
test-all: test-unit test-ai
	@echo "✅ All tests complete"

# Generate test coverage
coverage:
	@echo "📊 Generating coverage report..."
	@mkdir -p coverage
	go test -coverprofile=coverage/coverage.out \
		./internal/market/biz/... \
		./internal/market/service/... \
		./pkg/common/... \
		-count=1 || true
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html 2>/dev/null || true
	@echo "📈 Coverage summary:"
	@go tool cover -func=coverage/coverage.out | tail -1
	@echo "📁 Report location: coverage/coverage.html"

# Clean
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	rm -rf coverage/
	@echo "✅ Clean complete"

# API code generation
gen-api:
	@echo "🔧 Generating API code..."
	cd api && buf generate
	@echo "✅ API code generation complete"

# OpenAPI 3 documentation generation
docs:
	@echo "📄 Generating OpenAPI 3 documentation..."
	./scripts/gen_openapi3.sh
	@echo "✅ Documentation generation complete"

# Help
help:
	@echo "Available targets:"
	@echo "  build           - Build the project"
	@echo "  test-unit       - Run unit tests"
	@echo "  test-auth       - Run authentication tests"
	@echo "  test-integration- Run AI integration tests"
	@echo "  test-ai         - Run AI tests (auth + integration)"
	@echo "  test-all        - Run all tests"
	@echo "  coverage        - Generate coverage report"
	@echo "  gen-api         - Generate API code (protobuf)"
	@echo "  docs            - Generate OpenAPI 3 documentation"
	@echo "  clean           - Clean build artifacts"
