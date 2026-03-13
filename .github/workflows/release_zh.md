# MCPCan v2.1 版本发布说明 (Release Notes)

> **重大版本声明**：v2.x 是对 v1.x 架构的完全重写，**v1.x 系列因存在根本性设计缺陷，即日起停止维护，不接受任何 PR 或 Issue**。请尽快迁移到 v2.x。

**统计范围**：`feature/v1.12` → `feature/v2.1`（mcpcan + mcpcan-tools）

---

## ⚠️ 升级必读：v1.x → v2.x 重大破坏性变更

| 变更点 | v1.x 旧设计 | v2.x 新设计 | 影响 |
|--------|------------|------------|------|
| **网关架构** | 独立 `mcp-gateway` 全局网关服务 | Traefik Sidecar 按实例注入，`mcp-gateway` 服务已废弃 | **不兼容**：Helm 需全量重装 |
| **初始化服务** | `mcp-init` 独立初始化容器 | 初始化逻辑合并进 `mcp-market`，`mcp-init` 已废弃 | **不兼容**：移除 init 部署 |
| **镜像策略** | 开源版/企业版双镜像 | 单镜像 + 运行时 `CODE_MODE` 环境变量切换 | **不兼容**：更新镜像拉取策略 |
| **Helm Chart** | `mcpcan-deploy` 独立仓库，Chart 名 `mcpcan-deploy` | 合并进主仓库 `mcpcan/deploy`，Chart 名改为 `mcpcan` | **不兼容**：重新添加 Helm repo |
| **子模块 URL** | Codeup / Gitee | 统一迁移至 GitHub `Kymo-MCP` 组织 | 需重新 clone 或更新 remote |
| **服务启动顺序** | 无顺序约束 | `mcp-authz` 强依赖 `mcp-market` 先就绪（initContainer 等待） | 首次部署自动处理 |
| **菜单/权限数据** | 由 `mcp-init` 手动注入 | `mcp-market` 启动时幂等自动同步（含企业版专属菜单） | 无需手动操作 |
| **CodeMode 常量值** | 无统一约定 | 必须使用 `EnterpriseCode`（非 `Enterprise`） | 部署脚本需更新 |

---

## 🏗️ 架构重构 (Refactor)

- **`refactor(gateway)`**: 废弃全局 `mcp-gateway` 服务，全面切换至 Traefik Sidecar 架构，每个实例独立代理，消除单点瓶颈 (`e55e5a7`)
- **`refactor(ee)`**: 实现单镜像 + 运行时企业版特性开关，`CODE_MODE=EnterpriseCode` 环境变量控制，不再维护双发行版镜像 (`5e68c68`)
- **`refactor`**: 集中化容器镜像名称生成逻辑，统一 Sidecar 容器后缀与服务名称命名规范 (`1888a53`, `5f4d8f3`)
- **`refactor`**: market 模块初始化与数据 Seeding 全部收归 `biz` 层统一管理，废弃 `mcp-init` 服务 (`9a5adfe`)
- **`refactor`**: 统一容器镜像 tag 为 `latest`，移除平台专属配置冗余 (`6235a22`)
- **`refactor`**: API 错误处理改进，支持动态 HTTP 状态码 + URI 参数实例 ID (`3068ba9`)

---

## 🚀 新功能 (Features)

### Gateway 与认证
- **Traefik 网关集成**：实现动态路由与认证，支持按实例注入 Traefik Sidecar (`faa2537`)
- **网关认证增强**：详细请求信息提取、分层 Header 管理、成功/失败事件数据库日志 (`ddb07e6`, `6fb9d85`)
- **新增 `toolName` 字段**：网关日志支持按工具名过滤 (`c53af79`)
- **X-Mcp-Authorization Header**：新增 fallback Token 提取机制 (`27b83a6`)
- **网关代理 Handler**：支持可配置 Service Name，更新网关路由鉴权中间件 (`864dd69`)

### 容器与实例管理
- **Sidecar Proxy + Hosting 模式**：Docker create/copy 工作流，支持 ARM64 (`3f38a02`)
- **Docker 多平台支持**：配置文件注入 Sidecar/宿主容器，独立 ARM64/AMD64 编译 (`4bc5da8`, `509b70a`)
- **动态端口配置**：Sidecar 与 Hosting 服务端口通过环境变量灵活配置 (`9dae8e6`, `7717bdd`)
- **前端 MCP 客户端**：全新实现，支持 Streamable HTTP 和 SSE 双协议，集成调试工具 (`0c38a70`)

