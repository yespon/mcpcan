# MCPCan K8s 私有化部署 - 部署说明

## 📋 部署流程概览

MCPCan K8s 私有化部署脚本采用**分步骤、模块化**设计，依赖**外部 MySQL、Redis** 和**火山云 TOS 对象存储**。

```
┌─────────────────────────────────────────────────────────────────┐
│ 完整部署流程                                                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│ Phase 0：基础设施准备（部署前，手动操作）                        │
│    ├─ A. 准备外部 MySQL（建库、建用户、初始化表结构）            │
│    ├─ B. 准备外部 Redis（启用密码认证）                          │
│    ├─ C. 火山云 TOS 创建 Bucket + IAM 访问密钥                  │
│    ├─ D. Harbor 镜像仓库准备 + imagePullSecret                  │
│    └─ E. 生成并写入 kubeconfig Secret（market 服务必需）        │
│                                                                  │
│ Phase 1：配置文件修改                                            │
│    ├─ 修改 1-configmap-secret.yaml（数据库地址、密码、域名）     │
│    └─ 修改 4-storage-s3.yaml（TOS 凭证、Bucket 名称、端点）     │
│                                                                  │
│ Phase 2：镜像构建 & 推送                                        │
│    └─ build-images.sh 构建 3 个镜像 → 推送到 Harbor             │
│                                                                  │
│ Phase 3：K8s 资源部署（顺序执行）                               │
│    3.1 Namespace + RBAC       (0-namespace-rbac.yaml)           │
│    3.2 ConfigMap + Secret     (1-configmap-secret.yaml)         │
│    3.3 TOS CSI 存储           (4-storage-s3.yaml)               │
│        └─ 等待 csi-s3 DaemonSet 就绪                           │
│    3.4 Services + Deployments (2-services-deployment.yaml)     │
│    3.5 Ingress               (3-ingress.yaml)                  │
│                                                                  │
│ Phase 4：验证                                                    │
│    └─ Pod 就绪检查 / 健康探针 / 访问地址确认                    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🚀 快速开始

### 前置条件
- 私有 K8s 集群，版本建议 v1.20+
- 已安装 kubectl，且当前上下文可访问目标集群
- 已安装 Docker，用于本地构建镜像
- 已准备外部 MySQL 和 Redis
- 已准备火山云 TOS Bucket 与 AccessKey
- 已准备 Harbor 或其他可访问的私有镜像仓库
- K8s 节点可以访问以下外部依赖：MySQL、Redis、TOS、Harbor

## Phase 0：部署前准备

### 0.1 准备外部 MySQL

MCPCan 当前部署配置假定 **MySQL 由外部提供**，不在 K8s 内创建 MySQL Pod。

建议准备：
- 数据库：mcp_dev
- 业务用户：mcp_user
- 权限：对 mcp_dev 拥有完整读写权限
- 字符集：utf8mb4

示例：

```sql
CREATE DATABASE IF NOT EXISTS mcp_dev DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'mcp_user'@'%' IDENTIFIED BY 'YourStrongPassword';
GRANT ALL PRIVILEGES ON mcp_dev.* TO 'mcp_user'@'%';
FLUSH PRIVILEGES;
```

需要确认：
- K8s 工作节点到 MySQL 地址端口网络可达
- MySQL 防火墙 / 安全组允许集群出口访问
- 账号密码与 [deploy/k8s/1-configmap-secret.yaml](deploy/k8s/1-configmap-secret.yaml) 中配置一致

### 0.2 准备外部 Redis

Redis 也由外部提供，建议至少开启密码认证。

需要确认：
- Redis 地址和端口可由 K8s 节点访问
- 已设置访问密码
- 如启用了 ACL，请确保默认用户或指定用户有读写权限

### 0.3 准备火山云 TOS

在火山云控制台完成以下操作：

1. 创建 Bucket，例如 `mcpcan-data`
2. 记录 Bucket 所在地域，例如 `cn-beijing`
3. 创建 IAM 访问密钥（AccessKey ID / SecretAccessKey）
4. 为该密钥授予目标 Bucket 的读写权限
5. 确认 K8s 节点可访问对应地域端点

常见端点：

```text
北京:   https://tos-s3-cn-beijing.volces.com
上海:   https://tos-s3-cn-shanghai.volces.com
广州:   https://tos-s3-cn-guangzhou.volces.com
桂林:   https://tos-s3-cn-guilin.volces.com
```

### 0.4 准备 Harbor 镜像仓库

需要至少准备 3 个业务镜像仓库：
- mcp-authz
- mcp-market
- mcp-web

如果私有集群不能直接访问公网，还需要提前同步以下基础镜像到 Harbor：
- traefik:v3.6.9
- ghcr.io/yandex-cloud/k8s-csi-s3/csi-s3:v0.35.5
- registry.k8s.io/sig-storage/csi-provisioner:v4.0.0
- registry.k8s.io/sig-storage/csi-attacher:v4.4.0
- registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.9.0

### 0.5 创建 Harbor 拉取凭证

如果镜像仓库需要认证，先创建 imagePullSecret：

```bash
kubectl create namespace mcpcan --dry-run=client -o yaml | kubectl apply -f -

