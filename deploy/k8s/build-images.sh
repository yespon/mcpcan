#!/bin/bash

# ============================================================================
# MCPCan 镜像构建和推送脚本
# 
# 功能：从本地源代码构建 Docker 镜像并推送到 Harbor 私有仓库
#
# 使用方式：
#   bash build-images.sh [options]
#
# 选项：
#   --harbor-registry   私有 Harbor 地址（默认：harbor.example.com）
#   --harbor-user       Harbor 用户名（默认：admin）
#   --harbor-pass       Harbor 密码（默认：Harbor12345）
#   --kube-context      K8s 上下文（默认：当前）
#   --project-path      项目路径（默认：../../）
#   --push              构建后推送到 Harbor
#   --skip-build        跳过构建，仅推送
#   --clean             构建前清理旧镜像
# ============================================================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 默认参数
HARBOR_REGISTRY="${HARBOR_REGISTRY:-harbor.example.com}"
HARBOR_USER="${HARBOR_USER:-admin}"
HARBOR_PASS="${HARBOR_PASS:-Harbor12345}"
HARBOR_PROJECT="${HARBOR_PROJECT:-mcpcan}"
PROJECT_PATH="../../"
PUSH_FLAG=0
SKIP_BUILD=0
CLEAN_FLAG=0

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 脚本帮助
show_help() {
    cat << EOF
用法: bash build-images.sh [options]

选项:
    --harbor-registry   - 私有 Harbor 地址 (default: harbor.example.com)
    --harbor-user       - Harbor 用户名 (default: admin)
    --harbor-pass       - Harbor 密码 (default: Harbor12345)
    --harbor-project    - Harbor 项目名 (default: mcpcan)
    --project-path      - MCPCan 项目路径 (default: ../../)
    --push              - 构建后推送到 Harbor
    --skip-build        - 跳过构建，仅推送
    --clean             - 构建前清理旧镜像
    --help              - 显示此帮助信息

示例:
    # 构建并推送到 Harbor
    bash build-images.sh --harbor-registry harbor.company.com --harbor-user admin --harbor-pass MyPass123 --push

    # 仅构建 (不推送)
    bash build-images.sh --project-path /path/to/mcpcan

    # 清理旧镜像后重新构建并推送
    bash build-images.sh --clean --push --harbor-registry harbor.company.com

EOF
}

# 解析命令行参数
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --harbor-registry)
                HARBOR_REGISTRY="$2"
                shift 2
                ;;
            --harbor-user)
                HARBOR_USER="$2"
                shift 2
                ;;
            --harbor-pass)
                HARBOR_PASS="$2"
                shift 2
                ;;
            --harbor-project)
                HARBOR_PROJECT="$2"
                shift 2
                ;;
            --project-path)
                PROJECT_PATH="$2"
                shift 2
                ;;
            --push)
                PUSH_FLAG=1
                shift
                ;;
            --skip-build)
                SKIP_BUILD=1
                shift
                ;;
            --clean)
                CLEAN_FLAG=1
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# 检查 Docker
check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装或不在 PATH 中"
        exit 1
    fi
    log_info "Docker 检查通过: $(docker --version)"
}

# 读取版本号
get_version() {
    if [[ -f "${PROJECT_PATH}/VERSION" ]]; then
        cat "${PROJECT_PATH}/VERSION"
    else
        echo "v2.1.1"
    fi
}

# 登录 Harbor
login_harbor() {
    if [[ $PUSH_FLAG -eq 1 ]]; then
        log_info "登录到 Harbor: $HARBOR_REGISTRY"
        echo "$HARBOR_PASS" | docker login -u "$HARBOR_USER" --password-stdin "$HARBOR_REGISTRY" || {
            log_error "Harbor 登录失败"
            exit 1
        }
    fi
}

# 清理旧镜像
clean_images() {
    if [[ $CLEAN_FLAG -eq 1 ]]; then
        log_info "清理旧镜像..."
        # 删除本地镜像标签
        docker rmi -f "$HARBOR_REGISTRY/$HARBOR_PROJECT/mcp-authz:$VERSION" 2>/dev/null || true
        docker rmi -f "$HARBOR_REGISTRY/$HARBOR_PROJECT/mcp-market:$VERSION" 2>/dev/null || true
        docker rmi -f "$HARBOR_REGISTRY/$HARBOR_PROJECT/mcp-web:$VERSION" 2>/dev/null || true
        log_info "清理完成"
    fi
}

