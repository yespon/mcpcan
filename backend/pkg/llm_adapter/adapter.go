package llm_adapter

import (
	"context"

	orig_llm "github.com/kymo-mcp/mcpcan/pkg/llm"
)

// NewProvider creates a new provider using the default underlying implementation (pkg/llm)
func NewProvider(typ ProviderType, config ProviderConfig) (Provider, error) {
	// Convert config
	pConfig := orig_llm.ProviderConfig{
		BaseURL: config.BaseURL,
		APIKey:  config.APIKey,
	}

	// Create original provider
	p, err := orig_llm.NewProvider(orig_llm.ProviderType(typ), pConfig)
	if err != nil {
		return nil, err
	}

	return &adapterImpl{internal: p}, nil
}

type adapterImpl struct {
	internal orig_llm.Provider
}

func (a *adapterImpl) StreamChat(ctx context.Context, req ChatRequest) (<-chan StreamResponse, error) {
	// 1. Convert Request
	origReq := convertRequest(req)

	// 2. Call internal provider
	origChan, err := a.internal.StreamChat(ctx, origReq)
	if err != nil {
		return nil, err
	}

	// 3. Convert Response Channel (Bridge)
	outChan := make(chan StreamResponse)
	go func() {
		defer close(outChan)
		for resp := range origChan {
			outChan <- convertResponse(resp)
		}
	}()

	return outChan, nil
}

// --- Converters ---

func convertRequest(req ChatRequest) orig_llm.ChatRequest {
	// Deep copy / conversion
	msgs := make([]orig_llm.Message, len(req.Messages))
	for i, m := range req.Messages {
		multi := make([]orig_llm.MessageContentPart, len(m.MultiContent))
		for j, mc := range m.MultiContent {
			var imgURL *orig_llm.MessageImageURL
			if mc.ImageURL != nil {
				imgURL = &orig_llm.MessageImageURL{URL: mc.ImageURL.URL}
			}
			multi[j] = orig_llm.MessageContentPart{
				Type:     mc.Type,
				Text:     mc.Text,
				ImageURL: imgURL,
			}
		}

		tcs := make([]orig_llm.ToolCall, len(m.ToolCalls))
		for j, tc := range m.ToolCalls {
			tcs[j] = orig_llm.ToolCall{
				Index: tc.Index,
				ID:    tc.ID,
				Type:  tc.Type,
				Function: orig_llm.ToolCallFunction{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			}
		}

		msgs[i] = orig_llm.Message{
			Role:         m.Role,
			Content:      m.Content,
			MultiContent: multi,
			ToolCalls:    tcs,
			ToolCallID:   m.ToolCallID,
			ToolCallName: m.ToolCallName,
		}
	}

	tools := make([]orig_llm.Tool, len(req.Tools))
	for i, t := range req.Tools {
		// Deep copy Function parameters if needed, but interface{} can be passed directly as JSON marshaling handles it
		tools[i] = orig_llm.Tool{
			Type: t.Type,
			Function: orig_llm.Function{
				Name:        t.Function.Name,
				Description: t.Function.Description,
				Parameters:  t.Function.Parameters,
			},
		}
	}

	return orig_llm.ChatRequest{
		Model:       req.Model,
		Messages:    msgs,
		Tools:       tools,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Stream:      req.Stream,
	}
}

func convertResponse(resp orig_llm.StreamResponse) StreamResponse {
	tcs := make([]ToolCall, len(resp.ToolCalls))
	for i, tc := range resp.ToolCalls {
		tcs[i] = ToolCall{
			Index: tc.Index,
			ID:    tc.ID,
			Type:  tc.Type,
			Function: ToolCallFunction{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		}
	}
	
	// ToolOutputs conversion if any (though currently orig_llm might not populate it, added for future)
	tos := make([]ToolOutput, len(resp.ToolOutputs))
	for i, to := range resp.ToolOutputs {
		tos[i] = ToolOutput{
			ID:     to.ID,
			Name:   to.Name,
			Result: to.Result,
		}
	}

	var usage *Usage
	if resp.Usage != nil {
		usage = &Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		}
	}

	return StreamResponse{
		Content:     resp.Content,
		ToolCalls:   tcs,
		ToolOutputs: tos,
		Usage:       usage,
		Error:       resp.Error,
	}
}
