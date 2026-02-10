package models

// DoubaoProvider 火山引擎豆包提供商信息
var DoubaoProvider = ProviderInfo{
	ID:          "doubao",
	Name:        "火山引擎豆包 (Doubao)",
	BaseURL:     "https://ark.cn-beijing.volces.com/api/v3",
	RegisterURL: "https://console.volcengine.com/ark/region:ark+cn-beijing/apiKey",
	DocsURL:     "https://www.volcengine.com/docs/82379",
	Models: []ModelInfo{

		{
			ID:            "doubao-pro-32k",
			Name:          "Doubao Pro 32K",
			Description:   "基础效果 - 必须替换为 Endpoint ID",
			ContextLength: 32768,
			Modality:      "text->text",
			Provider:      "doubao",
			SupportTools:  true,
		},
		{
			ID:            "doubao-pro-128k",
			Name:          "Doubao Pro 128K",
			Description:   "长文本效果 - 必须替换为 Endpoint ID",
			ContextLength: 131072,
			Modality:      "text->text",
			Provider:      "doubao",
			SupportTools:  true,
		},
		{
			ID:            "doubao-lite-32k",
			Name:          "Doubao Lite 32K",
			Description:   "低成本效果 - 必须替换为 Endpoint ID",
			ContextLength: 32768,
			Modality:      "text->text",
			Provider:      "doubao",
			SupportTools:  true,
		},
		{
			ID:            "doubao-vision-pro-32k",
			Name:          "Doubao Vision Pro",
			Description:   "视觉理解效果 - 必须替换为 Endpoint ID",
			ContextLength: 32768,
			Modality:      "text+image->text",
			Provider:      "doubao",
			SupportTools:  true,
		},
		{
			ID:            "doubao-character-pro-32k",
			Name:          "Doubao Character",
			Description:   "角色扮演效果 - 必须替换为 Endpoint ID",
			ContextLength: 32768,
			Modality:      "text->text",
			Provider:      "doubao",
			SupportTools:  true,
		},
	},
}

// DoubaoModels Doubao 模型 ID 列表 (兼容旧接口)
var DoubaoModels = DoubaoProvider.GetModelIDs()
