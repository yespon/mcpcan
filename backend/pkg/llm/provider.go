package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/tmc/langchaingo/llms"
)

// LangChainAdapter wraps a LangChainGo model to implement the internal Provider interface
type LangChainAdapter struct {
	model llms.Model
}

// StreamChat implements the Provider interface using LangChainGo
func (p *LangChainAdapter) StreamChat(ctx context.Context, req ChatRequest) (<-chan StreamResponse, error) {
	responseChan := make(chan StreamResponse)

	go func() {
		defer close(responseChan)

		// Convert internal messages to LangChainGo messages
		content := make([]llms.MessageContent, 0, len(req.Messages))
		for _, msg := range req.Messages {
			role := llms.ChatMessageTypeHuman
			switch msg.Role {
			case "system":
				role = llms.ChatMessageTypeSystem
			case "assistant":
				role = llms.ChatMessageTypeAI
			case "function":
				role = llms.ChatMessageTypeFunction
			case "tool":
				role = llms.ChatMessageTypeTool
			}

			// Handle multi-content (text + images)
			parts := []llms.ContentPart{}

			if msg.Role == "tool" {
				// Tool role 消息必须使用 ToolCallResponse 类型
				// langchaingo 的 OpenAI/Google provider 都要求此格式
				// Google FunctionResponse.Name 必须是函数名（非 call ID）
				parts = append(parts, llms.ToolCallResponse{
					ToolCallID: msg.ToolCallID,
					Name:       msg.ToolCallName,
					Content:    msg.Content,
				})
			} else if len(msg.MultiContent) > 0 {
				for _, part := range msg.MultiContent {
					if part.Type == "text" {
						parts = append(parts, llms.TextPart(part.Text))
					} else if part.Type == "image_url" && part.ImageURL != nil {
						parts = append(parts, llms.ImageURLPart(part.ImageURL.URL))
					}
				}
			} else {
				// Simple text content
				parts = append(parts, llms.TextPart(msg.Content))
			}

			// Assistant 消息如果包含 ToolCalls，需要附加 ToolCall parts
			if msg.Role == "assistant" && len(msg.ToolCalls) > 0 {
				for _, tc := range msg.ToolCalls {
					parts = append(parts, llms.ToolCall{
						ID:   tc.ID,
						Type: tc.Type,
						FunctionCall: &llms.FunctionCall{
							Name:      tc.Function.Name,
							Arguments: tc.Function.Arguments,
						},
					})
				}
			}

			// Google Gemini API 要求并行 Function Calling 的结果必须在一条 user 消息中返回
			// 但 langchaingo 库的 googleai 实现可能严格校验了 Tool 消息只能包含一个 Part
			// 报错：[System Error: expected exactly one part for role tool, got 2]
			// 因此暂时禁用合并逻辑，让 langchaingo 自行处理（或 API 支持多条 Tool 消息）
			/*
			if role == llms.ChatMessageTypeTool && len(content) > 0 {
				lastIdx := len(content) - 1
				if content[lastIdx].Role == llms.ChatMessageTypeTool {
					content[lastIdx].Parts = append(content[lastIdx].Parts, parts...)
					continue
				}
			}
			*/

			content = append(content, llms.MessageContent{
				Role:  role,
				Parts: parts,
			})
		}

		// Configure generation options
		opts := []llms.CallOption{
			llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
				select {
				case responseChan <- StreamResponse{
					Content: string(chunk),
				}:
				case <-ctx.Done():
					return ctx.Err()
				}
				return nil
			}),
		}

		if req.Temperature > 0 {
			opts = append(opts, llms.WithTemperature(float64(req.Temperature)))
		}
		if req.MaxTokens > 0 {
			opts = append(opts, llms.WithMaxTokens(req.MaxTokens))
		}

		// Set model if specified (some providers support model switching per request)
		// OpenAI supports WithModel, others might ignore or support it.
		if req.Model != "" {
			opts = append(opts, llms.WithModel(req.Model))
		}

		// Convert tools if present
		if len(req.Tools) > 0 {
			var llmTools []llms.Tool
			for _, t := range req.Tools {
				var params interface{} = t.Function.Parameters
				// If params is json.RawMessage, unmarshal it to map
				if raw, ok := params.(json.RawMessage); ok {
					var p map[string]interface{}
					if err := json.Unmarshal(raw, &p); err == nil {
						params = p
					}
				}

				log.Printf("[Provider Debug] Tool: name=%s, params_type=%T, params=%+v", t.Function.Name, params, params)

				llmTools = append(llmTools, llms.Tool{
					Type: t.Type,
					Function: &llms.FunctionDefinition{
						Name:        t.Function.Name,
						Description: t.Function.Description,
						Parameters:  params,
					},
				})
			}
			opts = append(opts, llms.WithTools(llmTools))
		}

		// Debug: 打印发送给 LLM 的消息结构
		if contentJson, err := json.MarshalIndent(content, "", "  "); err == nil {
			log.Printf("[Provider Debug] Full Content Payload:\n%s", string(contentJson))
		} else {
			log.Printf("[Provider Error] Failed to marshal content: %v", err)
		}

		for i, mc := range content {
			partTypes := make([]string, 0, len(mc.Parts))
			for _, p := range mc.Parts {
				partTypes = append(partTypes, fmt.Sprintf("%T", p))
			}
			log.Printf("[Provider Debug] Message[%d]: Role=%s, Parts=%v", i, mc.Role, partTypes)
		}

		// Call LangChainGo GenerateContent
		result, err := p.model.GenerateContent(ctx, content, opts...)
		if err != nil {
			log.Printf("[Provider Error] GenerateContent failed (full): %+v", err)
			log.Printf("[Provider Error] GenerateContent failed (string): %s", err.Error())
			select {
			case responseChan <- StreamResponse{Error: err}:
			case <-ctx.Done():
			}
			return
		}

		// Process the response and handle tool calls, content, and usage information
		if len(result.Choices) > 0 {
			choice := result.Choices[0]

			// Handle tool calls if present
			if len(choice.ToolCalls) > 0 {
				var toolCalls []ToolCall
				for i, tc := range choice.ToolCalls {
					// Google provider 不返回 ToolCall ID，我们需要生成一个或者依赖 Index
					// 这里的关键是必须设置 Index，否则 ai_session 会把所有 tool call 堆积到 Index 0
					toolID := tc.ID
					if toolID == "" {
						toolID = fmt.Sprintf("call_%d_%d", time.Now().UnixNano(), i)
					}

					toolCall := ToolCall{
						Index: i,
						ID:    toolID,
						Type:  tc.Type,
						Function: ToolCallFunction{
							Name:      tc.FunctionCall.Name,
							Arguments: tc.FunctionCall.Arguments,
						},
					}
					log.Printf("[Provider Debug] Received ToolCall: Name=%s, Args=%s", tc.FunctionCall.Name, tc.FunctionCall.Arguments)
					toolCalls = append(toolCalls, toolCall)
				}

				// Send tool calls through the response channel
				select {
				case responseChan <- StreamResponse{
					ToolCalls: toolCalls,
				}:
				case <-ctx.Done():
				}
			}

			// Send usage information if available
			if choice.GenerationInfo != nil {
				// Extract token usage if available in GenerationInfo
				var usage *Usage

				if promptTokens, ok := choice.GenerationInfo["prompt_tokens"].(float64); ok {
					if completionTokens, ok := choice.GenerationInfo["completion_tokens"].(float64); ok {
						if totalTokens, ok := choice.GenerationInfo["total_tokens"].(float64); ok {
							usage = &Usage{
								PromptTokens:     int(promptTokens),
								CompletionTokens: int(completionTokens),
								TotalTokens:      int(totalTokens),
							}

							select {
							case responseChan <- StreamResponse{
								Usage: usage,
							}:
							case <-ctx.Done():
							}
						}
					}
					// Alternative key names for token counts used by different providers
				} else if usageMetadata, ok := choice.GenerationInfo["usage_metadata"].(map[string]interface{}); ok {
					if inputTokens, ok := usageMetadata["input_tokens"]; ok {
						if outputTokens, ok := usageMetadata["output_tokens"]; ok {
							var promptTok, completionTok int
							if pt, ok := inputTokens.(float64); ok {
								promptTok = int(pt)
							} else if pt, ok := inputTokens.(int); ok {
								promptTok = pt
							}
							if ct, ok := outputTokens.(float64); ok {
								completionTok = int(ct)
							} else if ct, ok := outputTokens.(int); ok {
								completionTok = ct
							}

							usage = &Usage{
								PromptTokens:     promptTok,
								CompletionTokens: completionTok,
								TotalTokens:      promptTok + completionTok,
							}

							select {
							case responseChan <- StreamResponse{
								Usage: usage,
							}:
							case <-ctx.Done():
							}
						}
					}
				}
			}
		}
	}()

	return responseChan, nil
}
