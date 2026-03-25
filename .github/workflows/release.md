# Release Notes

**Changes**: `v2.1.1` ... `v2.1.2`

---

## 🐛 Bug Fixes

- **Intelligent Agent Sync Callback Domain Fix**: Fixed the "domain is required" error that occurred when synchronizing MCP instances to platforms such as Dify, Coze, and N8N. The root cause was the backend's unreliable reliance on the `Market.Host` environment variable, which was empty in most Proxy/Hosting deployments. The domain is now injected directly from the frontend via `window.location.origin` and carried through the API request payload, enabling the backend to correctly construct callback URLs without any environment variable dependency. (`beefa15`)

- **LLM Streaming Tool Call JSON Parameter Collision Fix**: Fixed a bug where multiple concurrent tool calls in a streaming LLM response were incorrectly merged into a single call due to a missing `index` field to distinguish them. This primarily affected models like MiniMax and caused 400 API errors. Added an `*int` typed `index` field to the `oaiToolCall` struct to correctly group and accumulate streaming chunks per individual tool call. (`60b32ae`)

---

## 🔧 CI/CD & Chore

- **Switch Release Trigger from Main Branch Push to Tag Push**: Updated the GitHub Actions Release workflow trigger from pushing to the `main` branch to pushing a version tag matching `v*.*.*`. This prevents accidental releases triggered by routine code merges, while retaining the `workflow_dispatch` option for manual re-runs. (`0e37330`)

- **Version Bump to v2.1.2**: Updated the `VERSION` file and the image tags in the Helm Chart `values.yaml` to `v2.1.2`. (`6d72f06`)
