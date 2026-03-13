# MCPCan v2.1 Release Notes

> **Major Version Notice**: v2.x is a complete rewrite of v1.x. **v1.x is end-of-life due to fundamental design flaws and will no longer receive any PRs, security patches, or bug fixes.** Please migrate to v2.x immediately.

**Scope**: `feature/v1.12` → `feature/v2.1` (mcpcan + mcpcan-tools, 67+ commits)

---

## ⚠️ Breaking Changes: v1.x → v2.x

> **Full reinstall is required. In-place upgrades are not supported.**

| Area | v1.x (Old) | v2.x (New) | Impact |
|------|-----------|-----------|--------|
| **Gateway Architecture** | Standalone global `mcp-gateway` service | Traefik Sidecar injected per-instance; `mcp-gateway` removed | **Breaking** — requires full Helm reinstall |
| **Initialization Service** | Dedicated `mcp-init` container | Init logic merged into `mcp-market` startup; `mcp-init` removed | **Breaking** — remove init from deployment |
| **Image Strategy** | Separate open-source / enterprise images | Single image with runtime `CODE_MODE` env var switch | **Breaking** — update image pull strategy |
| **Helm Chart** | `mcpcan-deploy` repo, chart name `mcpcan-deploy` | Merged into main repo `mcpcan/deploy`, chart name `mcpcan` | **Breaking** — re-add Helm repo |
| **Repository URLs** | Codeup / Gitee | Migrated to GitHub `Kymo-MCP` org | Re-clone or update remotes |
| **Service Boot Order** | No ordering constraint | `mcp-authz` requires `mcp-market` ready (initContainer wait) | Handled automatically on first deploy |
| **Menu/Permission Data** | Manually seeded by `mcp-init` | `mcp-market` idempotently syncs menus on startup (incl. Enterprise menus) | No manual action needed |
| **CodeMode Constant** | No unified convention | Must use `EnterpriseCode` (not `Enterprise`) | Update deployment scripts |

---

## 🏗️ Refactor

- **`refactor(gateway)`**: Replace global `mcp-gateway` with per-instance Traefik Sidecar architecture, eliminating single-point-of-failure (`e55e5a7`)
- **`refactor(ee)`**: Single-image enterprise toggle via `CODE_MODE=EnterpriseCode` env var; no more dual-release images (`5e68c68`)
- **`refactor`**: Centralize container image name generation; unify Sidecar suffix and service naming conventions (`1888a53`, `5f4d8f3`)
- **`refactor`**: Market module init and data seeding fully consolidated into `biz` layer; `mcp-init` service retired (`9a5adfe`)
- **`refactor`**: Standardize container image tags to `latest`, remove platform-specific container config redundancy (`6235a22`)
- **`refactor`**: Improved API error handling with dynamic HTTP status codes and URI parameter support for instance IDs (`3068ba9`)

---

## 🚀 Features

### Gateway & Authentication
- **Traefik gateway integration**: Dynamic routing and authentication with per-instance Sidecar injection (`faa2537`)
- **Enhanced gateway auth**: Detailed request info extraction, hierarchical header management, DB logging for auth events (`ddb07e6`, `6fb9d85`)
- **`toolName` field in gateway logs**: Filter logs by tool name (`c53af79`)
- **`X-Mcp-Authorization` header**: Fallback token retrieval mechanism (`27b83a6`)
- **Gateway proxy handler**: Configurable service name with updated auth middleware for gateway routes (`864dd69`)

### Container & Instance Management
- **Sidecar proxy + hosting modes**: Docker create/copy workflow with ARM64 support (`3f38a02`)
- **Docker multi-platform support**: Config file injection for containers and sidecars, independent AMD64/ARM64 builds (`4bc5da8`, `509b70a`)
- **Dynamic port configuration**: Sidecar and hosting service ports configurable via environment variables (`9dae8e6`, `7717bdd`)
- **New frontend MCP client**: Full support for Streamable HTTP and SSE protocols, integrated into debug tools (`0c38a70`)

