# 版本发布说明 (Release Notes)

**统计范围**: `v2.1.1` ... `v2.1.2`

---

## 🐛 问题修复 (Bug Fixes)

- **智能体同步回调域名修复**: 修复了将 MCP 实例同步到 Dify / Coze / N8N 等平台时，因后端依赖无法可靠读取的环境变量 `Market.Host` 导致的"domain is required"错误。现在由前端通过 `window.location.origin` 将当前访问域名随请求一并传递给后端，从而正确构建回调地址，彻底消除了 Proxy / Hosting 模式下的域名缺失问题。 (`beefa15`)

- **LLM 流式工具调用 JSON 参数碰撞修复**: 修复了多个工具调用（Tool Calls）在并发流式响应中，因缺少 `index` 字段区分而导致参数字符串被错误拼接到同一个 Call 上的问题（主要影响 MiniMax 等模型，触发 400 错误）。在 `oaiToolCall` 结构体中新增 `*int` 类型的 `index` 字段，以正确区分并分组各个并发工具调用的流式 chunk。 (`60b32ae`)

---

## 🔧 CI/CD 优化 (Chore / CI)

- **Release 触发机制改为 Tag 推送**: 将 GitHub Actions Release 工作流的触发条件从 `main` 分支 push 改为正式版本 Tag（`v*.*.*`）push，避免合并代码误触发发版，同时保留 `workflow_dispatch` 手动触发能力。 (`0e37330`)

- **版本号升级至 v2.1.2**: 更新 `VERSION` 文件及 Helm Chart `values.yaml` 中的镜像标签至 `v2.1.2`。 (`6d72f06`)