kubectl create secret docker-registry harbor-registry-secret \
    --namespace mcpcan \
    --docker-server=your-harbor.com \
    --docker-username=admin \
    --docker-password='YourPassword123' \
    --docker-email=devops@example.com
```

注意：当前 [deploy/k8s/2-services-deployment.yaml](deploy/k8s/2-services-deployment.yaml) 尚未自动引用 `imagePullSecrets`。如果 Harbor 需要认证，需要在各 Deployment 的 `spec.template.spec` 下补充：

```yaml
imagePullSecrets:
    - name: harbor-registry-secret
```

### 0.6 生成 kubeconfig Secret

market 服务需要访问 Kubernetes API 来管理运行环境，因此必须提供 kubeconfig。

创建方式示例：

```bash
kubectl create secret generic mcpcan-kubeconfig \
    --namespace mcpcan \
    --from-file=kubeconfig.yaml=/path/to/your/kubeconfig
```

建议：
- 使用最小权限专用 kubeconfig
- 对应权限需和 [deploy/k8s/0-namespace-rbac.yaml](deploy/k8s/0-namespace-rbac.yaml) 中设计一致
- 不要直接使用集群管理员 kubeconfig 投产

## Phase 1：修改配置文件

### 1.1 修改应用配置

重点修改 [deploy/k8s/1-configmap-secret.yaml](deploy/k8s/1-configmap-secret.yaml)：

- `market.yaml` 中的 `database.mysql.host`
- `market.yaml` 中的 `database.redis.host`
- `authz.yaml` 中的 `database.mysql.host`
- `authz.yaml` 中的 `database.redis.host`
- `secret`
- `init.admin_username` 和 `init.admin_password`
- `env.js` 中的前端访问地址

当前文件中的默认值仍是：

```yaml
mysql:
    host: mysql-svc
redis:
    host: redis-svc
```

部署外部数据库时，必须改成真实地址，例如：

```yaml
mysql:
    host: mysql.company.local
    port: 3306
redis:
    host: redis.company.local
    port: 6379
```

### 1.2 修改 TOS 存储配置

编辑 [deploy/k8s/4-storage-s3.yaml](deploy/k8s/4-storage-s3.yaml)，至少替换以下值：

- `accessKeyID`
- `secretAccessKey`
- `endpoint`
- `region`
- `bucket`

如果基础镜像已经同步到 Harbor，也应一并把其中的 `image` 字段改成 Harbor 地址。

### 1.3 检查 values.yaml

[deploy/k8s/values.yaml](deploy/k8s/values.yaml) 现在已经对齐当前部署架构，可作为部署前参数核对清单使用。

但要注意：当前 deploy/k8s 目录下仍是**静态 YAML**，修改 values.yaml 不会自动渲染到部署清单，仍需手动同步到对应 YAML 文件。

另外，当前静态清单还存在两个固定约束：
- namespace 固定为 mcpcan
- ingressClass 固定为 nginx

部署时应以以下文件为准：
- [deploy/k8s/1-configmap-secret.yaml](deploy/k8s/1-configmap-secret.yaml)
- [deploy/k8s/2-services-deployment.yaml](deploy/k8s/2-services-deployment.yaml)
- [deploy/k8s/3-ingress.yaml](deploy/k8s/3-ingress.yaml)
- [deploy/k8s/4-storage-s3.yaml](deploy/k8s/4-storage-s3.yaml)

## Phase 2：构建并推送镜像

```bash
# 仅构建
bash build-images.sh \
    --harbor-registry your-harbor.com \
    --harbor-user admin \
    --harbor-pass 'YourPassword123'

