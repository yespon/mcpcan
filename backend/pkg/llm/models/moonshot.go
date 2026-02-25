package models

// MoonshotProvider 月之暗面 Kimi 提供商信息
// 数据来源：https://platform.moonshot.cn/docs/api/chat
var MoonshotProvider = ProviderInfo{
	ID:          "moonshot",
	Name:        "月之暗面 Kimi (Moonshot)",
	BaseURL:     "https://api.moonshot.cn/v1",
	RegisterURL: "https://platform.moonshot.cn/console/api-keys",
	DocsURL:     "https://platform.moonshot.cn/docs",
	Models: []ModelInfo{
		// Kimi K2.5 Series (2026-01 最新)
		{
			ID:              "kimi-k2.5",
			Name:            "Kimi K2.5",
			Description:     "Kimi 最新旗舰多模态模型，支持思考/非思考双模式 | 256k 上下文 (2026-01)",
			ContextLength:   262144,
			Modality:        "text+image->text",
			Provider:        "moonshot",
			SupportThinking: true,
			SupportTools:    true,
		},
		// Kimi K2 Series (2025-09)
		{
			ID:              "kimi-k2-thinking",
			Name:            "Kimi K2 Thinking",
			Description:     "Kimi K2 深度推理版，强化学习训练 | 256k 上下文",
			ContextLength:   262144,
			Modality:        "text->text",
			Provider:        "moonshot",
			SupportThinking: true,
			SupportTools:    true,
		},
		{
			ID:              "kimi-k2-thinking-turbo",
			Name:            "Kimi K2 Thinking Turbo",
			Description:     "Kimi K2 推理快速版 | 256k 上下文",
			ContextLength:   262144,
			Modality:        "text->text",
			Provider:        "moonshot",
			SupportThinking: true,
			SupportTools:    true,
		},
		{
			ID:           "kimi-k2-0905-preview",
			Name:         "Kimi K2 Preview (Sep 2025)",
			Description:  "Kimi K2 基座模型，编程和 Agent 能力强 | 256k 上下文，1T 参数",
			ContextLength: 262144,
			Modality:     "text->text",
			Provider:     "moonshot",
			SupportTools: true,
		},
		{
			ID:           "kimi-k2-turbo-preview",
			Name:         "Kimi K2 Turbo Preview",
			Description:  "Kimi K2 Turbo，高性价比版本 | 256k 上下文",
			ContextLength: 262144,
			Modality:     "text->text",
			Provider:     "moonshot",
			SupportTools: true,
		},
		// Moonshot V1 Classic (按上下文长度区分)
		{
			ID:           "moonshot-v1-128k",
			Name:         "Moonshot V1 128K",
			Description:  "Kimi 经典版，超长上下文 | 128k 上下文",
			ContextLength: 128000,
			Modality:     "text->text",
			Provider:     "moonshot",
			SupportTools: true,
		},
		{
			ID:           "moonshot-v1-32k",
			Name:         "Moonshot V1 32K",
			Description:  "Kimi 经典版，标准上下文 | 32k 上下文",
			ContextLength: 32000,
			Modality:     "text->text",
			Provider:     "moonshot",
			SupportTools: true,
		},
		{
			ID:           "moonshot-v1-8k",
			Name:         "Moonshot V1 8K",
			Description:  "Kimi 经典版，低延迟 | 8k 上下文",
			ContextLength: 8000,
			Modality:     "text->text",
			Provider:     "moonshot",
			SupportTools: true,
		},
	},
}

// MoonshotModels Moonshot 模型 ID 列表
var MoonshotModels = MoonshotProvider.GetModelIDs()
