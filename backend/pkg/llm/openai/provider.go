package openai

import (
	"context"
	"errors"
	"io"

	"github.com/kymo-mcp/mcpcan/pkg/llm"
	openai "github.com/sashabaranov/go-openai"
)

func init() {
	factory := func(config llm.ProviderConfig) llm.Provider {
		return NewProvider(config)
	}
	llm.RegisterProvider(llm.ProviderOpenAI, factory)
	llm.RegisterProvider(llm.ProviderDeepSeek, factory)
	llm.RegisterProvider(llm.ProviderMoonshot, factory)
	llm.RegisterProvider(llm.ProviderDoubao, factory)
}

type Provider struct {
	client *openai.Client
}

// NewProvider creates a new OpenAI provider
func NewProvider(config llm.ProviderConfig) *Provider {
	openaiConfig := openai.DefaultConfig(config.APIKey)
	if config.BaseURL != "" {
		openaiConfig.BaseURL = config.BaseURL
	}
	return &Provider{
		client: openai.NewClientWithConfig(openaiConfig),
	}
}

// StreamChat implements llm.Provider interface
func (p *Provider) StreamChat(ctx context.Context, req llm.ChatRequest) (<-chan llm.StreamResponse, error) {
	// Convert messages
	openaiMessages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		openaiMsg := openai.ChatCompletionMessage{
			Role:       msg.Role,
			Content:    msg.Content,
			ToolCallID: msg.ToolCallID,
		}

		if len(msg.MultiContent) > 0 {
			parts := make([]openai.ChatMessagePart, len(msg.MultiContent))
			for j, p := range msg.MultiContent {
				part := openai.ChatMessagePart{
					Type: openai.ChatMessagePartType(p.Type),
					Text: p.Text,
				}
				if p.ImageURL != nil {
					part.ImageURL = &openai.ChatMessageImageURL{
						URL: p.ImageURL.URL,
					}
				}
				parts[j] = part
			}
			openaiMsg.MultiContent = parts
			openaiMsg.Content = "" // Clear simple content if multi-content exists
		}

		if len(msg.ToolCalls) > 0 {
			toolCalls := make([]openai.ToolCall, len(msg.ToolCalls))
			for j, tc := range msg.ToolCalls {
				toolCalls[j] = openai.ToolCall{
					ID:   tc.ID,
					Type: openai.ToolType(tc.Type),
					Function: openai.FunctionCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			}
			openaiMsg.ToolCalls = toolCalls
		}
		openaiMessages[i] = openaiMsg
	}

	// Convert Tools
	var tools []openai.Tool
	if len(req.Tools) > 0 {
		tools = make([]openai.Tool, len(req.Tools))
		for i, tool := range req.Tools {
			tools[i] = openai.Tool{
				Type: openai.ToolType(tool.Type),
				Function: &openai.FunctionDefinition{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					Parameters:  tool.Function.Parameters,
				},
			}
		}
	}

	chatReq := openai.ChatCompletionRequest{
		Model:    req.Model,
		Messages: openaiMessages,
		Tools:    tools,
		Stream:   true,
		StreamOptions: &openai.StreamOptions{
			IncludeUsage: true,
		},
	}
	if req.MaxTokens > 0 {
		chatReq.MaxTokens = req.MaxTokens
	}
	if req.Temperature > 0 {
		chatReq.Temperature = req.Temperature
	}

	stream, err := p.client.CreateChatCompletionStream(ctx, chatReq)
	if err != nil {
		return nil, err
	}

	responseChan := make(chan llm.StreamResponse)

	go func() {
		defer close(responseChan)
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				responseChan <- llm.StreamResponse{Error: err}
				return
			}

			resp := llm.StreamResponse{}

			// Handle Usage
			if response.Usage != nil {
				resp.Usage = &llm.Usage{
					PromptTokens:     response.Usage.PromptTokens,
					CompletionTokens: response.Usage.CompletionTokens,
					TotalTokens:      response.Usage.TotalTokens,
				}
			}

			// Handle Choices
			if len(response.Choices) > 0 {
				choice := response.Choices[0]
				resp.Content = choice.Delta.Content

				// Handle ToolCalls (Streaming)
				// Note: OpenAI streams tool calls in parts. This logic simplifies it for now.
				// A robust implementation needs to accumulate tool calls.
				// For this iteration, we focus on text content mostly as per Phase 1/2 goals.
				if len(choice.Delta.ToolCalls) > 0 {
					for _, tc := range choice.Delta.ToolCalls {
						idx := 0
						if tc.Index != nil {
							idx = *tc.Index
						}
						toolCall := llm.ToolCall{
							Index: idx,
							ID:    tc.ID,
							Type:  string(tc.Type),
							Function: llm.ToolCallFunction{
								Name:      tc.Function.Name,
								Arguments: tc.Function.Arguments,
							},
						}
						resp.ToolCalls = append(resp.ToolCalls, toolCall)
					}
				}
			}

			responseChan <- resp
		}
	}()

	return responseChan, nil
}