# 构建并推送
bash build-images.sh \
    --harbor-registry your-harbor.com \
    --harbor-user admin \
    --harbor-pass 'YourPassword123' \
    --push
```

镜像构建完成后，确认 [deploy/k8s/2-services-deployment.yaml](deploy/k8s/2-services-deployment.yaml) 中的镜像地址已改为实际 Harbor 地址。

## Phase 3：按顺序部署到 K8s

推荐使用手动分步部署，便于定位问题：

```bash
kubectl apply -f 0-namespace-rbac.yaml
kubectl apply -f 1-configmap-secret.yaml
kubectl apply -f 4-storage-s3.yaml

# 等待 S3 CSI 驱动就绪
kubectl get pods -n kube-system | grep csi-s3
kubectl rollout status daemonset/csi-s3 -n kube-system --timeout=300s

kubectl apply -f 2-services-deployment.yaml
kubectl apply -f 3-ingress.yaml
```

如果继续使用脚本，也可以：

```bash
bash deploy.sh --namespace mcpcan --harbor-registry your-harbor.com
```

[deploy/k8s/deploy.sh](deploy/k8s/deploy.sh) 现在已经纳入以下能力：
- 部署前校验外部 MySQL / Redis 地址是否仍是默认占位值
- 校验 TOS AccessKey / Secret 是否仍是占位符
- 自动应用 [deploy/k8s/4-storage-s3.yaml](deploy/k8s/4-storage-s3.yaml)
- 自动等待 csi-s3 controller、daemonset 和 PVC 就绪

建议：首轮投产仍优先采用上面的分步部署方式；环境验证完成后，再使用 deploy.sh 进行标准化部署。

---

## 📁 文件说明

| 文件 | 用途 | 描述 |
|------|------|------|
| `0-namespace-rbac.yaml` | Namespace + RBAC | 创建 mcpcan 命名空间、ServiceAccount、ClusterRole、权限绑定 |
| `1-configmap-secret.yaml` | 配置存储 | 所有服务的配置文件（authz.yaml、market.yaml、Traefik config）和敏感信息（密码、密钥） |
| `2-services-deployment.yaml` | 服务部署 | 4 个 Service + 4 个 Deployment（authz、market、web、entry），依赖外部 MySQL / Redis |
| `3-ingress.yaml` | 网络入口 | Nginx Ingress 配置（路由规则、跨域、速率限制等） |
| `4-storage-s3.yaml` | S3 持久化存储 | 火山云 TOS 的 CSI 驱动、StorageClass、PVC 配置 |
| `build-images.sh` | 镜像构建 | 构建 3 个 Docker 镜像（authz、market、web），支持推送到 Harbor |
| `deploy.sh` | 部署脚本 | 主部署脚本，包含静态清单校验、TOS 存储部署、资源等待和状态展示 |
| `values.yaml` | 参数参考 | 当前部署参数核对清单，需手动同步到各 YAML 清单 |
| `README.md` | 本文件 | 完整部署说明 |

---

## 🔧 常见场景和用法

### 场景 1：使用外部 MySQL 和 Redis

```yaml
# 修改 1-configmap-secret.yaml 中 market.yaml / authz.yaml 的数据库配置
mysql:
    host: mysql.company.com
    port: 3306
    database: mcp_dev
    username: mcp_user
    password: your_secure_password

redis:
    host: redis.company.com
    port: 6379
    password: your_redis_password
```

当前 [deploy/k8s/2-services-deployment.yaml](deploy/k8s/2-services-deployment.yaml) 已经不再包含 MySQL 和 Redis Deployment，无需额外删除。

### 场景 2：使用火山云 TOS 作为共享存储

```yaml
# 修改 4-storage-s3.yaml
stringData:
    accessKeyID: "AK..."
    secretAccessKey: "SK..."
    endpoint: "https://tos-s3-cn-beijing.volces.com"
    region: "cn-beijing"

parameters:
    bucket: "mcpcan-data"
```

部署后，应用目录映射关系为：
- authz → `/data/mcpcan`
- market → `/data/mcpcan`
- web → `/app/static`（来自 PVC 的 `static` 子目录）

### 场景 3：启用 HTTPS/TLS

```yaml
# 先按 values.yaml 规划参数，再实际修改 3-ingress.yaml / Secret
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

### 场景 4：多副本高可用部署

```yaml
# 先按 values.yaml 规划副本数，再实际修改 2-services-deployment.yaml
AUTHZ_REPLICAS=3
MARKET_REPLICAS=3
WEB_REPLICAS=2
ENTRY_REPLICAS=2
```

