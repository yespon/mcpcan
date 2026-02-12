package mcp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kymo-mcp/mcpcan/pkg/llm2/adaptor"
	"github.com/kymo-mcp/mcpcan/pkg/llm2/meta"
	"github.com/kymo-mcp/mcpcan/pkg/llm2/model"
	"github.com/kymo-mcp/mcpcan/pkg/llm2/relaymode"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type Adaptor struct {
	client *client.Client
}

func (a *Adaptor) Init(meta *meta.Meta) {
	// 初始化 MCP 客户端
	// 这里可以使用 meta 中的信息来初始化
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	// MCP 使用不同的通信方式，不一定需要 URL
	// 返回基础 URL 或使用其他标识
	if meta.BaseURL != "" {
		return meta.BaseURL, nil
	}
	return "", fmt.Errorf("MCP base URL not provided")
}

func (a *Adaptor) SetupRequestHeader(ctx context.Context, req *http.Request, meta *meta.Meta) error {
	// MCP 可能使用不同的认证机制
	req.Header.Set("Authorization", "Bearer "+meta.APIKey)
	return nil
}

func (a *Adaptor) ConvertRequest(ctx context.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, fmt.Errorf("request is nil")
	}

	// 对于 MCP，需要将 OpenAI 格式的请求转换为 MCP 格式
	// 这里进行格式转换
	switch relayMode {
	case relaymode.ChatCompletions:
		// 转换聊天请求
		return a.convertChatRequest(request)
	default:
		// 其他类型请求转换
		return request, nil
	}
}

// convertChatRequest 将 OpenAI 格式聊天请求转换为 MCP 格式
func (a *Adaptor) convertChatRequest(request *model.GeneralOpenAIRequest) (*mcp.CallToolRequest, error) {
	// 提取系统提示、用户消息等
	var systemPrompt, userContent string

	for _, msg := range request.Messages {
		if msg.Role == "system" {
			systemPrompt = msg.StringContent()
		} else if msg.Role == "user" {
			userContent = msg.StringContent()
		}
	}

	if userContent == "" {
		return nil, fmt.Errorf("no user content found in messages")
	}

	// 构建 MCP 工具调用请求
	mcpRequest := &mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "chat", // 默认使用聊天工具
			Arguments: map[string]interface{}{
				"system_prompt": systemPrompt,
				"user_input":    userContent,
				"model":         request.Model,
			},
		},
	}

	return mcpRequest, nil
}

func (a *Adaptor) ConvertImageRequest(request *model.ImageRequest) (any, error) {
	if request == nil {
		return nil, fmt.Errorf("request is nil")
	}
	return request, nil
}

func (a *Adaptor) DoRequest(ctx context.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	// 对于 MCP，我们不使用 HTTP 请求，而是使用 MCP 协议
	// 这里模拟 MCP 调用过程
	var mcpClient *client.Client
	var err error

	// 根据配置确定 MCP 连接方式
	if strings.HasPrefix(meta.BaseURL, "http") {
		// SSE 模式
		mcpClient, err = client.NewSSEMCPClient(meta.BaseURL)
	} else {
		// TODO: 添加其他 MCP 连接方式支持 (stdio, streamable-http)
		return nil, fmt.Errorf("unsupported MCP connection type for: %s", meta.BaseURL)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create MCP client: %w", err)
	}

	// 初始化 MCP 客户端
	err = mcpClient.Start(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start MCP client: %w", err)
	}

	_, err = mcpClient.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcp.Implementation{
				Name:    "mcpcan-llm2-mcp",
				Version: "1.0.0",
			},
			Capabilities: mcp.ClientCapabilities{},
		},
	})
	if err != nil {
		mcpClient.Close()
		return nil, fmt.Errorf("failed to initialize MCP client: %w", err)
	}

	// 列出可用工具
	_, err = mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		mcpClient.Close()
		return nil, fmt.Errorf("failed to list MCP tools: %w", err)
	}

	// 这里应该实际处理请求，但现在我们返回一个模拟的 HTTP 响应
	mcpClient.Close()

	// 创建模拟响应
	dummyResp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Proto:      "HTTP/1.1",
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(`{"choices":[{"message":{"content":"MCP response"}}]}`)),
	}

	return dummyResp, nil
}

func (a *Adaptor) DoResponse(ctx context.Context, c adaptor.GinContext, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	// MCP 特定的响应处理
	if meta.IsStream {
		// 流式响应处理
		return a.handleStreamResponse(c, resp, meta.Mode)
	} else {
		// 非流式响应处理
		return a.handleTextResponse(c, resp, meta.PromptTokens, meta.ActualModelName)
	}
}

func (a *Adaptor) handleStreamResponse(c adaptor.GinContext, resp *http.Response, relayMode int) (*model.Usage, *model.ErrorWithStatusCode) {
	// MCP 流式响应处理
	return nil, nil
}

func (a *Adaptor) handleTextResponse(c adaptor.GinContext, resp *http.Response, promptTokens int, modelName string) (*model.Usage, *model.ErrorWithStatusCode) {
	// MCP 文本响应处理
	return nil, nil
}

func (a *Adaptor) GetModelList() []string {
	// 返回 MCP 支持的工具作为模型列表
	return []string{
		"mcp-tool-chat",
		"mcp-tool-web-search",
		"mcp-tool-code-executor",
		"mcp-tool-file-manager",
	}
}

func (a *Adaptor) GetChannelName() string {
	return "mcp"
}
