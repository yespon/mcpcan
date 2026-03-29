#!/bin/bash

# ============================================================================
# MCPCan K8s 部署脚本 - 分步骤部署到 Kubernetes 集群
#
# 功能：
# 1. 检查 K8s 集群连接和依赖工具
# 2. 分步骤应用 YAML 配置（namespace → RBAC → configmap → deployment → ingress）
# 3. 等待服务就绪
# 4. 显示部署结果和访问方式
#
# 使用方式：
#   bash deploy.sh [options]
#
# 选项：
#   --dry-run              - 模拟部署（不实际应用）
#   --namespace            - K8s 命名空间（默认：mcpcan）
#   --harbor-registry      - Harbor 地址（用于镜像拉取）
#   --skip-preflight       - 跳过预检查
#   --wait-timeout         - 等待超时时间（默认：300秒）
#   --ingress-class        - Ingress Class（默认：nginx）
# ============================================================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 默认参数
DRY_RUN=0
NAMESPACE="mcpcan"
SKIP_PREFLIGHT=0
WAIT_TIMEOUT=300
INGRESS_CLASS="nginx"
HARBOR_REGISTRY="harbor.example.com"

# 脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

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

log_step() {
    echo -e "${BLUE}[步骤]${NC} $1"
}

# 显示帮助
show_help() {
    cat << EOF
用法: bash deploy.sh [options]

选项:
    --dry-run              - 模拟部署（显示如何应用，不实际执行）
    --namespace            - K8s 命名空间 (default: mcpcan)
    --harbor-registry      - Harbor 私有仓库地址 (default: harbor.example.com)
    --skip-preflight       - 跳过预检查（检查 kubectl、集群连接等）
    --wait-timeout         - 等待 Pod 就绪的超时时间，单位秒 (default: 300)
    --ingress-class        - Ingress Class 名称 (default: nginx)
    --help                 - 显示此帮助信息

示例:
    # 完整部署（包含预检查）
    bash deploy.sh --namespace mcpcan --harbor-registry harbor.company.com

    # 模拟部署（不执行）
    bash deploy.sh --dry-run

    # 快速部署（跳过预检查，适合已验证的集群）
    bash deploy.sh --skip-preflight --namespace mcpcan

EOF
}

# 解析命令行参数
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --dry-run)
                DRY_RUN=1
                shift
                ;;
            --namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            --harbor-registry)
                HARBOR_REGISTRY="$2"
                shift 2
                ;;
            --skip-preflight)
                SKIP_PREFLIGHT=1
                shift
                ;;
            --wait-timeout)
                WAIT_TIMEOUT="$2"
                shift 2
                ;;
            --ingress-class)
                INGRESS_CLASS="$2"
                shift 2
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

# 预检查：验证环境
preflight_check() {
    log_step "执行预检查..."

    # 检查 kubectl
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl 未安装或不在 PATH 中"
        exit 1
    fi
    log_info "✓ kubectl 已安装: $(kubectl version --client --short)"

    # 检查集群连接
    if ! kubectl cluster-info &> /dev/null; then
        log_error "无法连接到 K8s 集群"
        exit 1
    fi
    log_info "✓ K8s 集群连接正常"

    # 获取集群版本
    local k8s_version=$(kubectl version --short 2>/dev/null | grep Server | awk '{print $3}')
    log_info "  K8s 版本: $k8s_version"

    # 检查节点
    local node_count=$(kubectl get nodes -o jsonpath='{.items | length}')
    if [[ $node_count -lt 1 ]]; then
        log_error "集群中无可用节点"
        exit 1
    fi
    log_info "✓ 节点数量: $node_count"

    # 尝试获取当前上下文
    local current_context=$(kubectl config current-context)
    log_info "  当前上下文: $current_context"

    log_info "✓ 预检查完成"
}

# 应用 YAML
apply_yaml() {
    local yaml_file="$1"
    local step_name="$2"

    if [[ ! -f "$SCRIPT_DIR/$yaml_file" ]]; then
        log_error "文件不存在: $SCRIPT_DIR/$yaml_file"
        return 1
    fi

    log_step "$step_name"

    if [[ $DRY_RUN -eq 1 ]]; then
        log_info "[模拟] 将应用: $yaml_file"
        kubectl apply -f "$SCRIPT_DIR/$yaml_file" --dry-run=client -o yaml | head -20
        echo "    ... (更多内容)"
    else
        log_info "应用配置: $yaml_file"
        kubectl apply -f "$SCRIPT_DIR/$yaml_file"
        log_info "✓ 应用成功"
    fi
}

# 等待 Deployment 就绪
wait_deployment() {
    local deployment="$1"
    local timeout="$2"

    log_info "等待 $deployment 就绪（超时: ${timeout}秒）..."

    if kubectl -n "$NAMESPACE" wait --for=condition=available --timeout=${timeout}s deployment/$deployment 2>/dev/null; then
        log_info "✓ $deployment 已就绪"
        return 0
    else
        log_warn "⚠ $deployment 未在时间内就绪，继续..."
        return 1
    fi
}

# 检查并显示 Pod 状态
show_pod_status() {
    log_step "Pod 状态"
    kubectl -n "$NAMESPACE" get pods -o wide 2>/dev/null || true
}

# 显示服务信息
show_services() {
    log_step "服务信息"
    kubectl -n "$NAMESPACE" get svc -o wide 2>/dev/null || true
}

