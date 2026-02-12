package llm2

import (
	"context"
	"fmt"
)

// ChatHandler 高级聊天处理器，整合 one-api relay 模式的功能
type ChatHandler struct {
	provider Provider
}

// NewChatHandler 创建新的聊天处理器
func NewChatHandler(providerType ProviderType, config ProviderConfig) (*ChatHandler, error) {
	provider, err := NewProvider(providerType, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider for type %s: %w", providerType, err)
	}

	return &ChatHandler{
		provider: provider,
	}, nil
}

// StreamChat 执行流式聊天
func (h *ChatHandler) StreamChat(ctx context.Context, req ChatRequest) (<-chan StreamResponse, error) {
	return h.provider.StreamChat(ctx, req)
}

// SimpleChat 执行简单的非流式聊天，返回完整结果
func (h *ChatHandler) SimpleChat(ctx context.Context, req ChatRequest) (string, *Usage, error) {
	stream, err := h.StreamChat(ctx, req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to start chat stream: %w", err)
	}

	var fullContent string
	var usage *Usage

	for resp := range stream {
		if resp.Error != nil {
			return "", nil, fmt.Errorf("error during chat: %w", resp.Error)
		}

		if resp.Content != "" {
			fullContent += resp.Content
		}

		if resp.Usage != nil {
			usage = resp.Usage
		}
	}

	return fullContent, usage, nil
}

// WithProvider 使用已存在的提供商创建聊天处理器
func (h *ChatHandler) WithProvider(provider Provider) *ChatHandler {
	h.provider = provider
	return h
}

// GetProvider 获取内部提供商
func (h *ChatHandler) GetProvider() Provider {
	return h.provider
}
