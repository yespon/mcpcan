package models

// DoubaoProvider 火山引擎豆包提供商信息
var DoubaoProvider = ProviderInfo{
	ID:          "doubao",
	Name:        "火山引擎豆包 (Doubao)",
	BaseURL:     "https://ark.cn-beijing.volces.com/api/v3",
	RegisterURL: "https://console.volcengine.com/ark/region:ark+cn-beijing/apiKey",
	DocsURL:     "https://www.volcengine.com/docs/82379",
	Models: []ModelInfo{

		// 通用豆包模型
		//doubao-seed-1-6-251015
		//doubao-seed-1-6-thinking-250715
		//doubao-seed-1-6-flash-250828
		//doubao-seed-1-6-lite-251015
		//doubao-1-5-pro-32k-250115
		//doubao-1-5-pro-256k-250115
		// doubao-1-5-lite-32k-250115
		// doubao-seed-code-preview-251028
		// doubao-1-5-thinking-pro-250415
		// kimi-k2-thinking-251104
		// glm-4-7-251222
		// deepseek-v3-1-terminus
		// deepseek-r1-250528
		{

			ID:            "glm-4-7-251222",
			Name:          "Doubao-OpenAI-glm-4-7",
			Description:   "豆包开源多模型",
			ContextLength: 200000,
			Modality:      "text+image->text",
			Provider:      "doubao",
			SupportTools:  true,
		},
	},
}

// DoubaoModels Doubao 模型 ID 列表 (兼容旧接口)
var DoubaoModels = DoubaoProvider.GetModelIDs()
