package models

// MoonshotProvider 月之暗面 Moonshot 提供商信息
var MoonshotProvider = ProviderInfo{
	ID:          "moonshot",
	Name:        "月之暗面 Kimi (Moonshot)",
	BaseURL:     "https://api.moonshot.cn/v1",
	RegisterURL: "https://platform.moonshot.cn/console/api-keys",
	DocsURL:     "https://platform.moonshot.cn/docs",
	Models: []ModelInfo{
		// Kimi k2.5 Series
		{
			ID:              "kimi-k2.5",
			Name:            "Kimi k2.5",
			Description:     "Kimi k2.5 Multimodal (Jan 2026)",
			ContextLength:   262144,
			Modality:        "text+image+video->text",
			Provider:        "moonshot",
			SupportThinking: true,
			SupportTools:    true,
		},
		// Kimi k2 Series
		{
			ID:              "kimi-k2-thinking",
			Name:            "Kimi k2 Thinking",
			Description:     "Kimi k2 Reasoning Model",
			ContextLength:   262144,
			Modality:        "text->text",
			Provider:        "moonshot",
			SupportThinking: true,
			SupportTools:    true,
		},
		{
			ID:            "kimi-k2-instruct",
			Name:            "Kimi k2 Instruct",
			Description:     "Kimi k2 Instruct",
			ContextLength:   262144,
			Modality:        "text->text",
			Provider:        "moonshot",
			SupportTools:    true,
		},
		// Legacy / Stable (Moonshot V1)
		{
			ID:            "moonshot-v1-128k",
			Name:          "Moonshot V1 128K",
			Description:   "Kimi Classic 128K",
			ContextLength: 128000,
			Modality:      "text->text",
			Provider:      "moonshot",
		},
		{
			ID:            "moonshot-v1-32k",
			Name:          "Moonshot V1 32K",
			Description:   "Kimi Classic 32K",
			ContextLength: 32000,
			Modality:      "text->text",
			Provider:      "moonshot",
		},
		{
			ID:            "moonshot-v1-8k",
			Name:          "Moonshot V1 8K",
			Description:   "Kimi Classic 8K",
			ContextLength: 8000,
			Modality:      "text->text",
			Provider:      "moonshot",
		},
	},
}

// MoonshotModels Moonshot 模型 ID 列表
var MoonshotModels = MoonshotProvider.GetModelIDs()
