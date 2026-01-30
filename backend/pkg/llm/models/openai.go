package models

// OpenAIProvider OpenAI 提供商信息
var OpenAIProvider = ProviderInfo{
	ID:          "openai",
	Name:        "OpenAI",
	BaseURL:     "https://api.openai.com/v1",
	RegisterURL: "https://platform.openai.com/api-keys",
	DocsURL:     "https://platform.openai.com/docs",
	Models: []ModelInfo{
		// GPT-4o
		{
			ID:              "gpt-4o",
			Name:            "GPT-4o",
			Description:     "GPT-4 Omni, Multimodal, Faster",
			ContextLength:   128000,
			Modality:        "text+image->text",
			Provider:        "openai",
			SupportThinking: false,
			SupportTools:    true,
		},
		{
			ID:              "gpt-4o-2024-08-06",
			Name:            "GPT-4o (0806)",
			Description:     "GPT-4 Omni (Aug 2024), Structured Output",
			ContextLength:   128000,
			Modality:        "text+image->text",
			Provider:        "openai",
			SupportThinking: false,
			SupportTools:    true,
		},
		{
			ID:              "gpt-4o-mini",
			Name:            "GPT-4o Mini",
			Description:     "Cost-efficient small model",
			ContextLength:   128000,
			Modality:        "text+image->text",
			Provider:        "openai",
			SupportThinking: false,
			SupportTools:    true,
		},
		// O1 Reasoning
		{
			ID:              "o1-preview",
			Name:            "o1 Preview",
			Description:     "Reasoning model, deep thinking",
			ContextLength:   128000,
			Modality:        "text->text",
			Provider:        "openai",
			SupportThinking: true,
			SupportTools:    false, // Currently o1 does not support tools well or at all in initial preview
		},
		{
			ID:              "o1-mini",
			Name:            "o1 Mini",
			Description:     "Faster reasoning model",
			ContextLength:   128000,
			Modality:        "text->text",
			Provider:        "openai",
			SupportThinking: true,
			SupportTools:    false,
		},
		// Legacy / Stable
		{
			ID:              "gpt-4-turbo",
			Name:            "GPT-4 Turbo",
			Description:     "GPT-4 Turbo with Vision",
			ContextLength:   128000,
			Modality:        "text+image->text",
			Provider:        "openai",
			SupportThinking: false,
			SupportTools:    true,
		},
		{
			ID:              "gpt-4",
			Name:            "GPT-4",
			Description:     "Classic GPT-4",
			ContextLength:   8192,
			Modality:        "text->text",
			Provider:        "openai",
			SupportThinking: false,
			SupportTools:    true,
		},
		{
			ID:              "gpt-3.5-turbo",
			Name:            "GPT-3.5 Turbo",
			Description:     "Fast, cheap",
			ContextLength:   16385,
			Modality:        "text->text",
			Provider:        "openai",
			SupportThinking: false,
			SupportTools:    true,
		},
	},
}

// OpenAIModels OpenAI 模型 ID 列表 (兼容旧接口)
var OpenAIModels = OpenAIProvider.GetModelIDs()
