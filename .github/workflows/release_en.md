# Release Notes
**Scope**: `main` ... `HEAD` (v2.1.1)

## 🚀 Features
- **[gateway]**: Enhanced `ProxyHandler` request logging to support auditing and interception of structured MCP Tool invocations and responses. (93b07d3)
- **[frontend]**: Added a version number and workflow indicator section to the user's sidebar menu, providing better visibility for enterprise and private-deployment customers. (ca1f97c)
- **[openapi]**: Officially introduced comprehensive support for external enterprise OpenAPI mirrored instances, including dynamic routing fixes and robust UI optimizations. (8aa465b)
- **[template]**: Unified the Proxy and Host proxy form models, introducing a new module for mounting passthrough request headers. (a6da165)
- **[openapi]**: Underlying engine now officially supports appending passthrough interception headers to standard OpenAPI instances. (876ef24)

## 🐛 Bug Fixes
- **[frontend]**: Added the missing `__COMMIT_HASH__` global environment declaration to the TS environment, resolving a major production build failure caused by type-check loop breakage. (0d59a20)
- **[enterprise]**: Refactored to break a dead-lock loop dependency stack when triggering Vite HMR components from the request initiator layer; fixed the `/user-with-role` out-of-bounds 403 route interception and corrected fine-grained enterprise operation directives. (41831a5)
- **[frontend]**: Rectified the flawed route guard mechanism's misjudgment of non-menu whitelists. Forced access validation is now strictly enforced only within actual menu pages. (98563c6)
- **[openapi]**: The backend no longer forcefully overrides internal network path concatenations. The frontend's OpenAPI invocation links safely degrade to a path-less root trigger model to circumvent 504 Gateway errors. (ed5c61b)
- **[k8s]**: Resolved sidecar communication issues and pod namespace preemption conflicts that led to missing or dropped OpenAPI DNS deliveries causing unreachable deadlocks. (01f2b6d)
- **[k8s]**: Repaired the external communication gap for API models based on solid Sidecar namespace deductions and internal port overlaps. (c19ee05)
- **[k8s]**: Implemented idempotent replacements to eliminate Deployment reconstruction storms under high-frequency concurrent operations, injecting debounce thresholds on the frontend UI to resist invalid rapid clicks. (5fba403)
- **[k8s]**: Adjusted the recreation model to Background deletion policy to sidestep the object-lifecycle race condition errors resulting from upstream Delete requests. (493f799)
- **[k8s]**: Added strategies including reconstruction spin-locks and exponential backoffs when discovering existing homonymous entities blocked by lingering deletion connections during the Foreground clearing interval. (e8284e3)
- **[template]**: Removed legacy Notes guard clauses. Now, any access to the `/api/market` gateway domain forcefully stimulates a configuration update to prevent stalling from aging states. (ae32a1f)
- **[template]**: Shifted the focus down to YAML configuration distribution node mounts, detaching from the internally hardcoded `openapi_base_url`. (4b20fcd)
- **[init]**: Appended a backdoor script to automatically synchronize the current version's `openapi_file_id` backwards to accommodate all deprecated or offline template base resource pools. (3703711)
- **[openapi]**: All routes now automatically prepend correct gateway directive prefixes like `/api`, enabling end-users to pair proxy link pools merely by passing the corresponding principal host domains. (0f0f4ec)
- **[gateway]**: Debugged and severed a vulnerability where the gateway prematurely truncated overly long response payload packets, while extending the audit log expiration tolerance to 7 days. (5ccb204)
- **[gateway]**: Reattached and patched the GatewayLog processing hooks at the gateway layer, eliminating potential 404 hazards caused by double slashes arising from OpenAPI base addresses with trailing slashes. (feb409b)
- **[k8s]**: Alleviated frequent repetitive conflict rejection popups when making secondary updates to Service interfaces in the operations panel by utilizing idempotent object overlay patches. (5076a03)
- **[k8s]**: Enforced fetching limits exclusively targeting specific containers (e.g., openapi or sidecar) under the GetLogs interaction layer to avoid multi-container mix-ups making query APIs unable to distinguish log streams resulting in rejected signatures. (24c7f42)
- **[openapi]**: Correctly utilized `/bin/sh` as the embedded interactive layer to bypass container startup probe alarm reboot disconnections triggered when the `bash` runtime is unsupported by micro-images. (988a25b)
- **[k8s]**: Fixed an issue by pushing down Traefik directive domains within underlying load mappings, now adopting annotations instead of labels to complete scheduling handshakes with externalized services. (6fea864)

## ♻️ Refactor
- **[style]**: Comprehensively optimized modules involving redundant parameter mounts when the backend directly passes through to the UI alongside secure display protocols for streamable inner-network connections. (ecdb3be)

## 📚 Documentation
- **[docs]**: Appended bilingual (Chinese/English) white-paper annotations covering internal template lifecycle logic and cross-container topological routing. (dfea43c)
