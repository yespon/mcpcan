# 从原始 llm 包迁移到 llm2 包的完整指南

## 概述

本指南详细说明如何将系统从原始的 `pkg/llm` 包迁移到基于 one-api relay 模式的新 `pkg/llm2` 包。这将解决原始包中的数据流解析错误，并提供更好的 MCP 工具调试支持。

## 迁移的好处

1. **更稳定的数据流解析** - 基于 one-api relay 模式的稳健流处理
2. **改进的 MCP 工具支持** - 优化的工具调用和调试功能  
3. **更好的错误处理** - 更详细和准确的错误反馈
4. **one-api 风格架构** - 借鉴业界最佳实践
5. **更好的性能** - 优化的请求和响应处理

## 迁移步骤

### 第一步：理解接口兼容性

`llm2` 包设计为与原始 `llm` 包保持接口兼容，因此大部分代码无需修改：

| 原始包 (`pkg/llm`) | 新包 (`pkg/llm2`) |
|-------------------|------------------|
| `llm.Provider` | `llm2.Provider` |
| `llm.ChatRequest` | `llm2.ChatRequest` |
| `llm.StreamResponse` | `llm2.StreamResponse` |
| `llm.Message` | `llm2.Message` |
| `llm.Tool` | `llm2.Tool` |

### 第二步：修改导入语句

修改 `internal/market/biz/ai_session.go` 文件中的导入：

```go
// 原来的导入
import (
    // ...
    "github.com/kymo-mcp/mcpcan/pkg/llm"
    // ...
)

// 修改为
import (
    // ...
    "github.com/kymo-mcp/mcpcan/pkg/llm2"
    // ...
)
```

### 第三步：更新提供商创建代码

将提供商初始化部分替换：

```go
// 原来的代码
provider, err := llm.NewProvider(providerType, llm.ProviderConfig{
    BaseURL: modelAccess.BaseUrl,
    APIKey:  modelAccess.ApiKey,
})
if err != nil {
    return nil, fmt.Errorf("failed to init provider: %s", err.Error())
}

// 修改为
provider, err := llm2.NewProvider(providerType, llm2.ProviderConfig{
    BaseURL: modelAccess.BaseUrl,
    APIKey:  modelAccess.ApiKey,
})
if err != nil {
    return nil, fmt.Errorf("failed to init provider: %s", err.Error())
}
```

### 第四步：更新 StreamChat 调用

```go
// 原来的方法调用保持不变，因为接口兼容
stream, err := provider.StreamChat(ctx, reqChat)
```

### 第五步：验证类型转换

确保消息、工具等类型的使用方式保持不变，因为接口完全兼容。

## 实施适配器模式（推荐方法）

为了最小化风险，可以使用适配器模式逐步迁移：

1. 创建适配器包 `pkg/llm_adapter`（已创建）
2. 使用适配器包作为中间层
3. 逐步验证功能正常后再完全切换

## 代码修改示例

### 原始 ai_session.go 的修改

在 `backend/internal/market/biz/ai_session.go` 中：

```go
// 文件顶部导入修改
import (
    // 原有导入...
    "github.com/kymo-mcp/mcpcan/pkg/llm2"  // 替换原来的 "github.com/kymo-mcp/mcpcan/pkg/llm"
    // 其他导入...
)

// 在 Chat 方法中修改提供商创建
func (b *AiSessionBiz) Chat(ctx context.Context, req *pb.ChatRequest) (<-chan llm2.StreamResponse, error) {
    // ... 其他代码保持不变 ...
    
    // 使用 llm2 替换原来 llm 的调用
    provider, err := llm2.NewProvider(providerType, llm2.ProviderConfig{
        BaseURL: modelAccess.BaseUrl,
        APIKey:  modelAccess.ApiKey,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to init provider: %s", err.Error())
    }
    
    // StreamChat 调用不变，因为接口兼容
    stream, err := provider.StreamChat(ctx, reqChat)
    // ... 其余代码不变 ...
}
```

## 验证步骤

1. **单元测试** - 运行所有相关单元测试
2. **集成测试** - 测试聊天功能是否正常工作
3. **MCP 工具测试** - 验证 MCP 工具调用是否正常
4. **错误处理测试** - 确认错误处理逻辑正确
5. **性能测试** - 确认性能没有显著下降

## 已知差异和注意事项

1. **流式处理** - llm2 使用更稳健的流处理机制，错误率更低
2. **MCP 工具** - llm2 提供更好的 MCP 工具集成和调试支持
3. **错误信息** - llm2 提供更详细的错误信息，便于调试
4. **性能** - llm2 有更好的性能表现

## 回滚策略

如果迁移过程中出现问题：

1. 将导入语句改回 `pkg/llm`
2. 恢复提供商初始化代码
3. 恢复相关类型和方法调用

## 总结

通过使用基于 one-api relay 模式的 llm2 包，系统将获得：
- 更稳定的数据流处理
- 更好的错误恢复机制  
- 优化的 MCP 工具支持
- 更高的性能和可靠性

迁移过程简单，主要是导入语句的修改，其余代码可以保持不变。