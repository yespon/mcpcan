package models

// AzureOpenAIProvider Azure OpenAI 服务提供商信息
var AzureOpenAIProvider = ProviderInfo{
	ID:          "azure_openai",
	Name:        "Azure OpenAI",
	BaseURL:     "",
	RegisterURL: "https://portal.azure.com/",
	DocsURL:     "https://learn.microsoft.com/en-us/azure/cognitive-services/openai/",
	Models:      AzureOpenAIModels,
}

// AzureOpenAIModels Azure OpenAI 支持的模型列表
var AzureOpenAIModels = []ModelInfo{
	{
		ID:              "gpt-4o",
		Name:            "GPT-4o",
		Description:     "Most advanced multimodal model, optimized for speed and cost",
		ContextLength:   128000,
		Modality:        "text+image->text",
		Provider:        "azure_openai",
		SupportThinking: true,
		SupportTools:    true,
	},
	{
		ID:              "gpt-4o-mini",
		Name:            "GPT-4o mini",
		Description:     "Affordable multimodal model, optimized for speed",
		ContextLength:   128000,
		Modality:        "text+image->text",
		Provider:        "azure_openai",
		SupportThinking: false,
		SupportTools:    true,
	},
	{
		ID:              "gpt-4",
		Name:            "GPT-4",
		Description:     "More capable than GPT-3.5, able to do more complex tasks",
		ContextLength:   8192,
		Modality:        "text->text",
		Provider:        "azure_openai",
		SupportThinking: true,
		SupportTools:    true,
	},
	{
		ID:              "gpt-35-turbo",
		Name:            "GPT-3.5 Turbo",
		Description:     "Legacy model, still capable but less powerful than GPT-4",
		ContextLength:   16385,
		Modality:        "text->text",
		Provider:        "azure_openai",
		SupportThinking: false,
		SupportTools:    true,
	},
	{
		ID:              "gpt-4-turbo",
		Name:            "GPT-4 Turbo",
		Description:     "GPT-4 with improved performance and reduced latency",
		ContextLength:   128000,
		Modality:        "text+image->text",
		Provider:        "azure_openai",
		SupportThinking: true,
		SupportTools:    true,
	},
	{
		ID:              "text-embedding-ada-002",
		Name:            "Text Embedding Ada v2",
		Description:     "High-quality embedding model for text",
		ContextLength:   8191,
		Modality:        "text->text",
		Provider:        "azure_openai",
		SupportThinking: false,
		SupportTools:    false,
	},
}
