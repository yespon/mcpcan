package models

// GoogleProvider Google AI 提供商信息
var GoogleProvider = ProviderInfo{
	ID:          "google",
	Name:        "Google AI",
	BaseURL:     "https://generativelanguage.googleapis.com/v1beta",
	RegisterURL: "https://aistudio.google.com/app/apikey",
	DocsURL:     "https://ai.google.dev/docs",
	Models: []ModelInfo{

		// Gemini 3.0 Series (Preview)
		{
			ID:            "gemini-3-pro-preview",
			Name:          "Gemini 3 Pro Preview",
			Description:   "Gemini 3 Next Gen Preview",
			ContextLength: 1000000,
			Modality:      "text+image+video->text",
			Provider:      "google",
			SupportTools:  true,
		},
		{
			ID:            "gemini-3-flash-preview",
			Name:          "Gemini 3 Flash Preview",
			Description:   "Gemini 3 Fast Preview",
			ContextLength: 1048576,
			Modality:      "text+image+video->text",
			Provider:      "google",
			SupportTools:  true,
		},
		// Gemini 2.5 Series
		{
			ID:            "gemini-2.5-pro",
			Name:          "Gemini 2.5 Pro",
			Description:   "Gemini 2.5 Pro High Intelligence",
			ContextLength: 1048576,
			Modality:      "text+image+video->text",
			Provider:      "google",
			SupportTools:  true,
		},
		{
			ID:            "gemini-2.5-flash",
			Name:          "Gemini 2.5 Flash",
			Description:   "Gemini 2.5 Flash High Speed",
			ContextLength: 1048576,
			Modality:      "text+image+video->text",
			Provider:      "google",
			SupportTools:  true,
		},
		// Gemini 2.0 Series
		{
			ID:              "gemini-2.0-flash-exp",
			Name:            "Gemini 2.0 Flash Exp",
			Description:     "Gemini 2.0 预览版,多模态实时交互",
			ContextLength:   1048576,
			Modality:        "text+image->text",
			Provider:        "google",
			SupportThinking: false,
			SupportTools:    true,
		},
		{
			ID:              "gemini-2.0-flash-thinking-exp-1219",
			Name:            "Gemini 2.0 Flash Thinking",
			Description:     "Gemini 2.0 思考模型",
			ContextLength:   32768,
			Modality:        "text+image->text",
			Provider:        "google",
			SupportThinking: true,
			SupportTools:    false,
		},
		// Gemini 1.5 Series (Latest 002)
		{
			ID:              "gemini-1.5-pro-002",
			Name:            "Gemini 1.5 Pro-002",
			Description:     "Gemini 1.5 Pro 最新稳定版",
			ContextLength:   2000000,
			Modality:        "text+image->text",
			Provider:        "google",
			SupportThinking: false,
			SupportTools:    true,
		},
		{
			ID:              "gemini-1.5-flash-002",
			Name:            "Gemini 1.5 Flash-002",
			Description:     "Gemini 1.5 Flash 最新稳定版",
			ContextLength:   1000000,
			Modality:        "text+image->text",
			Provider:        "google",
			SupportThinking: false,
			SupportTools:    true,
		},
		{
			ID:              "gemini-1.5-pro",
			Name:            "Gemini 1.5 Pro",
			Description:     "Gemini 1.5 Pro 经典版",
			ContextLength:   1000000,
			Modality:        "text+image->text",
			Provider:        "google",
			SupportThinking: false,
			SupportTools:    true,
		},
		{
			ID:              "gemini-1.5-flash",
			Name:            "Gemini 1.5 Flash",
			Description:     "Gemini 1.5 Flash 经典版",
			ContextLength:   1000000,
			Modality:        "text+image->text",
			Provider:        "google",
			SupportThinking: false,
			SupportTools:    true,
		},
	},
}

// GoogleModels Google 模型 ID 列表
var GoogleModels = GoogleProvider.GetModelIDs()
