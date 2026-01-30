package models

// AnthropicProvider Anthropic 提供商信息
var AnthropicProvider = ProviderInfo{
	ID:          "anthropic",
	Name:        "Anthropic",
	BaseURL:     "https://api.anthropic.com/v1",
	RegisterURL: "https://console.anthropic.com/account/keys",
	DocsURL:     "https://docs.anthropic.com/",
	Models: []ModelInfo{
		// Claude 4 系列
		{
			ID:            "claude-sonnet-4-20250514",
			Name:          "Claude Sonnet 4",
			Description:   "Anthropic 最新旗舰模型,顶级智能与工具能力",
			ContextLength: 200000,
			Modality:      "text+image->text",
			Provider:      "anthropic",
		},
		// Claude 3.5 系列
		{
			ID:            "claude-3-5-sonnet-20241022",
			Name:          "Claude 3.5 Sonnet",
			Description:   "Claude 3.5 Sonnet,平衡智能与速度",
			ContextLength: 200000,
			Modality:      "text+image->text",
			Provider:      "anthropic",
		},
		{
			ID:            "claude-3-5-haiku-20241022",
			Name:          "Claude 3.5 Haiku",
			Description:   "Claude 3.5 Haiku,快速响应低成本",
			ContextLength: 200000,
			Modality:      "text+image->text",
			Provider:      "anthropic",
		},
		// Claude 3 系列
		{
			ID:            "claude-3-opus-20240229",
			Name:          "Claude 3 Opus",
			Description:   "Claude 3 旗舰模型,强大推理能力",
			ContextLength: 200000,
			Modality:      "text+image->text",
			Provider:      "anthropic",
		},
	},
}

// AnthropicModels Anthropic 模型 ID 列表
var AnthropicModels = AnthropicProvider.GetModelIDs()
