# MCPCan K8s 私有化部署 - 快速开始指南

## 📋 部署流程概览

MCPCan K8s 私有化部署脚本采用**分步骤、模块化**设计，支持完整的集群级部署和灵活的自定义扩展。

```
┌─────────────────────────────────────────────────────────────────┐
│ 部署流程                                                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│ 1. [前置准备] 集群和环境检查                                     │
│    └─ 检查 kubectl、K8s 集群、节点、网络                         │
│                                                                  │
│ 2. [镜像构建] 根据源代码构建 Docker 镜像                         │
│    └─ build-images.sh 构建 3 个镜像 → 推送到 Harbor             │
│                                                                  │
│ 3. [配置部署] 按顺序应用 Kubernetes YAML 配置                   │
│    3.1 Namespace + RBAC (0-namespace-rbac.yaml)                 │
│    3.2 ConfigMap + Secret (1-configmap-secret.yaml)            │
│    3.3 Services + Deployments (2-services-deployment.yaml)     │
│    3.4 Ingress (3-ingress.yaml)                                │
│                                                                  │
│ 4. [就绪检查] 等待所有服务就绪                                   │
│    └─ kubectl wait --for=condition=available deployment/xxx    │
│                                                                  │
│ 5. [验证访问] 显示服务状态和访问地址                             │
│    └─ Pod 状态、Service 信息、Ingress 地址                       │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🚀 快速开始（3 步部署）

### 前置条件
- ✅ 私有 K8s 集群（支持 v1.20+）
- ✅ 集群有可用存储路径映射（`/data/mcpcan`）
- ✅ 安装 kubectl 且能连接集群
- ✅ Docker 环境用于镜像构建
- ✅ 私有 Harbor 或其他镜像仓库（可选，默认用公网 77kymo）

### 步骤 1：修改配置文件

编辑 `values.yaml`，根据实际环境修改关键参数：

```bash
# 编辑参数文件
vim values.yaml

# 修改以下关键项：
HARBOR_REGISTRY=your-harbor.com        # ✏️ 改为你的 Harbor 地址
HARBOR_USER=admin                      # ✏️ Harbor 用户名
HARBOR_PASS=YourPassword123            # ✏️ Harbor 密码
HOST_DATA_PATH=/data/mcpcan            # ✏️ K8s 节点存储路径

# 如使用外部 MySQL/Redis，注释掉 USE_INTERNAL_* 并配置 EXTERNAL_*
```

### 步骤 2：构建镜像

```bash
# 基础构建（仅本地，不推送）
bash build-images.sh \
    --harbor-registry your-harbor.com \
    --harbor-user admin \
    --harbor-pass YourPassword123

# 构建 + 推送到 Harbor
bash build-images.sh \
    --harbor-registry your-harbor.com \
    --harbor-user admin \
    --harbor-pass YourPassword123 \
    --push
```

### 步骤 3：部署到 K8s

```bash
# 完整部署（包含预检查）
bash deploy.sh \
    --namespace mcpcan \
    --harbor-registry your-harbor.com

# 快速部署（跳过预检查，适合已验证环境）
bash deploy.sh --skip-preflight

# 模拟部署（查看即将执行的操作）
bash deploy.sh --dry-run
```

---

## 📁 文件说明

| 文件 | 用途 | 描述 |
|------|------|------|
| `0-namespace-rbac.yaml` | Namespace + RBAC | 创建 mcpcan 命名空间、ServiceAccount、ClusterRole、权限绑定 |
| `1-configmap-secret.yaml` | 配置存储 | 所有服务的配置文件（authz.yaml、market.yaml、Traefik config）和敏感信息（密码、密钥） |
| `2-services-deployment.yaml` | 服务部署 | 6 个 Service + 6 个 Deployment（可选 MySQL/Redis + 3 个业务服务 + Traefik入口） |
| `3-ingress.yaml` | 网络入口 | Nginx Ingress 配置（路由规则、跨域、速率限制等） |
| `build-images.sh` | 镜像构建 | 构建 3 个 Docker 镜像（authz、market、web），支持推送到 Harbor |
| `deploy.sh` | 部署脚本 | 主部署脚本，分步应用 YAML、等待就绪、显示访问信息 |
| `values.yaml` | 参数配置 | 所有可配置的参数（镜像地址、数据库、存储、副本数等） |
| `README.md` | 本文件 | 快速开始指南 |

---

## 🔧 常见场景和用法

### 场景 1：使用外部 MySQL 和 Redis

```yaml
# values.yaml 修改（注释掉内部，配置外部）
USE_INTERNAL_MYSQL=false
EXTERNAL_MYSQL_HOST=mysql.company.com
EXTERNAL_MYSQL_PORT=3306
EXTERNAL_MYSQL_DATABASE=mcp_dev

