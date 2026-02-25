package models

// OpenAIProvider OpenAI 提供商信息
// 数据来源：https://platform.openai.com/docs/models
var OpenAIProvider = ProviderInfo{
	ID:          "openai",
	Name:        "OpenAI",
	BaseURL:     "https://api.openai.com/v1",
	RegisterURL: "https://platform.openai.com/api-keys",
	DocsURL:     "https://platform.openai.com/docs",
	Models: []ModelInfo{
		// GPT-4.1 Series (2025-04 正式发布，API 持续可用)
		{
			ID:            "gpt-4.1",
			Name:          "GPT-4.1",
			Description:   "GPT-4.1 旗舰版，1M 上下文，多模态 | API 持续可用 (2025-04)",
			ContextLength: 1000000,
			Modality:      "text+image->text",
			Provider:      "openai",
			SupportTools:  true,
		},
		{
			ID:            "gpt-4.1-mini",
			Name:          "GPT-4.1 Mini",
			Description:   "GPT-4.1 轻量版，低延迟高性价比 | 1M 上下文",
			ContextLength: 1000000,
			Modality:      "text+image->text",
			Provider:      "openai",
			SupportTools:  true,
		},
		{
			ID:            "gpt-4.1-nano",
			Name:          "GPT-4.1 Nano",
			Description:   "GPT-4.1 极速版，最低延迟最低成本 | 1M 上下文",
			ContextLength: 1000000,
			Modality:      "text+image->text",
			Provider:      "openai",
			SupportTools:  true,
		},
		// GPT-4o Series (Stable)
		{
			ID:            "gpt-4o",
			Name:          "GPT-4o",
			Description:   "GPT-4 Omni 多模态旗舰，API 持续可用 | 128k 上下文",
			ContextLength: 128000,
			Modality:      "text+image->text",
			Provider:      "openai",
			SupportTools:  true,
		},
		{
			ID:            "gpt-4o-mini",
			Name:          "GPT-4o Mini",
			Description:   "GPT-4o 轻量版，高性价比 | 128k 上下文",
			ContextLength: 128000,
			Modality:      "text+image->text",
			Provider:      "openai",
			SupportTools:  true,
		},
		// O-Series Reasoning
		{
			ID:              "o4-mini",
			Name:            "o4 Mini",
			Description:     "o4 高效推理模型，低成本编程视觉推理 | 200k 上下文 (2025)",
			ContextLength:   200000,
			Modality:        "text+image->text",
			Provider:        "openai",
			SupportThinking: true,
			SupportTools:    true,
		},
		{
			ID:              "o3",
			Name:            "o3",
			Description:     "o3 旗舰推理模型，编程数学能力最强 | 200k 上下文 (2025)",
			ContextLength:   200000,
			Modality:        "text->text",
			Provider:        "openai",
			SupportThinking: true,
		},
		{
			ID:              "o3-mini",
			Name:            "o3 Mini",
			Description:     "o3 高效推理，低成本高速 | 200k 上下文",
			ContextLength:   200000,
			Modality:        "text->text",
			Provider:        "openai",
			SupportThinking: true,
		},
		{
			ID:              "o1",
			Name:            "o1",
			Description:     "o1 强推理模型，稳定版 | 128k 上下文",
			ContextLength:   128000,
			Modality:        "text->text",
			Provider:        "openai",
			SupportThinking: true,
		},
		{
			ID:              "o1-mini",
			Name:            "o1 Mini",
			Description:     "o1 Mini 轻量快速推理 | 128k 上下文",
			ContextLength:   128000,
			Modality:        "text->text",
			Provider:        "openai",
			SupportThinking: true,
		},
		// Legacy / Stable
		{
			ID:            "gpt-4-turbo",
			Name:          "GPT-4 Turbo",
			Description:   "GPT-4 Turbo 稳定版 | 128k 上下文",
			ContextLength: 128000,
			Modality:      "text+image->text",
			Provider:      "openai",
			SupportTools:  true,
		},
	},
}

// OpenAIModels OpenAI 模型 ID 列表 (兼容旧接口)
var OpenAIModels = OpenAIProvider.GetModelIDs()
