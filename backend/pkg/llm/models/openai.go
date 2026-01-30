package models

import (
	openai "github.com/sashabaranov/go-openai"
)

// OpenAIProvider OpenAI 提供商信息
var OpenAIProvider = ProviderInfo{
	ID:          "openai",
	Name:        "OpenAI",
	BaseURL:     "https://api.openai.com/v1",
	RegisterURL: "https://platform.openai.com/api-keys",
	DocsURL:     "https://platform.openai.com/docs",
	Models: []ModelInfo{
		// GPT-5 系列
		{
			ID:            openai.GPT5,
			Name:          "GPT-5",
			Description:   "OpenAI 最新旗舰模型,顶级智能",
			ContextLength: 400000,
			Modality:      "text+image->text",
			Provider:      "openai",
		},
		{
			ID:            openai.GPT5Mini,
			Name:          "GPT-5 Mini",
			Description:   "GPT-5 轻量版,平衡性能和成本",
			ContextLength: 400000,
			Modality:      "text+image->text",
			Provider:      "openai",
		},
		{
			ID:            openai.GPT5Nano,
			Name:          "GPT-5 Nano",
			Description:   "GPT-5 最小版,低延迟快速响应",
			ContextLength: 400000,
			Modality:      "text+image->text",
			Provider:      "openai",
		},
		// GPT-4.1 系列
		{
			ID:            openai.GPT4Dot1,
			Name:          "GPT-4.1",
			Description:   "GPT-4.1 标准版,1M 上下文",
			ContextLength: 1047576,
			Modality:      "text+image->text",
			Provider:      "openai",
		},
		{
			ID:            openai.GPT4Dot1Mini,
			Name:          "GPT-4.1 Mini",
			Description:   "GPT-4.1 轻量版",
			ContextLength: 1047576,
			Modality:      "text+image->text",
			Provider:      "openai",
		},
		{
			ID:            openai.GPT4Dot1Nano,
			Name:          "GPT-4.1 Nano",
			Description:   "GPT-4.1 最小版,低成本高速度",
			ContextLength: 1047576,
			Modality:      "text+image->text",
			Provider:      "openai",
		},
		// GPT-4o 系列
		{
			ID:            openai.GPT4o,
			Name:          "GPT-4o",
			Description:   "GPT-4 Omni 多模态模型",
			ContextLength: 128000,
			Modality:      "text+image->text",
			Provider:      "openai",
		},
		{
			ID:            openai.GPT4oMini,
			Name:          "GPT-4o Mini",
			Description:   "GPT-4o 轻量版,性价比高",
			ContextLength: 128000,
			Modality:      "text+image->text",
			Provider:      "openai",
		},
		// O 系列推理模型
		{
			ID:            openai.O4Mini,
			Name:          "O4 Mini",
			Description:   "O4 轻量推理模型",
			ContextLength: 200000,
			Modality:      "text+image->text",
			Provider:      "openai",
		},
		{
			ID:            openai.O3,
			Name:          "O3",
			Description:   "O3 推理模型,强大逻辑能力",
			ContextLength: 200000,
			Modality:      "text+image->text",
			Provider:      "openai",
		},
		{
			ID:            openai.O3Mini,
			Name:          "O3 Mini",
			Description:   "O3 轻量推理模型",
			ContextLength: 200000,
			Modality:      "text+image->text",
			Provider:      "openai",
		},
		{
			ID:            openai.O1,
			Name:          "O1",
			Description:   "O1 推理模型,深度思考",
			ContextLength: 200000,
			Modality:      "text+image->text",
			Provider:      "openai",
		},
		{
			ID:            openai.O1Mini,
			Name:          "O1 Mini",
			Description:   "O1 轻量推理模型",
			ContextLength: 128000,
			Modality:      "text+image->text",
			Provider:      "openai",
		},
		// GPT-4 Turbo
		{
			ID:            openai.GPT4Turbo,
			Name:          "GPT-4 Turbo",
			Description:   "GPT-4 Turbo 版本",
			ContextLength: 128000,
			Modality:      "text+image->text",
			Provider:      "openai",
		},
	},
}

// OpenAIModels OpenAI 模型 ID 列表 (兼容旧接口)
var OpenAIModels = OpenAIProvider.GetModelIDs()
