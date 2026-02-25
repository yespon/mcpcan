package models

// DeepSeekProvider DeepSeek 提供商信息
var DeepSeekProvider = ProviderInfo{
	ID:          "deepseek",
	Name:        "DeepSeek",
	BaseURL:     "https://api.deepseek.com/v1",
	RegisterURL: "https://platform.deepseek.com/api_keys",
	DocsURL:     "https://api-docs.deepseek.com",
	Models: []ModelInfo{
		// DeepSeek V3 (Chat)
		{
			ID:              "deepseek-chat",
			Name:            "DeepSeek V3",
			Description:     "DeepSeek V3 旗舰对话模型,综合能力强",
			ContextLength:   65536,
			Modality:        "text->text",
			Provider:        "deepseek",
			SupportThinking: false,
			SupportTools:    true,
		},
		// DeepSeek R1 (Reasoning)
		{
			ID:              "deepseek-reasoner",
			Name:            "DeepSeek R1",
			Description:     "DeepSeek R1 推理模型,深度思考能力",
			ContextLength:   65536,
			Modality:        "text->text",
			Provider:        "deepseek",
			SupportThinking: true,
			SupportTools:    true,
		},
	},
}

// DeepSeekModels DeepSeek 模型 ID 列表 (兼容旧接口)
var DeepSeekModels = DeepSeekProvider.GetModelIDs()
