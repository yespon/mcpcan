package models

// QwenProvider 阿里通义千问提供商信息
var QwenProvider = ProviderInfo{
	ID:          "qwen",
	Name:        "阿里通义千问 (Qwen)",
	BaseURL:     "https://dashscope.aliyuncs.com/compatible-mode/v1",
	RegisterURL: "https://dashscope.console.aliyun.com/apiKey",
	DocsURL:     "https://help.aliyun.com/zh/model-studio",
	Models: []ModelInfo{
		// Qwen3 系列 - 最新一代
		{
			ID:            "qwen3-235b-a22b",
			Name:          "Qwen3 235B A22B",
			Description:   "Qwen3 旗舰 MoE 模型,235B 总参数,22B 激活参数,顶级推理能力",
			ContextLength: 262144,
			Modality:      "text->text",
			Provider:        "qwen",
			SupportTools:    true,
		},
		{
			ID:            "qwen3-32b",
			Name:          "Qwen3 32B",
			Description:   "Qwen3 32B 密集模型,支持复杂推理和对话",
			ContextLength: 131072,
			Modality:      "text->text",
			Provider:        "qwen",
			SupportTools:    true,
		},
		{
			ID:            "qwen3-14b",
			Name:          "Qwen3 14B",
			Description:   "Qwen3 14B 密集模型,平衡性能和效率",
			ContextLength: 131072,
			Modality:      "text->text",
			Provider:        "qwen",
			SupportTools:    true,
		},
		{
			ID:            "qwen3-8b",
			Name:          "Qwen3 8B",
			Description:   "Qwen3 8B 轻量模型,适合低延迟场景",
			ContextLength: 131072,
			Modality:      "text->text",
			Provider:        "qwen",
			SupportTools:    true,
		},
		{
			ID:            "qwen3-30b-a3b",
			Name:          "Qwen3 30B A3B",
			Description:   "Qwen3 MoE 模型,30B 总参数 3B 激活,高效推理",
			ContextLength: 131072,
			Modality:      "text->text",
			Provider:        "qwen",
			SupportTools:    true,
		},
		// Qwen3 Coder 系列
		{
			ID:            "qwen3-coder-480b-a35b",
			Name:          "Qwen3 Coder 480B A35B",
			Description:   "Qwen3 代码专用旗舰模型,480B 总参数,35B 激活,顶级代码能力",
			ContextLength: 1048576,
			Modality:      "text->text",
			Provider:        "qwen",
			SupportTools:    true,
		},
		{
			ID:            "qwen3-coder-30b-a3b",
			Name:          "Qwen3 Coder 30B A3B",
			Description:   "Qwen3 代码专用 MoE 模型,高效代码生成",
			ContextLength: 256000,
			Modality:      "text->text",
			Provider:        "qwen",
			SupportTools:    true,
		},
		// Qwen3 VL 系列 - 视觉语言
		{
			ID:            "qwen3-vl-8b",
			Name:          "Qwen3 VL 8B",
			Description:   "Qwen3 视觉语言模型,支持图像理解",
			ContextLength: 256000,
			Modality:      "text+image->text",
			Provider:        "qwen",
			SupportTools:    true,
		},
		// Qwen 通用系列 (基于 Qwen2.5)
		{
			ID:            "qwen-turbo",
			Name:          "Qwen Turbo",
			Description:   "快速响应模型,低成本高速度,适合简单任务",
			ContextLength: 1000000,
			Modality:      "text->text",
			Provider:        "qwen",
			SupportTools:    true,
		},
		{
			ID:            "qwen-plus",
			Name:          "Qwen Plus",
			Description:   "能力均衡模型,性价比高",
			ContextLength: 131072,
			Modality:      "text->text",
			Provider:        "qwen",
			SupportTools:    true,
		},
		{
			ID:            "qwen-max",
			Name:          "Qwen Max",
			Description:   "旗舰大模型,复杂任务首选",
			ContextLength: 131072,
			Modality:      "text->text",
			Provider:        "qwen",
			SupportTools:    true,
		},
		{
			ID:            "qwen-long",
			Name:          "Qwen Long",
			Description:   "长文本专用模型,支持 1M 上下文",
			ContextLength: 1000000,
			Modality:      "text->text",
			Provider:        "qwen",
			SupportTools:    true,
		},
		// Qwen2.5 系列
		{
			ID:            "qwen2.5-7b-instruct",
			Name:          "Qwen2.5 7B Instruct",
			Description:   "Qwen2.5 指令微调模型,多语言支持",
			ContextLength: 131072,
			Modality:      "text->text",
			Provider:        "qwen",
			SupportTools:    true,
		},
	},
}

// QwenModels Qwen 模型 ID 列表 (兼容旧接口)
var QwenModels = QwenProvider.GetModelIDs()