# 构建镜像
build_authz() {
    if [[ $SKIP_BUILD -eq 0 ]]; then
        log_info "构建 mcp-authz 镜像..."
        docker build \
            -f "${PROJECT_PATH}/dockerfiles/Dockerfile.authz" \
            -t "$HARBOR_REGISTRY/$HARBOR_PROJECT/mcp-authz:$VERSION" \
            --build-arg CodeMode=OpenCode \
            "$PROJECT_PATH" || {
            log_error "mcp-authz 构建失败"
            exit 1
        }
        log_info "✓ mcp-authz 构建成功: $HARBOR_REGISTRY/$HARBOR_PROJECT/mcp-authz:$VERSION"
    fi

    if [[ $PUSH_FLAG -eq 1 ]]; then
        log_info "推送 mcp-authz 镜像到 Harbor..."
        docker push "$HARBOR_REGISTRY/$HARBOR_PROJECT/mcp-authz:$VERSION" || {
            log_error "mcp-authz 推送失败"
            exit 1
        }
        log_info "✓ mcp-authz 推送成功"
    fi
}

build_market() {
    if [[ $SKIP_BUILD -eq 0 ]]; then
        log_info "构建 mcp-market 镜像..."
        docker build \
            -f "${PROJECT_PATH}/dockerfiles/Dockerfile.market" \
            -t "$HARBOR_REGISTRY/$HARBOR_PROJECT/mcp-market:$VERSION" \
            --build-arg CodeMode=OpenCode \
            "$PROJECT_PATH" || {
            log_error "mcp-market 构建失败"
            exit 1
        }
        log_info "✓ mcp-market 构建成功: $HARBOR_REGISTRY/$HARBOR_PROJECT/mcp-market:$VERSION"
    fi

    if [[ $PUSH_FLAG -eq 1 ]]; then
        log_info "推送 mcp-market 镜像到 Harbor..."
        docker push "$HARBOR_REGISTRY/$HARBOR_PROJECT/mcp-market:$VERSION" || {
            log_error "mcp-market 推送失败"
            exit 1
        }
        log_info "✓ mcp-market 推送成功"
    fi
}

build_web() {
    if [[ $SKIP_BUILD -eq 0 ]]; then
        log_info "构建 mcp-web 镜像..."
        docker build \
            -f "${PROJECT_PATH}/dockerfiles/Dockerfile.frontend" \
            -t "$HARBOR_REGISTRY/$HARBOR_PROJECT/mcp-web:$VERSION" \
            --build-arg CodeMode=OpenCode \
            "$PROJECT_PATH" || {
            log_error "mcp-web 构建失败"
            exit 1
        }
        log_info "✓ mcp-web 构建成功: $HARBOR_REGISTRY/$HARBOR_PROJECT/mcp-web:$VERSION"
    fi

    if [[ $PUSH_FLAG -eq 1 ]]; then
        log_info "推送 mcp-web 镜像到 Harbor..."
        docker push "$HARBOR_REGISTRY/$HARBOR_PROJECT/mcp-web:$VERSION" || {
            log_error "mcp-web 推送失败"
            exit 1
        }
        log_info "✓ mcp-web 推送成功"
    fi
}

# 主函数
main() {
    log_info "=========================================="
    log_info "MCPCan K8s 镜像构建脚本"
    log_info "=========================================="

    parse_args "$@"

    check_docker
    VERSION=$(get_version)

    log_info "配置信息："
    log_info "  项目路径: $PROJECT_PATH"
    log_info "  版本号: $VERSION"
    log_info "  Harbor 地址: $HARBOR_REGISTRY"
    log_info "  Harbor 项目: $HARBOR_PROJECT"
    log_info "  推送到 Harbor: $([ $PUSH_FLAG -eq 1 ] && echo 'Yes' || echo 'No')"
    log_info "  清理旧镜像: $([ $CLEAN_FLAG -eq 1 ] && echo 'Yes' || echo 'No')"
    log_info ""

    login_harbor
    clean_images

    # 构建三个镜像
    build_authz
    build_market
    build_web

    log_info "=========================================="
    log_info "✓ 所有镜像构建成功！"
    log_info "=========================================="

    if [[ $PUSH_FLAG -eq 1 ]]; then
        log_info "✓ 所有镜像已推送到 Harbor"
        log_info ""
        log_info "下一步："
        log_info "  1. 在 K8s 集群中应用配置："
        log_info "     kubectl apply -f 0-namespace-rbac.yaml"
        log_info "     kubectl apply -f 1-configmap-secret.yaml"
        log_info "     kubectl apply -f 2-services-deployment.yaml"
        log_info "     kubectl apply -f 3-ingress.yaml"
        log_info "  2. 检查部署状态："
        log_info "     kubectl get pods -n mcpcan"
        log_info "  3. 查看服务："
        log_info "     kubectl get svc -n mcpcan"
    else
        log_warn "镜像已构建但未推送到 Harbor"
        log_warn "若要推送，请使用 --push 参数或手动推送："
        log_warn "  docker push $HARBOR_REGISTRY/$HARBOR_PROJECT/mcp-authz:$VERSION"
        log_warn "  docker push $HARBOR_REGISTRY/$HARBOR_PROJECT/mcp-market:$VERSION"
        log_warn "  docker push $HARBOR_REGISTRY/$HARBOR_PROJECT/mcp-web:$VERSION"
    fi
}

# 执行主函数
main "$@"
