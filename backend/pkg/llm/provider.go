package llm

import (
	"context"
	"log"

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
			if len(msg.MultiContent) > 0 {
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
				llmTools = append(llmTools, llms.Tool{
					Type: t.Type,
					Function: &llms.FunctionDefinition{
						Name:        t.Function.Name,
						Description: t.Function.Description,
						Parameters:  t.Function.Parameters,
					},
				})
			}
			opts = append(opts, llms.WithTools(llmTools))
		}

		// Call LangChainGo GenerateContent
		result, err := p.model.GenerateContent(ctx, content, opts...)
		if err != nil {
			log.Printf("[Provider Error] GenerateContent failed: %v", err)
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
				for _, tc := range choice.ToolCalls {
					toolCall := ToolCall{
						ID:   tc.ID,
						Type: tc.Type,
						Function: ToolCallFunction{
							Name:      tc.Function.Name,
							Arguments: tc.Function.Arguments,
						},
					}
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