### Enterprise Edition
- **Enterprise code mode**: Admin department initialization + GORM data-permission plugin registration (`3be721b`)
- **Enterprise environment config**: Development defaults to OpenCode; dedicated enterprise config provided (`5008a55`)
- **`CODE_MODE` Helm injection**: `mcp-market` and `mcp-authz` receive CodeMode via unified global Helm value (`460d4c3`)
- **Enterprise permission menu auto-sync**: `mcp-market` idempotently writes `mcpcan_rbac_manage` and related system menus on startup based on `CODE_MODE`

### Infrastructure
- **Helm Chart consolidation**: `mcpcan-deploy` repo merged into main repo `mcpcan/deploy`; chart name unified as `mcpcan` (`4340903`)
- **CI/CD integration**: Helm chart deployment integrated into main release workflow (`663bae2`)
- **Static asset decoupling**: System `/static` and user upload directories separated to prevent volume overwrites (`b58945a`, `9693882`)
- **Local dev hot-reload**: Air hot-reload + Docker Compose local development environment (`2c1cabe`, `252756e`)
- **Frontend auth context**: `useAuth` hook with automatic redirect to login for unauthenticated users (`ffa8d07`)
- **OpenCode mode frontend**: Menu display and route auth separated by CodeMode (`141e7da`)

### mcpcan-tools Submodule
- **mcp-sidecar proxy service**: New implementation with multi-architecture build support (`f55d5ac`)
- **Multi-platform image builds**: `openapi-mcp` and `mcp-hosting` support AMD64/ARM64 multi-arch pushes (`fdb4a9d`)
- **SSE endpoint rewriting**: Use `ModifyResponse` for more robust SSE path handling (`9813345`)

---

## 🐛 Bug Fixes

- **Container URL routing fix**: Sidecar container URL routed correctly; hosting protocol uses root path (`6f3d431`)
- **Docker env config field fix**: Correctly populate Docker environment configuration and pass ID to connection handler (`3b0e801`)
- **Dockerfile build context fix**: Standardize source copy to use current build context (`7c69593`)

---

## ⚡ Performance

- **Frontend Dockerfile multi-stage cache**: Reduce redundant layer builds, speed up CI builds (`9e11ea7`)
- **Kubernetes internal URL normalization**: Automatic instance access path normalization, reducing routing errors (`252756e`)
- **Makefile multi-arch refactor**: Independent AMD64/ARM64 compilation with separate registry pushes and parallel build support (`d940953` in mcpcan-tools)

---

## 📚 Documentation

- **Full new documentation**: Covers installation, configuration, and features in both English and Chinese (`7f9b06f`)
- **Helm deployment guide**: Updated chart name from `mcpcan-deploy` → `mcpcan`, pointing to main repo (`9ba72b8`)
- **Deployment instructions consolidated**: Removed Gitee option, unified references to `mcpcan/deploy` directory (`4525bd3`)

---

## 🔧 Chore

- **Version bump**: v2.1 version number updated across all components (`c9abe46`)
- **Cleanup**: Removed redundant Dockerfiles, deprecated `mcp-init`/`mcp-gateway` Helm templates (`a0dd608`, `8b8a752`)
- **Repository migration**: Migrated from Codeup to GitHub `Kymo-MCP` org; `.gitmodules` updated

---

## 📦 Migration Guide

```bash
# 1. Full uninstall required (no in-place upgrade)
helm uninstall <release-name> -n <namespace>
kubectl delete namespace <namespace>

# 2. Re-add Helm repo (chart name has changed)
helm repo add mcpcan https://kymo-mcp.github.io/mcpcan-mainsite
helm repo update

# 3. Install v2.x
helm install mcpcan mcpcan/mcpcan \
  --set global.codeMode=OpenCode \       # Enterprise: EnterpriseCode
  -n <namespace> --create-namespace
```

> For detailed migration documentation, see: [docs/guide/install.md](../../docs/en/guide/install.md)
