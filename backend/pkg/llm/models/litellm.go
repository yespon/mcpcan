package models

// LiteLLMProvider LiteLLM 服务提供商信息
var LiteLLMProvider = ProviderInfo{
	ID:          "litellm",
	Name:        "LiteLLM",
	BaseURL:     "http://localhost:8000", // 默认本地部署地址
	RegisterURL: "https://litellm.vercel.app/docs/proxy/configurations",
	DocsURL:     "https://litellm.vercel.app/",
	Models:      LiteLLMModels,
}

// LiteLLMModels LiteLLM 支持的模型列表（动态，依赖后端配置）
var LiteLLMModels = []ModelInfo{
	{
		ID:              "openai/gpt-4o",
		Name:            "GPT-4o (via LiteLLM)",
		Description:     "OpenAI GPT-4o model accessed through LiteLLM proxy",
		ContextLength:   128000,
		Modality:        "text+image->text",
		Provider:        "litellm",
		SupportThinking: true,
		SupportTools:    true,
	},
	{
		ID:              "openai/gpt-4o-mini",
		Name:            "GPT-4o Mini (via LiteLLM)",
		Description:     "OpenAI GPT-4o mini model accessed through LiteLLM proxy",
		ContextLength:   128000,
		Modality:        "text+image->text",
		Provider:        "litellm",
		SupportThinking: false,
		SupportTools:    true,
	},
	{
		ID:              "openai/gpt-35-turbo",
		Name:            "GPT-3.5 Turbo (via LiteLLM)",
		Description:     "OpenAI GPT-3.5 Turbo model accessed through LiteLLM proxy",
		ContextLength:   16385,
		Modality:        "text->text",
		Provider:        "litellm",
		SupportThinking: false,
		SupportTools:    true,
	},
	{
		ID:              "anthropic/claude-3-5-sonnet",
		Name:            "Claude 3.5 Sonnet (via LiteLLM)",
		Description:     "Anthropic Claude 3.5 Sonnet accessed through LiteLLM proxy",
		ContextLength:   200000,
		Modality:        "text+image->text",
		Provider:        "litellm",
		SupportThinking: false,
		SupportTools:    true,
	},
	{
		ID:              "google/gemini-1.5-pro",
		Name:            "Gemini 1.5 Pro (via LiteLLM)",
		Description:     "Google Gemini 1.5 Pro accessed through LiteLLM proxy",
		ContextLength:   1000000,
		Modality:        "text+image->text",
		Provider:        "litellm",
		SupportThinking: false,
		SupportTools:    true,
	},
	{
		ID:              "ollama/llama3",
		Name:            "Llama 3 (via LiteLLM)",
		Description:     "Ollama Llama 3 model accessed through LiteLLM proxy",
		ContextLength:   8192,
		Modality:        "text->text",
		Provider:        "litellm",
		SupportThinking: false,
		SupportTools:    false,
	},
}
