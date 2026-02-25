package models

// XAIProvider xAI 提供商信息
var XAIProvider = ProviderInfo{
	ID:          "xai",
	Name:        "xAI (Grok)",
	BaseURL:     "https://api.x.ai/v1",
	RegisterURL: "https://console.x.ai/",
	DocsURL:     "https://docs.x.ai/",
	Models: []ModelInfo{
		{
			ID:            "grok-3",
			Name:          "Grok 3",
			Description:   "xAI 旗舰模型,强大推理与工具能力",
			ContextLength: 128000,
			Modality:      "text+image->text",
			Provider:      "xai",
		},
		{
			ID:            "grok-3-mini",
			Name:          "Grok 3 Mini",
			Description:   "Grok 3 轻量版,快速响应",
			ContextLength: 128000,
			Modality:      "text+image->text",
			Provider:      "xai",
		},
		{
			ID:            "grok-2",
			Name:          "Grok 2",
			Description:   "Grok 2 标准模型",
			ContextLength: 128000,
			Modality:      "text+image->text",
			Provider:      "xai",
		},
	},
}

// XAIModels xAI 模型 ID 列表
var XAIModels = XAIProvider.GetModelIDs()
