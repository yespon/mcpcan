# LLM2 包适配器使用指南

## 简介

此文档说明如何在现有代码中使用 `llm2` 包来替代原始的 `llm` 包，以利用基于 one-api relay 模式的改进架构。

## 迁移步骤

### 1. 更新导入语句

在需要使用 LLM 功能的文件中，将：

```go
import "github.com/kymo-mcp/mcpcan/pkg/llm"
```

替换为：

```go
import "github.com/kymo-mcp/mcpcan/pkg/llm2"
```

### 2. 更新提供商初始化代码

原始代码：
```go
provider, err := llm.NewProvider(llm.ProviderOpenAI, llm.ProviderConfig{
    BaseURL: "https://api.openai.com/v1",
    APIKey: "your-api-key",
})
```

更新为：
```go
provider, err := llm2.NewProvider(llm2.ProviderOpenAI, llm2.ProviderConfig{
    BaseURL: "https://api.openai.com/v1",
    APIKey: "your-api-key",
})
```

### 3. 更新接口引用

所有接口、类型和方法名保持一致，只需将包前缀从 `llm.` 更改为 `llm2.`：

- `llm.Provider` → `llm2.Provider`
- `llm.Message` → `llm2.Message`
- `llm.ChatRequest` → `llm2.ChatRequest`
- `llm.StreamResponse` → `llm2.StreamResponse`
- `llm.ChatRequest` → `llm2.ChatRequest`

### 4. 保持不变的调用

以下调用方式保持不变，因为接口签名完全兼容：

```go
// StreamChat 调用方式不变
stream, err := provider.StreamChat(ctx, req)

// 消息处理方式不变
for resp := range stream {
    if resp.Error != nil {
        // 错误处理
    }
    // 响应处理
}
```

## 在 ai_session.go 中的具体实现

在 `internal/market/biz/ai_session.go` 中：

1. 修改导入语句
2. 替换提供商初始化部分
3. 保持 StreamChat 和消息处理逻辑不变

## 示例对比

### 原始实现（使用原始 llm 包）：
```go
func (b *AiSessionBiz) Chat(ctx context.Context, req *pb.ChatRequest) (<-chan llm.StreamResponse, error) {
    // ...
    provider, err := llm.NewProvider(providerType, llm.ProviderConfig{
        BaseURL: modelAccess.BaseUrl,
        APIKey:  modelAccess.ApiKey,
    })
    if err != nil {
        return nil, err
    }
    
    stream, err := provider.StreamChat(ctx, llm.ChatRequest{
        Model:    session.ModelName,
        Messages: messages,
        Tools:    tools,
    })
    // ...
}
```

### 更新实现（使用 llm2 包）：
```go
func (b *AiSessionBiz) Chat(ctx context.Context, req *pb.ChatRequest) (<-chan llm2.StreamResponse, error) {
    // ...
    provider, err := llm2.NewProvider(llm2.ProviderType(providerType), llm2.ProviderConfig{
        BaseURL: modelAccess.BaseUrl,
        APIKey:  modelAccess.ApiKey,
    })
    if err != nil {
        return nil, err
    }
    
    stream, err := provider.StreamChat(ctx, llm2.ChatRequest{
        Model:    session.ModelName,
        Messages: messages,
        Tools:    tools,
    })
    // ...
}
```

## 优势提升

1. **更稳健的数据流解析** - 基于 one-api relay 模式的流处理
2. **更好的错误处理** - 更详细的错误信息和恢复机制
3. **改进的 MCP 支持** - 专门优化的 MCP 工具集成
4. **更高的性能** - 优化的请求和响应处理

## 验证

迁移后验证以下几点：

1. 聊天功能正常工作
2. 工具调用正常执行
3. 错误处理按预期工作
4. 性能表现符合预期