# 显示 Ingress 信息
show_ingress() {
    log_step "Ingress 信息"
    kubectl -n "$NAMESPACE" get ingress -o wide 2>/dev/null || true
}

# 显示访问信息
show_access_info() {
    log_step "访问信息"

    # 获取外部 IP 或 NodePort
    local service_type=$(kubectl -n "$NAMESPACE" get svc mcp-entry-svc -o jsonpath='{.spec.type}' 2>/dev/null || echo "ClusterIP")

    case "$service_type" in
        LoadBalancer)
            local external_ip=$(kubectl -n "$NAMESPACE" get svc mcp-entry-svc -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null)
            if [[ -z "$external_ip" ]]; then
                external_ip=$(kubectl -n "$NAMESPACE" get svc mcp-entry-svc -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null)
            fi
            
            if [[ -n "$external_ip" ]]; then
                log_info "  Web 前端: http://${external_ip}"
                log_info "  API 网关: http://${external_ip}/api"
                log_info "  Traefik Dashboard: http://${external_ip}:8080"
            else
                log_warn "  外部 IP 尚未分配，请稍候..."
            fi
            ;;
        NodePort)
            local node_port=$(kubectl -n "$NAMESPACE" get svc mcp-entry-svc -o jsonpath='{.spec.ports[0].nodePort}' 2>/dev/null)
            local node_ip=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="ExternalIP")].address}' 2>/dev/null)
            if [[ -z "$node_ip" ]]; then
                node_ip=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}' 2>/dev/null)
            fi
            
            if [[ -n "$node_ip" && -n "$node_port" ]]; then
                log_info "  Web 前端: http://${node_ip}:${node_port}"
                log_info "  API 网关: http://${node_ip}:${node_port}/api"
            else
                log_warn "  NodePort 尚未就绪"
            fi
            ;;
        *)
            log_warn "  Service 类型为 ClusterIP，仅在集群内可访问"
            log_info "  使用 port-forward 访问:"
            log_info "    kubectl port-forward -n $NAMESPACE svc/mcp-entry-svc 8080:80"
            log_info "    kubectl port-forward -n $NAMESPACE svc/mcp-entry-svc 8090:8080"
            ;;
    esac

    log_info ""
    log_info "默认登录凭证:"
    log_info "  用户名: admin"
    log_info "  密码: admin123"
}

# 主函数
main() {
    echo ""
    log_info "=========================================="
    log_info "MCPCan K8s 部署脚本"
    log_info "=========================================="
    echo ""

    parse_args "$@"

    # 显示配置
    log_info "部署配置:"
    log_info "  命名空间: $NAMESPACE"
    log_info "  Harbor 地址: $HARBOR_REGISTRY"
    log_info "  Ingress Class: $INGRESS_CLASS"
    log_info "  模拟模式: $([ $DRY_RUN -eq 1 ] && echo 'YES' || echo 'NO')"
    log_info "  等待超时: ${WAIT_TIMEOUT}秒"
    echo ""

    # 预检查
    if [[ $SKIP_PREFLIGHT -eq 0 ]]; then
        preflight_check
        echo ""
    else
        log_warn "跳过预检查"
    fi

    # 应用配置（按顺序）
    apply_yaml "0-namespace-rbac.yaml" "步骤1: 应用 Namespace 和 RBAC"
    echo ""

    apply_yaml "1-configmap-secret.yaml" "步骤2: 应用 ConfigMap 和 Secret"
    echo ""

    apply_yaml "2-services-deployment.yaml" "步骤3: 应用服务和 Deployment"
    echo ""

    apply_yaml "3-ingress.yaml" "步骤4: 应用 Ingress"
    echo ""

    if [[ $DRY_RUN -eq 0 ]]; then
        # 等待各服务就绪
        log_step "等待服务就绪..."
        wait_deployment "mysql" "$WAIT_TIMEOUT" || true
        sleep 5
        wait_deployment "redis" "$WAIT_TIMEOUT" || true
        sleep 10
        wait_deployment "mcp-market" "$WAIT_TIMEOUT" || true
        wait_deployment "mcp-authz" "$WAIT_TIMEOUT" || true
        wait_deployment "mcp-web" "$WAIT_TIMEOUT" || true
        wait_deployment "mcp-entry" "$WAIT_TIMEOUT" || true
        echo ""

        # 显示状态
        show_pod_status
        echo ""

        show_services
        echo ""

        show_ingress
        echo ""

        show_access_info

        log_info "=========================================="
        log_info "✓ 部署完成！"
        log_info "=========================================="
        echo ""
        log_info "常用命令:"
        log_info "  查看 Pod 日志:"
        log_info "    kubectl logs -n $NAMESPACE -f deployment/mcp-market"
        log_info "  进入 Pod 执行命令:"
        log_info "    kubectl exec -it -n $NAMESPACE deployment/mcp-market -- /bin/sh"
        log_info "  重启服务:"
        log_info "    kubectl rollout restart deployment/mcp-market -n $NAMESPACE"
        log_info "  查看事件:"
        log_info "    kubectl get events -n $NAMESPACE --sort-by='.lastTimestamp'"
        echo ""
    else
        log_info "=========================================="
        log_info "✓ 模拟部署完成（未实际应用）"
        log_info "=========================================="
    fi
}

# 执行主函数
main "$@"
