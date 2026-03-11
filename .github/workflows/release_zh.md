# 版本发布说明 (Release Notes)
**统计范围**: `origin/main` ... `HEAD`

## 🚀 新功能 (Features)
- **[AI会话与模型]**: 引入全套 AI 会话管理及模型权限控制功能。包括新增 `ai_model_access.proto` 和 `ai_session.proto` 接口定义，及其对应的数据库模型与 MySQL 存储层实现 ([2493e6a])。
- **[数据库初始化]**: 完善 MySQL 数据库初始化逻辑，自动处理 AI 相关业务表的创建与维护 ([dc7c71a])。
- **[逻辑实现]**: 后端业务层新增 `ai_model_access` 和 `ai_session` 服务模块，支持 700+ 行代码逻辑扩展 ([dc7c71a])。

## 📚 文档 (Documentation)
- **[技术方案优化]**: 重新审视并更新了 AI 平台接入与 MCP 调用的技术方案文档，移除过时描述并补充了最新的调用链路说明 ([9cb0f2d], [266ad9b])。
