package llm2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kymo-mcp/mcpcan/pkg/llm2/adaptor"
	"github.com/kymo-mcp/mcpcan/pkg/llm2/adaptor/mcp"
	"github.com/kymo-mcp/mcpcan/pkg/llm2/adaptor/openai"
	"github.com/kymo-mcp/mcpcan/pkg/llm2/channeltype"
	"github.com/kymo-mcp/mcpcan/pkg/llm2/meta"
	"github.com/kymo-mcp/mcpcan/pkg/llm2/model"
)

// GetAdaptor 根据渠道类型获取适配器
func GetAdaptor(channelType int) adaptor.Adaptor {
	switch channelType {
	case channeltype.OpenAI:
		return &openai.Adaptor{}
	case channeltype.MCP: // 假设 MCP 有自己的类型
		return &mcp.Adaptor{}
	case channeltype.Azure:
		return &openai.Adaptor{}
	}
	return nil
}

// ProviderType 定义 LLM 提供商类型
type ProviderType string

const (
	// Core Providers (OpenAI-compatible)
	ProviderOpenAI      ProviderType = "openai"
	ProviderAzureOpenAI ProviderType = "azure_openai"
	ProviderDeepSeek    ProviderType = "deepseek"
	ProviderAnthropic   ProviderType = "anthropic"
	ProviderGoogle      ProviderType = "google"
	ProviderMistral     ProviderType = "mistral"
	ProviderXAI         ProviderType = "xai"
	ProviderMCP         ProviderType = "mcp" // 新增 MCP 类型

	// Aggregator/Proxy Providers
	ProviderOpenRouter ProviderType = "openrouter"
	ProviderLiteLLM    ProviderType = "litellm"
	ProviderOllama     ProviderType = "ollama"

	// Chinese Providers
	ProviderQwen     ProviderType = "qwen"     // Aliyun Tongyi Qwen
	ProviderDoubao   ProviderType = "doubao"   // Volcengine Doubao
	ProviderZhipu    ProviderType = "zhipu"    // Zhipu GLM
	ProviderMoonshot ProviderType = "moonshot" // Moonshot Kimi
)

// Provider 接口定义
type Provider interface {
	StreamChat(ctx context.Context, req ChatRequest) (<-chan StreamResponse, error)
}

// Message 代表一个聊天消息
type Message struct {
	Role         string               `json:"role"`
	Content      string               `json:"content"`                 // For backwards compatibility and simple text
	MultiContent []MessageContentPart `json:"multi_content,omitempty"` // For multimodal content
	ToolCalls    []ToolCall           `json:"tool_calls,omitempty"`
	ToolCallID   string               `json:"tool_call_id,omitempty"`
	ToolCallName string               `json:"tool_call_name,omitempty"` // 工具函数名，Google FunctionResponse 需要
}

// MessageContentPart represents a part of the message content (text or image)
type MessageContentPart struct {
	Type     string           `json:"type"` // text, image_url
	Text     string           `json:"text,omitempty"`
	ImageURL *MessageImageURL `json:"image_url,omitempty"`
}

// MessageImageURL represents image url
type MessageImageURL struct {
	URL string `json:"url"`
}

// ToolCall represents a tool call request from LLM
type ToolCall struct {
	Index    int              `json:"index,omitempty"` // Added for streaming aggregation
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction represents the function call details in a tool call (arguments are string JSON)
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Tool represents a tool definition
type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// Function represents the function call details
type Function struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters,omitempty"` // JSON Schema
}

// ChatRequest represents a request to the LLM
type ChatRequest struct {
	Model       string
	Messages    []Message
	Tools       []Tool // Structured tool definition
	Temperature float32
	MaxTokens   int
	Stream      bool
}

// ToolOutput represents the result of a tool execution
type ToolOutput struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Result string `json:"result"`
}

// StreamResponse represents a chunk of response from streaming API
type StreamResponse struct {
	Content     string
	ToolCalls   []ToolCall
	ToolOutputs []ToolOutput // Added for notifying execution results
	Usage       *Usage
	Error       error
}

// Usage represents token usage statistics
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// ProviderConfig holds configuration for LLM provider
type ProviderConfig struct {
	BaseURL string
	APIKey  string
}

// 适配器工厂
type adapterFactory struct{}

// NewProvider 创建新的提供商实例
func (f *adapterFactory) NewProvider(providerType ProviderType, config ProviderConfig) (Provider, error) {
	var channelType int

	// 映射 ProviderType 到 channeltype
	switch providerType {
	case ProviderOpenAI:
		channelType = channeltype.OpenAI
	case ProviderAzureOpenAI:
		channelType = channeltype.Azure
	case ProviderMCP:
		channelType = channeltype.Custom // 或定义 MCP 类型
	default:
		channelType = channeltype.OpenAI // 默认使用 OpenAI 类型
	}

	// 创建元数据
	meta := &meta.Meta{
		ChannelType:     channelType,
		BaseURL:         config.BaseURL,
		APIKey:          config.APIKey,
		OriginModelName: "gpt-3.5-turbo", // 示例默认模型
		ActualModelName: "gpt-3.5-turbo",
		IsStream:        true,
	}

	// 获取适配器
	adapter := GetAdaptor(channelType)
	if adapter == nil {
		return nil, nil
	}

	// 初始化适配器
	adapter.Init(meta)

	// 返回包装的提供商实现
	return &wrappedProvider{
		adapter: adapter,
		meta:    meta,
		config:  config,
	}, nil
}

