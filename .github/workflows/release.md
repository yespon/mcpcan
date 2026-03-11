# Release Notes
**Range**: `origin/main` ... `HEAD`

## 🚀 Features
- **[AI Sessions & Models]**: Introduced full AI session management and model permission control. Includes new `ai_model_access.proto` and `ai_session.proto` interface definitions, along with corresponding database models and MySQL storage implementations ([2493e6a]).
- **[Database Initialization]**: Improved MySQL initialization logic to automatically handle the creation and maintenance of AI-related service tables ([dc7c71a]).
- **[Logic Implementation]**: Added `ai_model_access` and `ai_session` service modules to the backend, expanding the codebase by 700+ lines of logic ([dc7c71a]).

## 📚 Documentation
- **[Tech Solution Optimization]**: Reviewed and updated technical documents for AI platform access and MCP invocation, removing outdated descriptions and adding the latest invocation flow details ([9cb0f2d], [266ad9b]).
