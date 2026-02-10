package models

// AnthropicProvider Anthropic 提供商信息
var AnthropicProvider = ProviderInfo{
	ID:          "anthropic",
	Name:        "Anthropic",
	BaseURL:     "https://api.anthropic.com/v1",
	RegisterURL: "https://console.anthropic.com/account/keys",
	DocsURL:     "https://docs.anthropic.com/",
	Models: []ModelInfo{
		// Claude 4.5 Series (2025)
		{
			ID:            "claude-opus-4-5",
			Name:          "Claude Opus 4.5",
			Description:   "Claude 4.5 Opus (Late 2025)",
			ContextLength: 200000,
			Modality:      "text+image->text",
			Provider:      "anthropic",
		},
		{
			ID:            "claude-sonnet-4-5",
			Name:          "Claude Sonnet 4.5",
			Description:   "Claude 4.5 Sonnet (Sep 2025)",
			ContextLength: 200000,
			Modality:      "text+image->text",
			Provider:      "anthropic",
		},
		{
			ID:            "claude-haiku-4-5",
			Name:          "Claude Haiku 4.5",
			Description:   "Claude 4.5 Haiku (Oct 2025)",
			ContextLength: 200000,
			Modality:      "text+image->text",
			Provider:      "anthropic",
		},
		// Claude 3.5 Series
		{
			ID:            "claude-3-5-sonnet-20241022",
			Name:          "Claude 3.5 Sonnet",
			Description:   "Claude 3.5 Sonnet (New)",
			ContextLength: 200000,
			Modality:      "text+image->text",
			Provider:      "anthropic",
		},
		{
			ID:            "claude-3-5-haiku-20241022",
			Name:          "Claude 3.5 Haiku",
			Description:   "Claude 3.5 Haiku",
			ContextLength: 200000,
			Modality:      "text+image->text",
			Provider:      "anthropic",
		},
		// Claude 3 Series
		{
			ID:            "claude-3-opus-20240229",
			Name:          "Claude 3 Opus",
			Description:   "Claude 3 Opus",
			ContextLength: 200000,
			Modality:      "text+image->text",
			Provider:      "anthropic",
		},
	},
}

// AnthropicModels Anthropic 模型 ID 列表
var AnthropicModels = AnthropicProvider.GetModelIDs()