### 企业版 (Enterprise Edition)
- **企业版代码模式**：admin 部门初始化 + GORM 数据权限插件注册 (`3be721b`)
- **企业版环境配置**：开发默认切换 OpenCode，提供企业版独立配置 (`5008a55`)
- **`CODE_MODE` Helm 注入**：`mcp-market` + `mcp-authz` 统一通过全局 Helm value 注入 (`460d4c3`)
- **企业版权限菜单自动同步**：`mcp-market` 启动时按 `CODE_MODE` 幂等写入 `mcpcan_rbac_manage` 等系统菜单

### 基础设施
- **Helm Chart 整合**：`mcpcan-deploy` 仓库合并进主仓库 `mcpcan/deploy`，Chart 名统一为 `mcpcan` (`4340903`)
- **CI/CD 集成**：Helm Chart 部署集成进主发布工作流 (`663bae2`)
- **静态资源解耦**：系统 static 与 uploads 目录分离，防止 Volume 覆盖 (`b58945a`, `9693882`)
- **本地开发热重载**：Air 热重载 + Docker Compose 本地开发环境 (`2c1cabe`, `252756e`)
- **前端认证上下文**：`useAuth` Hook 实现，未登录自动跳转登录页 (`ffa8d07`)
- **OpenCode 模式前端**：不同 CodeMode 下菜单展示与路由鉴权分离 (`141e7da`)

### mcpcan-tools 子模块
- **mcp-sidecar 代理服务**：全新实现，支持多架构构建 (`f55d5ac`)
- **多平台镜像构建**：`openapi-mcp` 和 `mcp-hosting` 支持 AMD64/ARM64 多架构推送 (`fdb4a9d`)
- **SSE 端点重写**：使用 `ModifyResponse` 优化路径处理 (`9813345`)

---

## 🐛 问题修复 (Bug Fixes)

- **容器 URL 路由修复**：Sidecar 容器 URL 正确路由，Hosting 协议使用根路径 (`6f3d431`)
- **Docker 环境配置字段修复**：正确填充并修复环境配置，ID 传递修复 (`3b0e801`)
- **Dockerfile 构建上下文修复**：统一使用当前构建上下文的 source copy (`7c69593`)

---

## ⚡ 性能优化 (Performance)

- **前端 Dockerfile 多阶段缓存优化**：减少重复层构建，提升 CI 构建速度 (`9e11ea7`)
- **Kubernetes 内部 URL 规范化**：实例访问路径自动规范，减少路由错误 (`252756e`)
- **Makefile 多架构重构**：独立 AMD64/ARM64 编译与分别推送，并行构建支持 (`d940953` in mcpcan-tools)

---

## 📚 文档 (Documentation)

- **全量新文档**：覆盖安装、配置、功能的中英双语指南 (`7f9b06f`)
- **Helm 部署指南更新**：Chart 名从 `mcpcan-deploy` → `mcpcan`，指向主仓库 (`9ba72b8`)
- **部署说明整合**：移除 Gitee 选项，统一指向主仓库 `deploy` 目录 (`4525bd3`)

---

## 🔧 其他 (Chore)

- **版本升级**：v2.1 版本号全量更新 (`c9abe46`)
- **依赖清理**：移除冗余 Dockerfile、废弃的 mcp-init/mcp-gateway Helm 模板 (`a0dd608`, `8b8a752`)
- **仓库迁移**：从 Codeup 迁移至 GitHub `Kymo-MCP` 组织，`.gitmodules` 更新

---

## 📌 升级指引

```bash
# 1. 全量重装（必须，不支持原地升级）
helm uninstall <release-name> -n <namespace>
kubectl delete namespace <namespace>

# 2. 添加新 Helm 仓库（Chart 名已变更）
helm repo add mcpcan https://kymo-mcp.github.io/mcpcan-mainsite
helm repo update

# 3. 安装 v2.x
helm install mcpcan mcpcan/mcpcan \
  --set global.codeMode=OpenCode \   # 企业版使用 EnterpriseCode
  -n <namespace> --create-namespace
```

> 详细迁移文档请参阅：[docs/upgrade-v2.md](../../docs/)