### 场景 5：集群中无外部 LoadBalancer，需要 NodePort 暴露

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

### 场景 6：制定计划内维护，快速更新镜像版本

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
# 检查 ConfigMap 中的数据库配置是否正确
kubectl get configmap mcpcan-config -n mcpcan -o yaml

# 在集群内测试到外部 MySQL / Redis 的网络可达性
kubectl run -n mcpcan net-debug --image=busybox --rm -it -- sh
nc -zv mysql.company.com 3306
nc -zv redis.company.com 6379
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
# 检查 S3 CSI 驱动是否就绪
kubectl get pods -n kube-system | grep csi-s3

# 查看 StorageClass
kubectl get storageclass volcengine-tos

# 查看 PVC 绑定状态
kubectl get pvc -n mcpcan

# 查看 PVC 事件
kubectl describe pvc mcpcan-data-pvc -n mcpcan
```

### TOS 挂载问题

```bash
# 查看 csi-s3 控制器日志
kubectl logs -n kube-system statefulset/csi-s3-controller -c csi-s3

# 查看 csi-s3 节点驱动日志
kubectl logs -n kube-system daemonset/csi-s3 -c csi-s3

# 查看节点侧注册日志
kubectl logs -n kube-system daemonset/csi-s3 -c node-driver-registrar
```

重点检查：
- TOS endpoint 是否与 bucket 所在地域一致
- AccessKey / SecretAccessKey 是否正确
- 集群节点是否能访问 `tos-s3-*.volces.com`
- Harbor 是否已同步 `csi-s3` 与 `sig-storage` 相关镜像

---

## 📊 部署状态验证

部署完成后，检查以下指标确保服务正常：

```bash
# 1. S3 CSI 驱动应正常运行
kubectl get pods -n kube-system | grep csi-s3

# 2. PVC 应为 Bound
kubectl get pvc -n mcpcan

# 3. 所有业务 Pod 应为 Running 或 Completed
kubectl get pods -n mcpcan

# 4. 所有 Service 应有 ClusterIP 和 Port
kubectl get svc -n mcpcan

# 5. Ingress 应显示外部 IP（如有 LoadBalancer）或规则
kubectl get ingress -n mcpcan -o wide

# 6. 检查事件日志，确保无异常
kubectl get events -n mcpcan --sort-by='.lastTimestamp'

# 7. 查询 API 健康状态（使用 port-forward）
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

1. **修改默认密码和密钥**
    - 修改 [deploy/k8s/1-configmap-secret.yaml](deploy/k8s/1-configmap-secret.yaml) 中的管理员密码
    - 修改外部 MySQL / Redis 实际生产密码
    - 修改应用 `secret` / `APP_SECRET`
    - 修改 [deploy/k8s/4-storage-s3.yaml](deploy/k8s/4-storage-s3.yaml) 中的 TOS 凭证

2. **启用 TLS/HTTPS**
   - 参考上面的"场景 2：启用 HTTPS/TLS"

3. **配置网络策略（NetworkPolicy）**
   - 参考 `3-ingress.yaml` 中的 NetworkPolicy 示例

4. **镜像仓库安全**
    - 使用私有 Harbor 并启用认证
    - 创建并绑定 `imagePullSecret`
    - 将第三方基础镜像同步到私有仓库，避免节点公网拉取

5. **RBAC 权限最小化**
   - `0-namespace-rbac.yaml` 中的 ClusterRole 可根据实际需求调整

---

## 📞 获取帮助

如遇到问题，可按以下步骤排查：

1. **先看 csi-s3 是否正常** - 这是 TOS 挂载的前提
2. **检查 K8s 事件** - `kubectl get events -n mcpcan`
3. **查看业务 Pod 日志** - `kubectl logs -f deployment/<service>`
4. **查看配置** - `kubectl get configmap/<name> -o yaml`
5. **查看 kube-system 中 csi-s3 日志** - 挂载失败优先查这里
6. **查看 K8s API 服务器日志**（若为自建集群）

---

## 📝 更新日志

- **v2.1.1** - 初始版本，支持私有化 K8s 部署
- 支持外部 MySQL / Redis
- 支持火山云 TOS 作为共享文件存储
- 支持多副本高可用
- 自动化脚本和参数化配置

---

## 📄 许可证

MCPCan K8s 部署脚本遵循与主项目相同的许可证。