USE_INTERNAL_REDIS=false
EXTERNAL_REDIS_HOST=redis.company.com
EXTERNAL_REDIS_PORT=6379
```

然后从 `2-services-deployment.yaml` 删除 MySQL 和 Redis Deployment 部分。

### 场景 2：启用 HTTPS/TLS

```yaml
# values.yaml 修改
TLS_ENABLED=true
TLS_CRT_PATH=/path/to/cert.crt
TLS_KEY_PATH=/path/to/cert.key

# 创建 TLS Secret
kubectl create secret tls mcpcan-tls \
    --cert=/path/to/cert.crt \
    --key=/path/to/cert.key \
    -n mcpcan
```

然后在 `3-ingress.yaml` 取消注释 TLS 部分。

### 场景 3：多副本高可用部署

```yaml
# values.yaml 修改
AUTHZ_REPLICAS=3
MARKET_REPLICAS=3
WEB_REPLICAS=2
ENTRY_REPLICAS=2

# 修改 2-services-deployment.yaml，更改各 Deployment 的 replicas 字段
```

### 场景 4：集群中无外部 LoadBalancer，需要 NodePort 暴露

```yaml
# 修改 2-services-deployment.yaml 中 mcp-entry-svc
spec:
  type: NodePort  # 改为 NodePort
  ports:
  - port: 80
    targetPort: 80
    nodePort: 30080  # 工作节点:30080 可访问
```

然后通过 `工作节点IP:30080` 访问服务。

### 场景 5：制定计划内维护，快速更新镜像版本

```bash
# 1. 修改 VERSION 文件或直接修改镜像标签
echo "v2.1.2" > VERSION

# 2. 重新构建并推送镜像
bash build-images.sh --push

# 3. 更新 Deployment（Kubernetes 自动重建 Pod）
kubectl set image deployment/mcp-market \
    mcp-market=your-harbor.com/mcpcan/mcp-market:v2.1.2 \
    -n mcpcan
```

---

## 🔍 故障排查

### 检查 Pod 状态

```bash
# 查看所有 Pod（应该都是 Running）
kubectl get pods -n mcpcan

# 查看特定 Pod 详情（包括错误信息）
kubectl describe pod mcp-market-xxxxx -n mcpcan

# 查看 Pod 日志（排查启动错误）
kubectl logs -f deployment/mcp-market -n mcpcan

# 进入 Pod 调试
kubectl exec -it deployment/mcp-market -n mcpcan -- /bin/sh
```

### 检查服务连接

```bash
# 查看 Service（应该有 ClusterIP 和 Port）
kubectl get svc -n mcpcan

# 测试服务连通性（在任意 Pod 内执行）
kubectl run -n mcpcan debug-pod --image=busybox --rm -it -- \
    wget -O- http://mcp-authz-svc:8081/health

# 查看 Ingress 状态
kubectl get ingress -n mcpcan -o wide
```

### 数据库连接问题

```bash
# 检查 MySQL Pod（如使用内部 MySQL）
kubectl get pods -n mcpcan | grep mysql
kubectl logs -f pod/mysql-xxxxx -n mcpcan

# 进入 MySQL 容器，使用 mysql 客户端测试
kubectl exec -it pod/mysql-xxxxx -n mcpcan -- mysql \
    -u root -p<root_password> -e "SHOW DATABASES;"

