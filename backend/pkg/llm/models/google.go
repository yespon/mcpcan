package models

// GoogleProvider Google AI 提供商信息
var GoogleProvider = ProviderInfo{
	ID:          "google",
	Name:        "Google AI",
	BaseURL:     "https://generativelanguage.googleapis.com/v1beta",
	RegisterURL: "https://aistudio.google.com/app/apikey",
	DocsURL:     "https://ai.google.dev/docs",
	Models: []ModelInfo{
		// Gemini 2.5 系列
		{
			ID:            "gemini-2.5-pro-preview-05-06",
			Name:          "Gemini 2.5 Pro",
			Description:   "Google 最新旗舰模型,增强推理和工具调用",
			ContextLength: 1048576,
			Modality:      "text+image->text",
			Provider:      "google",
		},
		{
			ID:            "gemini-2.5-flash-preview-05-20",
			Name:          "Gemini 2.5 Flash",
			Description:   "Gemini 2.5 快速版,平衡性能与速度",
			ContextLength: 1048576,
			Modality:      "text+image->text",
			Provider:      "google",
		},
		// Gemini 2.0 系列
		{
			ID:            "gemini-2.0-flash",
			Name:          "Gemini 2.0 Flash",
			Description:   "Gemini 2.0 快速模型",
			ContextLength: 1048576,
			Modality:      "text+image->text",
			Provider:      "google",
		},
		{
			ID:            "gemini-2.0-flash-lite",
			Name:          "Gemini 2.0 Flash Lite",
			Description:   "Gemini 2.0 轻量版,低延迟",
			ContextLength: 1048576,
			Modality:      "text+image->text",
			Provider:      "google",
		},
	},
}

// GoogleModels Google 模型 ID 列表
var GoogleModels = GoogleProvider.GetModelIDs()
