package models

// QwenProvider 阿里通义千问提供商信息
var QwenProvider = ProviderInfo{
	ID:          "qwen",
	Name:        "阿里通义千问 (Qwen)",
	BaseURL:     "https://dashscope.aliyuncs.com/compatible-mode/v1",
	RegisterURL: "https://dashscope.console.aliyun.com/apiKey",
	DocsURL:     "https://help.aliyun.com/zh/model-studio",
	Models: []ModelInfo{

		{
			ID:            "qwen-max",
			Name:          "Qwen Max",
			Description:   "通义千问旗舰模型，能力最强",
			ContextLength: 32768,
			Modality:      "text->text",
			Provider:      "qwen",
			SupportTools:  true,
		},
		{
			ID:            "qwen-plus",
			Name:          "Qwen Plus",
			Description:   "通义千问进阶模型，均衡",
			ContextLength: 131072,
			Modality:      "text->text",
			Provider:      "qwen",
			SupportTools:  true,
		},
		{
			ID:            "qwen-turbo",
			Name:          "Qwen Turbo",
			Description:   "通义千问快速模型，低成本",
			ContextLength: 131072,
			Modality:      "text->text",
			Provider:      "qwen",
			SupportTools:  true,
		},
		{
			ID:            "qwen-long",
			Name:          "Qwen Long",
			Description:   "通义千问长文本模型，支持10M上下文",
			ContextLength: 10000000,
			Modality:      "text->text",
			Provider:      "qwen",
			SupportTools:  true,
		},
		{
			ID:            "qwen2.5-coder-32b-instruct",
			Name:          "Qwen2.5 Coder 32B",
			Description:   "Qwen2.5 代码专用模型",
			ContextLength: 131072,
			Modality:      "text->text",
			Provider:      "qwen",
			SupportTools:  true,
		},
	},
}

// QwenModels Qwen 模型 ID 列表 (兼容旧接口)
var QwenModels = QwenProvider.GetModelIDs()
