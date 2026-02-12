package llm2

import (
	"context"
)

// LLMAdapter 将 llm2 功能适配到原有的 llm 接口头
type LLMAdapter struct {
	provider Provider
}

// NewLLMAdapter 创建新的适配器
func NewLLMAdapter(provider Provider) *LLMAdapter {
	return &LLMAdapter{
		provider: provider,
	}
}

// StreamChat 实现 llm 接口的 StreamChat 方法
func (a *LLMAdapter) StreamChat(ctx context.Context, req ChatRequest) (<-chan StreamResponse, error) {
	return a.provider.StreamChat(ctx, req)
}

// 实现 LLMProvider 接口的类型转换
type LLMProvider interface {
	StreamChat(ctx context.Context, req ChatRequest) (<-chan StreamResponse, error)
}

// 适配器函数，用于将 llm2 提供商适配到 llm 接口
func AdaptLLM2ToLLM(provider Provider) LLMProvider {
	return &LLMAdapter{
		provider: provider,
	}
}
