# AI 平台接入与 MCP 调用技术方案 (V3)

## 1. 核心需求变更

根据最新需求，我们需要明确区分以下概念，并增强会话管理能力：

1.  **AI Model Access (AI 模型接入)**: 专门用于管理 **LLM 提供商** 的凭证 (如 OpenAI, Anthropic, DeepSeek 的 API Key)。
2.  **MCPCan API Key (平台密钥)**: 供外部开发者/调试工具调用 MCPCan 平台接口的凭证。
3.  **AI Session Management (会话管理)**: 支持多会话窗口、重命名、历史记录持久化。
4.  **Tool Selection (工具选择)**: 每个会话可单独配置启用的 MCP 工具。
5.  **SSE Streaming**: 全链路使用 Server-Sent Events 实现实时打字机效果。

## 2. 总体架构设计

### 2.1. 服务拆分

| 服务名称               | 职责                                   | 数据表 (建议)              |
| :--------------------- | :------------------------------------- | :------------------------- |
| **AI Model Access**    | 管理原始 LLM API Key 和 BaseURL        | `ai_model_access`          |
| **AI Session Service** | 管理会话窗口、历史记录、工具绑定       | `ai_session`, `ai_message` |
| **User API Key**       | 管理用户调用 MCPCan 的开发者密钥       | `user_api_key`             |
| **AI Agent Service**   | 核心对话引擎，处理 "SSE + LLM + Tools" | 无 (纯逻辑服务)            |

### 2.2. 数据流向 (Run Phase)

1.  **初始化**: 前端调用 `AI Session Service` 创建/获取会话，加载历史消息和绑定的工具列表。
2.  **交互**:
    - 前端建立 SSE 连接: `POST /api/v1/chat/sse` (携带 `session_id`, `prompt`)。
    - 后端 `AI Agent Service`:
      - 验证 Session 归属。
      - 加载 `AiModelAccess` (LLM Key) 和 `Selected Tools` (MCP Definitions)。
      - 加载 `AiMessage` (Context Window)。
    - **Agent Loop**:
      - 请求 LLM -> 收到 `ToolCall` -> **Stream Event: tool_start** -> 执行 MCP 工具 -> **Stream Event: tool_result** -> 回传 LLM -> 收到文本 -> **Stream Event: text_delta**。
    - **持久化**: 异步将 User Prompt, Tool Inputs/Outputs, Assistant Reply 写入 `ai_message`。

## 3. 详细技术方案

### 3.1. AI Session & History (会话层)

负责管理聊天窗口状态和历史记录。

- **数据模型 (`model/ai_session.go`)**:

  ```go
  type AiSession struct {
      ID            int64           `gorm:"primaryKey"`
      UserID        int64           `gorm:"index"`
      Name          string          `gorm:"size:255"` // 会话标题 (可重命名)
      ModelAccessID int64           // 绑定的 LLM 配置
      ToolsConfig   json.RawMessage // 启用的工具列表 ["mcp-server-1:toolA", ...]
      CreateTime    time.Time
      UpdateTime    time.Time
  }
  ```

- **数据模型 (`model/ai_message.go`)**:

  ```go
  type AiMessage struct {
      ID         int64  `gorm:"primaryKey"`
      SessionID  int64  `gorm:"index"`
      Role       string // system, user, assistant, tool
      Content    string `gorm:"type:text"`
      ToolCalls  string `gorm:"type:text"` // JSON: 存储工具调用参数
      ToolCallID string // 关联的 tool_call_id
      CreateTime time.Time
  }
  ```

- **API 接口**:
  - `POST /sessions`: 创建新会话 (Name, ModelID, ToolIDs)。
  - `PUT /sessions/:id`: 修改标题或绑定的工具。
  - `GET /sessions`: 列表。
  - `GET /sessions/:id/messages`: 分页获取历史记录。

### 3.2. AI Model Access (模型层)

独立管理 AI 厂商凭证。

- **数据模型 (`model/ai_model_access.go`)**:
  ```go
  type AiModelAccess struct {
      ID         int64  `gorm:"primaryKey"`
      UserID     int64  `gorm:"index"`
      Name       string // e.g. "My DeepSeek"
      Provider   string // openai, azure, deepseek
      ApiKey     string // 加密存储
      BaseUrl    string
      ModelName  string // e.g. "deepseek-chat"
  }
  ```

### 3.3. User API Key (开发者层)

外部调用鉴权。

- **数据模型 (`model/user_api_key.go`)**:
  ```go
  type UserApiKey struct {
      ID     int64
      Key    string // Hash
      Status int
      // ...
  }
  ```

### 3.4. AI Agent Service (执行层 - SSE Core)

核心引擎，处理长连接和工具调用循环。

- **接口定义**:

  ```go
  // POST /api/v1/chat/sse
  type ChatRequest struct {
      SessionID int64  `json:"session_id"`
      Prompt    string `json:"prompt"`
  }
  ```

- **SSE 事件协议 (Server-Sent Events)**:
  客户端通过 EventSource 或 fetch stream 接收：

  1.  `event: message` -> `data: "H"` (文本增量)
  2.  `event: message` -> `data: "e"`
  3.  `event: tool_use` -> `data: {"name": "get_weather", "args": "..."}` (工具调用通知)
  4.  `event: tool_result` -> `data: "Sunny"` (工具执行结果)
  5.  `event: done` -> `data: [DONE]` (结束)

- **核心逻辑 (`service/ai_agent.go`)**:
  1.  **Context Loading**:
      - `session = sessionRepo.Get(req.SessionID)`
      - `history = messageRepo.GetLastN(req.SessionID, 20)`
      - `tools = mcpClient.GetTools(session.ToolsConfig)`
  2.  **OpenAI Stream Loop**:
      - 初始化 `go-openai` client。
      - `stream, err := client.CreateChatCompletionStream(...)`
      - `for { resp, err := stream.Recv() ... }`
  3.  **Tool Execution**:
      - 如果检测到 `FinishReason == "tool_calls"`, 暂停 Stream。
      - 执行 MCP 工具: `result = mcpClient.CallTool(...)`
      - 将结果作为 `ToolMessage` 插入 context。
      - **递归调用**: 带着新 context 再次发起 LLM 请求 (CreateChatCompletionStream)。
  4.  **Persistence**:
      - 整个流程结束后，异步将 User Prompt, Assistant Full Response, Tool Interactions 存入 `ai_message` 表。

## 4. 开发路线图 (Roadmap)

1.  **Phase 1: 数据层与配置 (Foundations)**

    - [ ] 创建 `AiModelAccess`, `AiSession`, `AiMessage`, `UserApiKey` 数据库模型。
    - [ ] 实现 Session CRUD 和 Model Config CRUD 接口。

2.  **Phase 2: 核心引擎 (Agent Engine)**

    - [ ] 引入 `go-openai`。
    - [ ] 实现 `MCP Client` 的 `ListTools` 和 `CallTool`。
    - [ ] 实现 SSE Handler：打通 "LLM -> SSE Client" 的流式传输。

3.  **Phase 3: 工具链与历史 (Integration)**

    - [ ] 实现 "LLM -> Tool Call -> MCP -> LLM" 的递归调用逻辑。
    - [ ] 实现历史记录的自动加载与持久化。

4.  **Phase 4: 外部调用 (External API)**
    - [ ] 实现 `UserApiKey` 鉴权中间件，允许外部通过 API Key 调用 Agent 接口。

## 5. 现有代码调整

- **Proto**: 创建 `ai_session.proto`, `ai_model.proto` 等。
- **Service**: 保持 `intelligent_access` 纯净，新建独立服务。
