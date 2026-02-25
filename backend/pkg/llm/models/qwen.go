package models

// QwenProvider 阿里通义千问提供商信息
var QwenProvider = ProviderInfo{
	ID:          "qwen",
	Name:        "阿里通义千问 (Qwen)",
	BaseURL:     "https://dashscope.aliyuncs.com/compatible-mode/v1",
	RegisterURL: "https://dashscope.console.aliyun.com/apiKey",
	DocsURL:     "https://help.aliyun.com/zh/model-studio",
	Models: []ModelInfo{
		// Qwen3.5 Series (2026-02 最新)
		{
			ID:              "qwen3.5-plus",
			Name:            "Qwen3.5 Plus",
			Description:     "全球最强开源模型，视觉语言 MoE (2026-02)",
			ContextLength:   131072,
			Modality:        "text+image->text",
			Provider:        "qwen",
			SupportThinking: true,
			SupportTools:    true,
		},
		// Qwen3 Series (2025 旗舰)
		{
			ID:              "qwen3-max",
			Name:            "Qwen3 Max",
			Description:     "Qwen3 最强通用模型，1T+ 参数 MoE",
			ContextLength:   131072,
			Modality:        "text->text",
			Provider:        "qwen",
			SupportThinking: true,
			SupportTools:    true,
		},
		{
			ID:              "qwen3-plus",
			Name:            "Qwen3 Plus",
			Description:     "Qwen3 进阶版，思考/非思考双模式",
			ContextLength:   131072,
			Modality:        "text->text",
			Provider:        "qwen",
			SupportThinking: true,
			SupportTools:    true,
		},
		{
			ID:              "qwen3-turbo",
			Name:            "Qwen3 Turbo",
			Description:     "Qwen3 快速版，低成本高并发",
			ContextLength:   131072,
			Modality:        "text->text",
			Provider:        "qwen",
			SupportThinking: true,
			SupportTools:    true,
		},
		// QwQ Reasoning Series
		{
			ID:              "qwq-plus",
			Name:            "QwQ Plus",
			Description:     "千问推理专项模型，深度思考",
			ContextLength:   131072,
			Modality:        "text->text",
			Provider:        "qwen",
			SupportThinking: true,
			SupportTools:    true,
		},
		// Qwen2.5 Legacy Series
		{
			ID:           "qwen-max",
			Name:         "Qwen Max",
			Description:  "通义千问商业旗舰模型",
			ContextLength: 32768,
			Modality:     "text->text",
			Provider:     "qwen",
			SupportTools: true,
		},
		{
			ID:           "qwen2.5-max",
			Name:         "Qwen2.5 Max",
			Description:  "超大规模 MoE，知识与推理领先 (2025-01)",
			ContextLength: 131072,
			Modality:     "text->text",
			Provider:     "qwen",
			SupportTools: true,
		},
		{
			ID:           "qwen-plus",
			Name:         "Qwen Plus",
			Description:  "通义千问进阶模型，均衡",
			ContextLength: 131072,
			Modality:     "text->text",
			Provider:     "qwen",
			SupportTools: true,
		},
		{
			ID:           "qwen-turbo",
			Name:         "Qwen Turbo",
			Description:  "通义千问快速模型，低成本",
			ContextLength: 131072,
			Modality:     "text->text",
			Provider:     "qwen",
			SupportTools: true,
		},
		{
			ID:           "qwen-long",
			Name:         "Qwen Long",
			Description:  "通义千问长文本模型，支持 10M 上下文",
			ContextLength: 10000000,
			Modality:     "text->text",
			Provider:     "qwen",
			SupportTools: true,
		},
		// Qwen Vision Series
		{
			ID:           "qwen-vl-max",
			Name:         "Qwen VL Max",
			Description:  "通义千问视觉旗舰，图文理解最强",
			ContextLength: 32768,
			Modality:     "text+image->text",
			Provider:     "qwen",
			SupportTools: true,
		},
		{
			ID:           "qwen-vl-plus",
			Name:         "Qwen VL Plus",
			Description:  "通义千问视觉增强版",
			ContextLength: 32768,
			Modality:     "text+image->text",
			Provider:     "qwen",
			SupportTools: true,
		},
		// Qwen Coder
		{
			ID:           "qwen2.5-coder-32b-instruct",
			Name:         "Qwen2.5 Coder 32B",
			Description:  "Qwen2.5 代码专用模型",
			ContextLength: 131072,
			Modality:     "text->text",
			Provider:     "qwen",
			SupportTools: true,
		},
	},
}

// QwenModels Qwen 模型 ID 列表 (兼容旧接口)
var QwenModels = QwenProvider.GetModelIDs()
