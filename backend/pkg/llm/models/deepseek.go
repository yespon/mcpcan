package models

// DeepSeekProvider DeepSeek 提供商信息
var DeepSeekProvider = ProviderInfo{
	ID:          "deepseek",
	Name:        "DeepSeek",
	BaseURL:     "https://api.deepseek.com/v1",
	RegisterURL: "https://platform.deepseek.com/api_keys",
	DocsURL:     "https://api-docs.deepseek.com",
	Models: []ModelInfo{
		// DeepSeek 主力模型 (支持 Function Calling)
		{
			ID:            "deepseek-chat",
			Name:          "DeepSeek Chat",
			Description:   "DeepSeek V3 对话模型,综合能力强,支持工具调用",
			ContextLength: 65536,
			Modality:      "text->text",
			Provider:      "deepseek",
		},
		{
			ID:            "deepseek-reasoner",
			Name:          "DeepSeek Reasoner",
			Description:   "DeepSeek R1 推理模型,深度思考能力,支持工具调用",
			ContextLength: 65536,
			Modality:      "text->text",
			Provider:      "deepseek",
		},
	},
}

// DeepSeekModels DeepSeek 模型 ID 列表 (兼容旧接口)
var DeepSeekModels = DeepSeekProvider.GetModelIDs()
