package doubao

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kymo-mcp/mcpcan/pkg/llm"
)

func init() {
	factory := func(config llm.ProviderConfig) llm.Provider {
		return NewProvider(config)
	}
	llm.RegisterProvider(llm.ProviderDoubao, factory)
}

type Provider struct {
	config llm.ProviderConfig
	client *http.Client
}

func NewProvider(config llm.ProviderConfig) *Provider {
	return &Provider{
		config: config,
		client: &http.Client{},
	}
}

func (p *Provider) StreamChat(ctx context.Context, req llm.ChatRequest) (<-chan llm.StreamResponse, error) {
	// 1. Prepare URL
	url := p.config.BaseURL
	if !strings.HasSuffix(url, "/chat/completions") {
		// If user didn't provide full path, assume typical structure
		// Removing trailing slash
		url = strings.TrimSuffix(url, "/")
		if !strings.HasSuffix(url, "/api/v3") {
			// Volcengine supports compatible /v1 too, but let's stick to what we received or generic append
			// If it's a raw host, append /api/v3/chat/completions?
			// Best practice: Trust generic structure unless specific known host?
			// Let's assume user provides base like 'https://ark.cn-beijing.volces.com/api/v3'
			url = url + "/chat/completions"
		} else {
			url = url + "/chat/completions"
		}
	}

	// 2. Prepare Request Body (OpenAI Compatible)
	openAIReq := map[string]interface{}{
		"model":       req.Model,
		"messages":    req.Messages,
		"stream":      true,
		"temperature": req.Temperature,
	}
    // Map messages specifically if needed, but struct should json marshal generally ok
    // Actually req.Messages has custom structs, better to manual map if they differ
    // For now assuming compatible JSON

    if req.MaxTokens > 0 {
        openAIReq["max_tokens"] = req.MaxTokens
    }
    // Tools
    if len(req.Tools) > 0 {
        openAIReq["tools"] = req.Tools
    }

	jsonData, err := json.Marshal(openAIReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %v", err)
	}

	// Debug Log
	fmt.Printf("[Doubao Debug] Request URL: %s\n", url)
	keyLen := len(p.config.APIKey)
	if keyLen > 4 {
		fmt.Printf("[Doubao Debug] API Key: %s... (%d chars)\n", p.config.APIKey[:4], keyLen)
	} else {
		fmt.Printf("[Doubao Debug] API Key: (too short) %s\n", p.config.APIKey)
	}

	// 3. Create HTTP Request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	// 4. Do Request
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	// 5. Stream Response
	responseChan := make(chan llm.StreamResponse)

	go func() {
		defer close(responseChan)
		defer resp.Body.Close()

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					responseChan <- llm.StreamResponse{Error: err}
				}
				return
			}

			line = bytes.TrimSpace(line)
			if !bytes.HasPrefix(line, []byte("data: ")) {
				continue
			}

			data := bytes.TrimPrefix(line, []byte("data: "))
			if string(data) == "[DONE]" {
				return
			}

			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
                        ToolCalls []struct {
                            Index int `json:"index"`
                            ID string `json:"id"`
                            Type string `json:"type"`
                            Function struct {
                                Name string `json:"name"`
                                Arguments string `json:"arguments"`
                            } `json:"function"`
                        } `json:"tool_calls"`
					} `json:"delta"`
				} `json:"choices"`
                Usage *struct {
                    PromptTokens int `json:"prompt_tokens"`
                    CompletionTokens int `json:"completion_tokens"`
                    TotalTokens int `json:"total_tokens"`
                } `json:"usage"`
			}

			if err := json.Unmarshal(data, &chunk); err != nil {
				continue
			}

			var streamResp llm.StreamResponse
			if len(chunk.Choices) > 0 {
				streamResp.Content = chunk.Choices[0].Delta.Content
                if len(chunk.Choices[0].Delta.ToolCalls) > 0 {
                    // Map tool calls
                    for _, tc := range chunk.Choices[0].Delta.ToolCalls {
                        streamResp.ToolCalls = append(streamResp.ToolCalls, llm.ToolCall{
                            Index: tc.Index,
                            ID: tc.ID,
                            Type: tc.Type,
                            Function: llm.ToolCallFunction{
                                Name: tc.Function.Name,
                                Arguments: tc.Function.Arguments,
                            },
                        })
                    }
                }
			}
            if chunk.Usage != nil {
                streamResp.Usage = &llm.Usage{
                    PromptTokens: chunk.Usage.PromptTokens,
                    CompletionTokens: chunk.Usage.CompletionTokens,
                    TotalTokens: chunk.Usage.TotalTokens,
                }
            }
			responseChan <- streamResp
		}
	}()

	return responseChan, nil
}