# 检查 ConfigMap 中的数据库配置是否正确
kubectl get configmap mcpcan-config -n mcpcan -o yaml
```

### 镜像拉取失败

```bash
# 检查镜像拉取 Secret（如需私有仓库认证）
kubectl get secret -n mcpcan

# 查看 Pod 事件（通常会显示拉取失败原因）
kubectl describe pod mcp-market-xxxxx -n mcpcan | grep -A 5 "Events"

# 手动测试 Harbor 连接
docker login your-harbor.com
docker pull your-harbor.com/mcpcan/mcp-market:v2.1.1
```

### 存储权限问题

```bash
# 检查工作节点上的存储目录权限
ssh user@worker-node
ls -ld /data/mcpcan

# 确保目录可写（必要时调整权限）
sudo chmod 755 /data/mcpcan
sudo chown -R 1000:1000 /data/mcpcan  # 根据容器 UID 调整
```

---

## 📊 部署状态验证

部署完成后，检查以下指标确保服务正常：

```bash
# 1. 所有 Pod 应为 Running 或 Completed
kubectl get pods -n mcpcan

# 2. 所有 Service 应有 ClusterIP 和 Port
kubectl get svc -n mcpcan

# 3. Ingress 应显示外部 IP（如有 LoadBalancer）或规则
kubectl get ingress -n mcpcan -o wide

# 4. 检查事件日志，确保无异常
kubectl get events -n mcpcan --sort-by='.lastTimestamp'

# 5. 查询 API 健康状态（使用 port-forward）
kubectl port-forward -n mcpcan svc/mcp-authz-svc 8081:8081 &
curl http://localhost:8081/health
```

---

## 🌐 访问 MCPCan

部署成功后，根据 Service 类型访问：

### 方式 1：LoadBalancer（云平台）

```
Web 前端：http://<EXTERNAL_IP>
API 网关：http://<EXTERNAL_IP>/api
Traefik Dashboard：http://<EXTERNAL_IP>:8080
```

### 方式 2：NodePort（私有集群）

```
Web 前端：http://<任意Node_IP>:<NodePort>
API 网关：http://<任意Node_IP>:<NodePort>/api
Traefik Dashboard：http://<任意Node_IP>:<NodePort+1>（或自定义端口）
```

### 方式 3：ClusterIP + port-forward（开发调试）

```bash
# 转发端口
kubectl port-forward -n mcpcan svc/mcp-entry-svc 8080:80 &

# 然后访问
http://localhost:8080
http://localhost:8080/api
```

---

## 🔐 安全建议

1. **修改默认密码**
   - Admin 密码（修改 values.yaml 中的 ADMIN_PASSWORD）
   - MySQL root 密码（修改 MYSQL_ROOT_PASSWORD）
   - Redis 密码（修改 REDIS_PASSWORD）
   - APP_SECRET（修改 APP_SECRET）

2. **启用 TLS/HTTPS**
   - 参考上面的"场景 2：启用 HTTPS/TLS"

3. **配置网络策略（NetworkPolicy）**
   - 参考 `3-ingress.yaml` 中的 NetworkPolicy 示例

4. **镜像仓库安全**
   - 使用私有 Harbor 并启用认证
   - 在 Secret 中配置仓库凭证

5. **RBAC 权限最小化**
   - `0-namespace-rbac.yaml` 中的 ClusterRole 可根据实际需求调整

---

## 📞 获取帮助

如遇到问题，可按以下步骤排查：

1. **查看脚本日志输出** - deploy.sh 会打印详细信息
2. **检查 K8s 事件** - `kubectl get events -n mcpcan`
3. **查看 Pod 日志** - `kubectl logs -f deployment/<service>`
4. **查看配置** - `kubectl get configmap/<name> -o yaml`
5. **查看 K8s API 服务器日志**（若为自建集群）

---

## 📝 更新日志

- **v2.1.1** - 初始版本，支持私有化 K8s 部署
- 支持外部 MySQL/Redis
- 支持多副本高可用
- 自动化脚本和参数化配置

---

## 📄 许可证

MCPCan K8s 部署脚本遵循与主项目相同的许可证。
