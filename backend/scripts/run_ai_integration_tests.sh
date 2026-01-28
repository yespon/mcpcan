#!/bin/bash
# AI 功能集成测试脚本
# 用法: ./scripts/run_ai_integration_tests.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "=========================================="
echo "  AI 功能集成测试"
echo "=========================================="
echo ""

# 检查环境变量
check_env() {
    if [ -z "$DB_HOST" ]; then
        export DB_HOST="localhost"
        echo "ℹ️  使用默认 DB_HOST: localhost"
    fi
    if [ -z "$DB_PORT" ]; then
        export DB_PORT="31306"
        echo "ℹ️  使用默认 DB_PORT: 31306"
    fi
}

# 运行单元测试
run_unit_tests() {
    echo "📋 运行单元测试..."
    cd "$PROJECT_ROOT"
    go test -v ./pkg/common/... -count=1
    echo "✅ 单元测试通过"
    echo ""
}

# 运行认证测试
run_auth_tests() {
    echo "🔐 运行认证测试..."
    cd "$PROJECT_ROOT"
    go test -v ./internal/market/service/... -run "Test.*Auth" -count=1
    echo "✅ 认证测试通过"
    echo ""
}

# 运行集成测试
run_integration_tests() {
    echo "🧪 运行集成测试..."
    cd "$PROJECT_ROOT"
    go test -v ./internal/market/service/... -run "TestAiIntegration" -timeout 120s -count=1
    echo "✅ 集成测试通过"
    echo ""
}

# 生成覆盖率报告
generate_coverage() {
    echo "📊 生成覆盖率报告..."
    cd "$PROJECT_ROOT"
    
    # 确保 coverage 目录存在
    mkdir -p coverage
    
    # 运行测试并生成覆盖率
    go test -coverprofile=coverage/coverage.out \
        ./internal/market/biz/... \
        ./internal/market/service/... \
        ./pkg/common/... \
        -count=1 || true
    
    # 生成 HTML 报告
    go tool cover -html=coverage/coverage.out -o coverage/coverage.html 2>/dev/null || true
    
    # 显示覆盖率摘要
    if [ -f coverage/coverage.out ]; then
        echo ""
        echo "📈 覆盖率摘要:"
        go tool cover -func=coverage/coverage.out | tail -1
        echo ""
        echo "📁 覆盖率报告: coverage/coverage.html"
    fi
    echo ""
}

# 主函数
main() {
    check_env
    
    case "${1:-all}" in
        unit)
            run_unit_tests
            ;;
        auth)
            run_auth_tests
            ;;
        integration)
            run_integration_tests
            ;;
        coverage)
            generate_coverage
            ;;
        all)
            run_unit_tests
            run_auth_tests
            run_integration_tests
            generate_coverage
            ;;
        *)
            echo "用法: $0 {unit|auth|integration|coverage|all}"
            exit 1
            ;;
    esac
    
    echo "=========================================="
    echo "  ✅ 所有测试完成!"
    echo "=========================================="
}

main "$@"