// wrappedProvider 包装适配器为 Provider 接口
type wrappedProvider struct {
	adapter adaptor.Adaptor
	meta    *meta.Meta
	config  ProviderConfig
}

// StreamChat 实现 Provider 接口
func (wp *wrappedProvider) StreamChat(ctx context.Context, req ChatRequest) (<-chan StreamResponse, error) {
	// 将内部请求格式转换为 one-api 通用格式
	generalReq := convertToGeneralRequest(req)

	// 创建响应通道
	responseChan := make(chan StreamResponse)

	// 启动一个 goroutine 执行实际的请求和响应处理
	go func() {
		defer close(responseChan)

		// 转换请求
		convertedReq, err := wp.adapter.ConvertRequest(ctx, 0, generalReq)
		if err != nil {
			responseChan <- StreamResponse{
				Error: fmt.Errorf("failed to convert request: %w", err),
			}
			return
		}

		// 创建 JSON 请求体
		jsonData, err := json.Marshal(convertedReq)
		if err != nil {
			responseChan <- StreamResponse{
				Error: fmt.Errorf("failed to marshal request: %w", err),
			}
			return
		}

		// 获取请求 URL
		requestURL, err := wp.adapter.GetRequestURL(wp.meta)
		if err != nil {
			responseChan <- StreamResponse{
				Error: fmt.Errorf("failed to get request URL: %w", err),
			}
			return
		}

		// 创建 HTTP 请求
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewBuffer(jsonData))
		if err != nil {
			responseChan <- StreamResponse{
				Error: fmt.Errorf("failed to create HTTP request: %w", err),
			}
			return
		}

		// 设置请求头
		err = wp.adapter.SetupRequestHeader(ctx, httpReq, wp.meta)
		if err != nil {
			responseChan <- StreamResponse{
				Error: fmt.Errorf("failed to set request header: %w", err),
			}
			return
		}

		// 执行请求
		resp, err := wp.adapter.DoRequest(ctx, wp.meta, bytes.NewBuffer(jsonData))
		if err != nil {
			responseChan <- StreamResponse{
				Error: fmt.Errorf("failed to make request: %w", err),
			}
			return
		}
		defer resp.Body.Close()

		// 使用一个包装器来处理响应
		wrapper := &GinContextWrapper{}

		// 处理响应
		usage, errWithStatus := wp.adapter.DoResponse(ctx, wrapper, resp, wp.meta)
		if errWithStatus != nil {
			responseChan <- StreamResponse{
				Error: fmt.Errorf(errWithStatus.Error.Message),
			}
			return
		}

		// 如果有用量信息，发送
		if usage != nil {
			responseChan <- StreamResponse{
				Usage: &Usage{
					PromptTokens:     usage.PromptTokens,
					CompletionTokens: usage.CompletionTokens,
					TotalTokens:      usage.TotalTokens,
				},
			}
		}
	}()

	return responseChan, nil
}

// GinContextWrapper 适配 adaptor 接口要求的 GinContext
type GinContextWrapper struct{}

// 实现 adaptor.GinContext 接口的方法
// 由于 adaptor.Interface 中的 DoResponse 期望的是 gin.Context 类型，
// 我们创建一个基本实现来满足接口要求

// convertToGeneralRequest 将内部请求转换为通用请求格式
func convertToGeneralRequest(internalReq ChatRequest) *model.GeneralOpenAIRequest {
	// 转换消息格式
	var messages []model.Message
	for _, msg := range internalReq.Messages {
		message := model.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
		messages = append(messages, message)
	}

	// 转换工具格式
	var tools []model.Tool
	for _, tool := range internalReq.Tools {
		modelTool := model.Tool{
			Type: tool.Type,
			Function: model.Function{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  tool.Function.Parameters,
			},
		}
		tools = append(tools, modelTool)
	}

	tempFloat64 := float64(internalReq.Temperature)
	return &model.GeneralOpenAIRequest{
		Messages:    messages,
		Model:       internalReq.Model,
		Tools:       tools,
		Temperature: &tempFloat64,
		MaxTokens:   internalReq.MaxTokens,
		Stream:      internalReq.Stream,
	}
}

// 全局工厂实例
var globalAdapterFactory = &adapterFactory{}

// NewProvider 创建新的提供商
func NewProvider(providerType ProviderType, config ProviderConfig) (Provider, error) {
	return globalAdapterFactory.NewProvider(providerType, config)
}
