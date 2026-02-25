package models

// AnthropicProvider Anthropic 提供商信息
// 数据来源：https://docs.anthropic.com/en/docs/about-claude/models/overview
var AnthropicProvider = ProviderInfo{
	ID:          "anthropic",
	Name:        "Anthropic",
	BaseURL:     "https://api.anthropic.com/v1",
	RegisterURL: "https://console.anthropic.com/account/keys",
	DocsURL:     "https://docs.anthropic.com/",
	Models: []ModelInfo{
		// Claude 4.6 Series (2026-02 最新，官方确认可用)
		{
			ID:            "claude-opus-4-6-20260205",
			Name:          "Claude Opus 4.6",
			Description:   "Anthropic 当前最强旗舰，编程推理最优 | 200k 上下文（beta: 1M）",
			ContextLength: 200000,
			Modality:      "text+image->text",
			Provider:      "anthropic",
			SupportTools:  true,
		},
		{
			ID:            "claude-sonnet-4-6",
			Name:          "Claude Sonnet 4.6",
			Description:   "Claude 4.6 Sonnet，Agent 与专业任务首选 | 200k 上下文",
			ContextLength: 200000,
			Modality:      "text+image->text",
			Provider:      "anthropic",
			SupportTools:  true,
		},
		// Claude 4.5 Series (2025 Q3-Q4)
		{
			ID:            "claude-opus-4-5-20251101",
			Name:          "Claude Opus 4.5",
			Description:   "Claude Opus 4.5，复杂推理与 Agentic 任务 | 200k 上下文",
			ContextLength: 200000,
			Modality:      "text+image->text",
			Provider:      "anthropic",
			SupportTools:  true,
		},
		{
			ID:            "claude-sonnet-4-5-20250929",
			Name:          "Claude Sonnet 4.5",
			Description:   "Claude Sonnet 4.5，编程与 Agent 均衡首选 | 200k 上下文",
			ContextLength: 200000,
			Modality:      "text+image->text",
			Provider:      "anthropic",
			SupportTools:  true,
		},
		{
			ID:            "claude-haiku-4-5-20251001",
			Name:          "Claude Haiku 4.5",
			Description:   "Claude Haiku 4.5，低成本高速推理 | 200k 上下文",
			ContextLength: 200000,
			Modality:      "text+image->text",
			Provider:      "anthropic",
			SupportTools:  true,
		},
		// Claude 4 Series (2025-05)
		{
			ID:            "claude-opus-4-20250514",
			Name:          "Claude Opus 4",
			Description:   "Claude 4 旗舰，复杂推理与 Agent 基准 | 200k 上下文",
			ContextLength: 200000,
			Modality:      "text+image->text",
			Provider:      "anthropic",
			SupportTools:  true,
		},
		{
			ID:            "claude-sonnet-4-20250514",
			Name:          "Claude Sonnet 4",
			Description:   "Claude 4 均衡版，性价比优秀 | 200k 上下文",
			ContextLength: 200000,
			Modality:      "text+image->text",
			Provider:      "anthropic",
			SupportTools:  true,
		},
		// Claude 3.7 Series (2025-02) — 混合推理
		{
			ID:              "claude-3-7-sonnet-20250219",
			Name:            "Claude 3.7 Sonnet",
			Description:     "首款混合推理模型，编程能力突出 | 200k 上下文",
			ContextLength:   200000,
			Modality:        "text+image->text",
			Provider:        "anthropic",
			SupportThinking: true,
			SupportTools:    true,
		},
		// Claude 3.5 Series (Stable)
		{
			ID:           "claude-3-5-sonnet-20241022",
			Name:         "Claude 3.5 Sonnet",
			Description:  "Claude 3.5 Sonnet 稳定版，综合能力优秀",
			ContextLength: 200000,
			Modality:     "text+image->text",
			Provider:     "anthropic",
			SupportTools: true,
		},
		{
			ID:           "claude-3-5-haiku-20241022",
			Name:         "Claude 3.5 Haiku",
			Description:  "Claude 3.5 Haiku，低成本快速",
			ContextLength: 200000,
			Modality:     "text+image->text",
			Provider:     "anthropic",
			SupportTools: true,
		},
	},
}

// AnthropicModels Anthropic 模型 ID 列表
var AnthropicModels = AnthropicProvider.GetModelIDs()
