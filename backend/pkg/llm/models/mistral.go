package models

// MistralProvider Mistral AI 提供商信息
var MistralProvider = ProviderInfo{
	ID:          "mistral",
	Name:        "Mistral AI",
	BaseURL:     "https://api.mistral.ai/v1",
	RegisterURL: "https://console.mistral.ai/api-keys/",
	DocsURL:     "https://docs.mistral.ai/",
	Models: []ModelInfo{
		{
			ID:            "mistral-large-latest",
			Name:          "Mistral Large",
			Description:   "Mistral 旗舰模型,强大工具调用能力",
			ContextLength: 128000,
			Modality:      "text->text",
			Provider:      "mistral",
		},
		{
			ID:            "mistral-medium-latest",
			Name:          "Mistral Medium",
			Description:   "Mistral 中等模型,平衡性能",
			ContextLength: 32000,
			Modality:      "text->text",
			Provider:      "mistral",
		},
		{
			ID:            "mistral-small-latest",
			Name:          "Mistral Small",
			Description:   "Mistral 小型模型,快速响应",
			ContextLength: 32000,
			Modality:      "text->text",
			Provider:      "mistral",
		},
		{
			ID:            "codestral-latest",
			Name:          "Codestral",
			Description:   "Mistral 代码专用模型,支持工具调用",
			ContextLength: 32000,
			Modality:      "text->text",
			Provider:      "mistral",
		},
	},
}

// MistralModels Mistral 模型 ID 列表
var MistralModels = MistralProvider.GetModelIDs()
