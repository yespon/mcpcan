# LLM2 - 基于 one-api relay 模式的 LLM 提供商

## 概述

LLM2 包是完全基于 one-api 项目的 relay 模式实现的新一代 LLM 提供商包，旨在解决原 `pkg/llm` 包中的数据流解析错误问题，并提供更好的 MCP 工具调试和聊天功能支持。

## 设计理念

此包完全借鉴 one-api 项目的 relay 架构模式：

1. **适配器模式** - 针对每种 LLM 提供商提供专门的适配器
2. **统一接口** - 所有提供商实现统一的接口
3. **请求转换** - 统一请求格式转换机制
4. **响应处理** - 标准化的响应处理流程
5. **流式支持** - 稳定的 SSE（Server-Sent Events）流处理

## 架构组件

### 1. 核心组件

- `adaptor/` - 提供商适配器接口和实现
- `model/` - 数据模型定义
- `meta/` - 请求元数据管理
- `channeltype/` - 渠道类型定义
- `relaymode/` - 中继模式定义

### 2. 适配器实现

- `adaptor/openai/` - OpenAI 兼容提供商适配器
- `adaptor/mcp/` - MCP (Model Context Protocol) 适配器

## 功能特性

1. **多提供商支持** - 支持所有在 `types.go` 中定义的提供商类型
2. **改进的流式处理** - 基于 one-api 的稳健 SSE 流解析
3. **MCP 集成** - 专为 MCP 工具调试优化
4. **错误处理** - 更好的错误传播和恢复机制
5. **扩展性** - 易于添加新的提供商适配器

## 使用方法

### 基本用法

```go
import "github.com/kymo-mcp/mcpcan/pkg/llm2"

// 创建提供商
config := llm2.ProviderConfig{
    BaseURL: "https://api.openai.com/v1",
    APIKey:  "your-api-key",
}

provider, err := llm2.NewProvider(llm2.ProviderOpenAI, config)
if err != nil {
    log.Fatal(err)
}

// 执行聊天请求
req := llm2.ChatRequest{
    Model: "gpt-4o",
    Messages: []llm2.Message{
        {
            Role:    "user",
            Content: "Hello, how can you help me?",
        },
    },
    Stream: true,
}

stream, err := provider.StreamChat(context.Background(), req)
if err != nil {
    log.Fatal(err)
}

// 处理流响应
for resp := range stream {
    if resp.Error != nil {
        log.Printf("Error: %v", resp.Error)
        break
    }
    fmt.Printf("Response: %s\n", resp.Content)
}
```

### MCP 提供商使用

```go
// MCP 配置
mcpConfig := llm2.ProviderConfig{
    BaseURL: "http://localhost:8080/mcp",
    APIKey:  "mcp-config",
}

mcpProvider, err := llm2.NewProvider(llm2.ProviderMCP, mcpConfig)
if err != nil {
    log.Fatal(err)
}
```

## 适配器模式

one-api relay 模式的核心在于适配器模式，每种提供商实现相同的接口：

```go
type Adaptor interface {
    Init(meta *meta.Meta)
    GetRequestURL(meta *meta.Meta) (string, error)
    SetupRequestHeader(ctx context.Context, req *http.Request, meta *meta.Meta) error
    ConvertRequest(ctx context.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error)
    ConvertImageRequest(request *model.ImageRequest) (any, error)
    DoRequest(ctx context.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error)
    DoResponse(ctx context.Context, c GinContext, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode)
    GetModelList() []string
    GetChannelName() string
}
```

## 优势对比

| 特性 | 旧 pkg/llm | 新 pkg/llm2 |
|------|------------|-------------|
| 架构模式 | LangChainGo 封装 | one-api relay 模式 |
| 流式处理 | 不稳定 | 稳健的 SSE 解析 |
| MCP 集成 | 复杂 | 优化的工具支持 |
| 扩展性 | 有限 | 高度可扩展 |
| 多提供商 | 基础支持 | 统一适配器模式 |
| 错误处理 | 一般 | 详细错误信息 |

## 支持的提供商

根据 `types.go` 中定义的常量，包支持以下提供商：

- OpenAI 系列 (OpenAI, Azure OpenAI)
- 大型模型提供商 (Anthropic, Google, Mistral, XAI)
- 中国提供商 (Qwen, Doubao, Zhipu, Moonshot)
- 自托管模型 (Ollama, LiteLLM)
- MCP (Model Context Protocol)

## 开发指南

要添加新提供商，只需：

1. 在 `adaptor/` 下创建新适配器实现
2. 实现 `Adaptor` 接口
3. 在 `GetAdaptor` 函数中添加路由

## 性能特点

- 连接复用
- 请求/响应转换优化
- 流式处理效率高
- 内存使用优化

## 错误处理

包提供结构化的错误处理，便于调试和监控。

## 集成建议

此包设计用于替代旧的 `pkg/llm`，提供更稳定和功能丰富的 LLM 提供商接入能力。