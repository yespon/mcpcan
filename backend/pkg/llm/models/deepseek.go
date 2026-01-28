package models

// DeepSeekProvider DeepSeek 提供商信息
var DeepSeekProvider = ProviderInfo{
	ID:          "deepseek",
	Name:        "DeepSeek",
	BaseURL:     "https://api.deepseek.com/v1",
	RegisterURL: "https://platform.deepseek.com/api_keys",
	DocsURL:     "https://api-docs.deepseek.com",
	Models: []ModelInfo{
		// DeepSeek 主力模型
		{
			ID:            "deepseek-chat",
			Name:          "DeepSeek Chat",
			Description:   "DeepSeek V3 对话模型,综合能力强",
			ContextLength: 65536,
			Modality:      "text->text",
			Provider:      "deepseek",
		},
		{
			ID:            "deepseek-coder",
			Name:          "DeepSeek Coder",
			Description:   "DeepSeek 代码专用模型",
			ContextLength: 65536,
			Modality:      "text->text",
			Provider:      "deepseek",
		},
		{
			ID:            "deepseek-reasoner",
			Name:          "DeepSeek Reasoner",
			Description:   "DeepSeek R1 推理模型,深度思考能力",
			ContextLength: 65536,
			Modality:      "text->text",
			Provider:      "deepseek",
		},
		// DeepSeek R1 Distill 系列
		{
			ID:            "deepseek-r1-distill-llama-70b",
			Name:          "DeepSeek R1 Distill Llama 70B",
			Description:   "基于 Llama 70B 的 R1 蒸馏模型",
			ContextLength: 128000,
			Modality:      "text->text",
			Provider:      "deepseek",
		},
		{
			ID:            "deepseek-r1-distill-qwen-32b",
			Name:          "DeepSeek R1 Distill Qwen 32B",
			Description:   "基于 Qwen 32B 的 R1 蒸馏模型",
			ContextLength: 128000,
			Modality:      "text->text",
			Provider:      "deepseek",
		},
	},
}

// DeepSeekModels DeepSeek 模型 ID 列表 (兼容旧接口)
var DeepSeekModels = DeepSeekProvider.GetModelIDs()
