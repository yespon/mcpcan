# MCPCan K8s 部署 - 快速参考

## 📌 3 步快速部署

### 1️⃣ 配置（2 分钟）
```bash
cd deploy/k8s
vim values.yaml
# 修改这 3 项：
# HARBOR_REGISTRY=your-harbor.com
# HARBOR_USER=admin  
# HARBOR_PASS=password
```

### 2️⃣ 构建镜像（5-10 分钟）
```bash
bash build-images.sh --harbor-registry your-harbor.com --push
```

### 3️⃣ 部署（3-5 分钟）
```bash
bash deploy.sh --namespace mcpcan --harbor-registry your-harbor.com
```

## 🔗 Harbor 私有仓库配置

### 示例配置
```bash
Harbor 地址：harbor.company.com/mcpcan
用户名：admin
密码：Harbor@123
```

构建脚本命令：
```bash
bash build-images.sh \
    --harbor-registry harbor.company.com \
    --harbor-user admin \
    --harbor-pass Harbor@123 \
    --push
```

## 常用命令速查

### Pod 和服务管理
```bash
# 查看 Pod 状态
kubectl get pods -n mcpcan

# 查看服务
kubectl get svc -n mcpcan

# 查看 Ingress
kubectl get ingress -n mcpcan

# 重启服务
kubectl rollout restart deployment/mcp-market -n mcpcan

# 查看日志
kubectl logs -f deployment/mcp-market -n mcpcan
```

### 故障排查
```bash
# Pod 详情（包括错误信息）
kubectl describe pod <pod-name> -n mcpcan

# 进入 Pod 调试
kubectl exec -it pod/<pod-name> -n mcpcan -- /bin/sh

# 查看事件
kubectl get events -n mcpcan --sort-by='.lastTimestamp'

# 查看配置
kubectl describe configmap mcpcan-config -n mcpcan
```

## 访问方式

### LoadBalancer 类型（云平台）
```
Web 前端：http://<EXTERNAL_IP>
API 网关：http://<EXTERNAL_IP>/api
Traefik Dashboard：http://<EXTERNAL_IP>:8080
```

### NodePort 类型（私有集群）
```
Web 前端：http://<Node_IP>:30080
API 网关：http://<Node_IP>:30080/api
```

### ClusterIP 类型（本地调试）
```bash
# Port forward
kubectl port-forward -n mcpcan svc/mcp-entry-svc 8080:80 &
# 访问：http://localhost:8080
```

## 默认凭证

- 用户名：`admin`
- 密码：`admin123`

⚠️ **生产环境必须修改！** 修改 values.yaml

## 常见问题排查

### 镜像无法拉取
```bash
docker login your-harbor.com
docker pull your-harbor.com/mcpcan/mcp-market:v2.1.1
```

### Pod 一直 Pending
```bash
kubectl describe pod <name> -n mcpcan
# 查看 Events 部分，通常是存储或资源不足
```

### 无法访问服务
```bash
# 检查 Service 配置
kubectl get svc -n mcpcan -o wide

# 测试连接
kubectl run debug --image=busybox --rm -it -- wget http://mcp-authz-svc:8081/health
```

### 数据库连接失败
```bash
# 检查 MySQL Pod
kubectl get pods -n mcpcan | grep mysql

# 查看日志
kubectl logs -f pod/mysql-xxx -n mcpcan
```

## 外部数据库配置

### 使用外部 MySQL
```bash
# 1. values.yaml 修改
USE_INTERNAL_MYSQL=false
EXTERNAL_MYSQL_HOST=mysql.company.com
EXTERNAL_MYSQL_DATABASE=mcp_dev

# 2. 删除 MySQL Deployment（从 2-services-deployment.yaml）

# 3. 重新部署
bash deploy.sh
```

## 存储路径要求

每个工作节点需要此目录（至少 100GB）：
```
/data/mcpcan/
├── code-package/    # 代码包
├── static/         # 静态文件
├── openapi-file/   # OpenAPI 文件
├── mysql/          # MySQL 数据（若使用内部）
└── redis/          # Redis 数据（若使用内部）
```

## 版本管理

```bash
# 查看当前版本
cat VERSION

# 更新版本
echo "v2.1.2" > VERSION

# 重新构建镜像
bash build-images.sh --push

# 更新 Deployment
kubectl set image deployment/mcp-market \
    mcp-market=your-harbor.com/mcpcan/mcp-market:v2.1.2 \
    -n mcpcan
```

## 安全检查清单

- [ ] 修改 APP_SECRET（values.yaml）
- [ ] 修改所有数据库密码
- [ ] 修改默认管理员密码
- [ ] 启用 TLS/HTTPS（参考 ARCHITECTURE.md）
- [ ] 配置网络策略
- [ ] 使用私有 Harbor 仓库

## 文档导航

| 文档 | 说明 |
|------|------|
| `README.md` | 完整部署指南 |
| `ARCHITECTURE.md` | 架构设计和原理 |
| `values.yaml` | 所有参数说明 |
| `QUICK_REFERENCE.md` | 本文件 |
