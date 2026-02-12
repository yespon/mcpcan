package llm

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
)

// NewProvider creates a new LLM provider instance based on the type and config
func NewProvider(typ ProviderType, config ProviderConfig) (Provider, error) {
	var model llms.Model
	var err error
	ctx := context.Background()

	switch typ {
	case ProviderOpenAI, ProviderDeepSeek, ProviderMoonshot, ProviderQwen, ProviderDoubao, ProviderZhipu, ProviderXAI, ProviderMistral, ProviderOpenRouter, ProviderLiteLLM, ProviderAzureOpenAI:
// ... (middle parts omitted for brevity in call, but I will include them to match context or just replace the end)
// Wait, replace_file_content doesn't support "..." in content. I should use exact content.
// Since I'm changing the function signature and the return statement, I might as well replace the whole function or the start and end.
// Let's replace the signature and the return statement separately or together if weak.
// Replacing the whole file content for factory.go is safer to avoid context issues.

		// All these providers use OpenAI-compatible API
		opts := []openai.Option{
			openai.WithToken(config.APIKey),
		}
		
		baseURL := config.BaseURL
		// Set default BaseURL if not provided
		if baseURL == "" {
			if defaultURL, ok := DefaultBaseURLs[typ]; ok && defaultURL != "" {
				baseURL = defaultURL
			}
		}

		if baseURL != "" {
			opts = append(opts, openai.WithBaseURL(baseURL))
		}

		if typ == ProviderOpenRouter {
			// Add OpenRouter specific headers
			client := &http.Client{
				Transport: &headerTransport{
					transport: http.DefaultTransport,
					headers: map[string]string{
						"HTTP-Referer": "https://github.com/kymo-mcp/mcpcan", // Optional. Site URL for rankings on openrouter.ai.
						"X-Title":      "MCPCan",                             // Optional. Site title for rankings on openrouter.ai.
					},
				},
			}
			opts = append(opts, openai.WithHTTPClient(client))
		}
		
		model, err = openai.New(opts...)

	case ProviderGoogle:
		// Google Gemini
		model, err = googleai.New(ctx,
			googleai.WithAPIKey(config.APIKey),
		)

	case ProviderAnthropic:
		// Anthropic Claude
		model, err = anthropic.New(
			anthropic.WithToken(config.APIKey),
		)

	case ProviderOllama:
		// Ollama Local
		opts := []ollama.Option{}
		baseURL := config.BaseURL
		if baseURL != "" {
			// LangChainGo 的 Ollama provider 使用 /api/chat 原生 API
			// 如果用户填写了 /v1 (OpenAI 兼容路径)，需要去掉，否则会变成 /v1/api/chat -> 404
			baseURL = strings.TrimSuffix(baseURL, "/v1")
			baseURL = strings.TrimSuffix(baseURL, "/v1/")
			opts = append(opts, ollama.WithServerURL(baseURL))
		}
		// Ollama client usually needs a model specified if it's not per-request, 
		// but LangChainGo ollama implementation allows per-request model most of the time.
		// Let's check if we need to set a default model. 
		// For now, we assume the model is passed in ChatRequest.
		model, err = ollama.New(opts...)

	default:
		return nil, fmt.Errorf("unsupported provider type: %s", typ)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialize provider %s: %w", typ, err)
	}

	return &LangChainAdapter{model: model}, nil
}

// headerTransport is a custom http.RoundTripper that adds headers to requests
type headerTransport struct {
	transport http.RoundTripper
	headers   map[string]string
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}
	
	// Debug logging
	// Avoid logging API Key
	logHeaders := make(http.Header)
	for k, v := range req.Header {
		if k == "Authorization" {
			logHeaders.Set(k, "Bearer ****")
		} else {
			logHeaders[k] = v
		}
	}
	log.Printf("[OpenRouter Debug] Request: %s %s, Headers: %v", req.Method, req.URL, logHeaders)

	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		// Read body for debugging
		bodyBytes, _ := io.ReadAll(resp.Body)
		// Restore body
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		
		log.Printf("[OpenRouter Error] Status: %d, Body: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}
