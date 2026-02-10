package models

// OpenAIProvider OpenAI 提供商信息
var OpenAIProvider = ProviderInfo{
	ID:          "openai",
	Name:        "OpenAI",
	BaseURL:     "https://api.openai.com/v1",
	RegisterURL: "https://platform.openai.com/api-keys",
	DocsURL:     "https://platform.openai.com/docs",
	Models: []ModelInfo{
		// GPT-5 Series (2025-2026)
		{
			ID:            "gpt-5",
			Name:          "GPT-5",
			Description:   "GPT-5 Flagship, Superior Intelligence (2025)",
			ContextLength: 400000,
			Modality:      "text+image->text",
			Provider:      "openai",
		},
		{
			ID:            "gpt-5-mini",
			Name:          "GPT-5 Mini",
			Description:   "GPT-5 Compact, High Efficiency",
			ContextLength: 400000,
			Modality:      "text+image->text",
			Provider:      "openai",
		},
		{
			ID:            "gpt-5.1",
			Name:          "GPT-5.1",
			Description:   "GPT-5.1 Enhanced (Nov 2025)",
			ContextLength: 400000,
			Modality:      "text+image->text",
			Provider:      "openai",
		},
		{
			ID:            "gpt-5.2",
			Name:          "GPT-5.2",
			Description:   "GPT-5.2 Advanced (Dec 2025)",
			ContextLength: 400000,
			Modality:      "text+image->text",
			Provider:      "openai",
		},
		// O-Series Reasoning (o3/o1)
		{
			ID:              "o3-mini",
			Name:            "o3 Mini",
			Description:     "o3 Reasoning Model (2025)",
			ContextLength:   200000,
			Modality:        "text->text",
			Provider:        "openai",
			SupportThinking: true,
		},
		{
			ID:              "o1",
			Name:            "o1",
			Description:     "o1 Reasoning Model (Final)",
			ContextLength:   128000,
			Modality:        "text->text",
			Provider:        "openai",
			SupportThinking: true,
		},
		{
			ID:              "o1-mini",
			Name:            "o1 Mini",
			Description:     "Faster reasoning model",
			ContextLength:   128000,
			Modality:        "text->text",
			Provider:        "openai",
			SupportThinking: true,
		},
		// GPT-4o Series
		{
			ID:            "gpt-4o",
			Name:          "GPT-4o",
			Description:   "GPT-4 Omni",
			ContextLength: 128000,
			Modality:      "text+image->text",
			Provider:      "openai",
			SupportTools:  true,
		},
		{
			ID:            "gpt-4o-mini",
			Name:          "GPT-4o Mini",
			Description:   "Cost-efficient small model",
			ContextLength: 128000,
			Modality:      "text+image->text",
			Provider:      "openai",
			SupportTools:  true,
		},
		// Legacy / Stable
		{
			ID:            "gpt-4-turbo",
			Name:          "GPT-4 Turbo",
			ContextLength: 128000,
			Modality:      "text+image->text",
			Provider:      "openai",
			SupportTools:  true,
		},
		{
			ID:            "gpt-3.5-turbo",
			Name:          "GPT-3.5 Turbo",
			ContextLength: 16385,
			Modality:      "text->text",
			Provider:      "openai",
			SupportTools:  true,
		},
	},
}

// OpenAIModels OpenAI 模型 ID 列表 (兼容旧接口)
var OpenAIModels = OpenAIProvider.GetModelIDs()
