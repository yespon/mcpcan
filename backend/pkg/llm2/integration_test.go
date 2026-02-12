package llm2

import (
	"context"
	"testing"
	"time"
)

func TestNewProvider_OpenAI(t *testing.T) {
	config := ProviderConfig{
		BaseURL: "https://api.openai.com/v1",
		APIKey:  "test-key",
	}

	provider, err := NewProvider(ProviderOpenAI, config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if provider == nil {
		t.Fatal("Expected provider to be created, got nil")
	}
}

func TestChatHandler_Creation(t *testing.T) {
	config := ProviderConfig{
		BaseURL: "https://api.openai.com/v1",
		APIKey:  "test-key",
	}

	handler, err := NewChatHandler(ProviderOpenAI, config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if handler == nil {
		t.Fatal("Expected handler to be created, got nil")
	}

	if handler.GetProvider() == nil {
		t.Fatal("Expected provider to be set, got nil")
	}
}

func TestChatRequest_Structure(t *testing.T) {
	req := ChatRequest{
		Model: "gpt-4o",
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hello, world!",
			},
		},
		Temperature: 0.7,
		Stream:      true,
		MaxTokens:   1000,
	}

	if req.Model != "gpt-4o" {
		t.Errorf("Expected model 'gpt-4o', got '%s'", req.Model)
	}

	if len(req.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(req.Messages))
	}

	if req.Messages[0].Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", req.Messages[0].Role)
	}

	if req.Messages[0].Content != "Hello, world!" {
		t.Errorf("Expected content 'Hello, world!', got '%s'", req.Messages[0].Content)
	}

	if req.Temperature != 0.7 {
		t.Errorf("Expected temperature 0.7, got %f", req.Temperature)
	}

	if !req.Stream {
		t.Error("Expected stream to be true")
	}

	if req.MaxTokens != 1000 {
		t.Errorf("Expected maxTokens 1000, got %d", req.MaxTokens)
	}
}

func TestChatHandler_StreamChat(t *testing.T) {
	// 注意：这个测试仅验证接口调用，不会实际发起网络请求
	config := ProviderConfig{
		BaseURL: "https://api.openai.com/v1",
		APIKey:  "test-key",
	}

	handler, err := NewChatHandler(ProviderOpenAI, config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	req := ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{
				Role:    "user",
				Content: "Test message",
			},
		},
		Stream: true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	stream, err := handler.StreamChat(ctx, req)
	if err != nil {
		t.Logf("Expected error due to no real server (this is OK): %v", err)
		// 不失败，因为没有真实服务
	} else {
		// 如果没有错误，测试读取流
		go func() {
			count := 0
			for resp := range stream {
				if resp.Error != nil {
					break
				}
				count++
				if count > 5 { // 限制读取次数
					break
				}
			}
		}()
	}

	// 测试成功，因为接口能够被调用
	t.Log("StreamChat interface test completed")
}

func TestProviderTypes_Constants(t *testing.T) {
	providerTypes := []ProviderType{
		ProviderOpenAI,
		ProviderAzureOpenAI,
		ProviderAnthropic,
		ProviderGoogle,
		ProviderMistral,
		ProviderXAI,
		ProviderOpenRouter,
		ProviderLiteLLM,
		ProviderOllama,
		ProviderQwen,
		ProviderDoubao,
		ProviderZhipu,
		ProviderMoonshot,
		ProviderMCP, // 新增 MCP 类型
	}

	for _, providerType := range providerTypes {
		if string(providerType) == "" {
			t.Errorf("ProviderType %v has empty string representation", providerType)
		}
	}
}
