package models

// MoonshotProvider 月之暗面 Moonshot 提供商信息
var MoonshotProvider = ProviderInfo{
	ID:          "moonshot",
	Name:        "月之暗面 Kimi (Moonshot)",
	BaseURL:     "https://api.moonshot.cn/v1",
	RegisterURL: "https://platform.moonshot.cn/console/api-keys",
	DocsURL:     "https://platform.moonshot.cn/docs",
	Models: []ModelInfo{
		{
			ID:            "moonshot-v1-128k",
			Name:          "Moonshot V1 128K",
			Description:   "Kimi 大窗口模型,128K 上下文",
			ContextLength: 128000,
			Modality:      "text->text",
			Provider:      "moonshot",
		},
		{
			ID:            "moonshot-v1-32k",
			Name:          "Moonshot V1 32K",
			Description:   "Kimi 标准模型,32K 上下文",
			ContextLength: 32000,
			Modality:      "text->text",
			Provider:      "moonshot",
		},
		{
			ID:            "moonshot-v1-8k",
			Name:          "Moonshot V1 8K",
			Description:   "Kimi 快速模型,8K 上下文",
			ContextLength: 8000,
			Modality:      "text->text",
			Provider:      "moonshot",
		},
	},
}

// MoonshotModels Moonshot 模型 ID 列表
var MoonshotModels = MoonshotProvider.GetModelIDs()
