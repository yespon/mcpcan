package llm

import (
	"context"
	"fmt"

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
			switch typ {
			case ProviderDeepSeek:
				baseURL = "https://api.deepseek.com" // langchaingo/openai will append /v1 if not using WithAPIType? No, usually expect full base or v1. OpenAI default is v1.
				// OpenAI SDK expects base URL. If it ends with /v1, it uses it.
				// Let's use the explicit /v1 to be safe as per docs.
				baseURL = "https://api.deepseek.com/v1"
			case ProviderQwen:
				baseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
			case ProviderDoubao:
				// Doubao/Volcengine OpenAI compatible endpoint
				baseURL = "https://ark.cn-beijing.volces.com/api/v3"
			case ProviderMoonshot:
				baseURL = "https://api.moonshot.cn/v1"
			}
		}

		if baseURL != "" {
			opts = append(opts, openai.WithBaseURL(baseURL))
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
		if config.BaseURL != "" {
			opts = append(opts, ollama.WithServerURL(config.BaseURL))
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
