package models

// DoubaoProvider 火山引擎豆包提供商信息
var DoubaoProvider = ProviderInfo{
	ID:          "doubao",
	Name:        "火山引擎豆包 (Doubao)",
	BaseURL:     "https://ark.cn-beijing.volces.com/api/v3",
	RegisterURL: "https://console.volcengine.com/ark/region:ark+cn-beijing/apiKey",
	DocsURL:     "https://www.volcengine.com/docs/82379",
	Models: []ModelInfo{
		// Doubao Pro 系列
		{
			ID:            "doubao-1-5-pro-32k-250115",
			Name:          "Doubao 1.5 Pro 32K",
			Description:   "豆包最新 1.5 Pro 版本,综合能力强劲",
			ContextLength: 32768,
			Modality:      "text->text",
			Provider:        "doubao",
			SupportTools:    true,
		},
		{
			ID:            "doubao-pro-256k",
			Name:          "Doubao Pro 256K",
			Description:   "豆包 Pro 超长上下文版本,支持 256K",
			ContextLength: 262144,
			Modality:      "text->text",
			Provider:        "doubao",
			SupportTools:    true,
		},
		{
			ID:            "doubao-pro-128k",
			Name:          "Doubao Pro 128K",
			Description:   "豆包 Pro 长上下文版本",
			ContextLength: 131072,
			Modality:      "text->text",
			Provider:        "doubao",
			SupportTools:    true,
		},
		{
			ID:            "doubao-pro-32k",
			Name:          "Doubao Pro 32K",
			Description:   "豆包 Pro 标准版本,能力均衡",
			ContextLength: 32768,
			Modality:      "text->text",
			Provider:        "doubao",
			SupportTools:    true,
		},
		{
			ID:            "doubao-pro-4k",
			Name:          "Doubao Pro 4K",
			Description:   "豆包 Pro 短上下文版本,速度快",
			ContextLength: 4096,
			Modality:      "text->text",
			Provider:        "doubao",
			SupportTools:    true,
		},
		// Doubao Lite 系列
		{
			ID:            "doubao-lite-128k",
			Name:          "Doubao Lite 128K",
			Description:   "豆包 Lite 长上下文版本,成本更低",
			ContextLength: 131072,
			Modality:      "text->text",
			Provider:        "doubao",
			SupportTools:    true,
		},
		{
			ID:            "doubao-lite-32k",
			Name:          "Doubao Lite 32K",
			Description:   "豆包 Lite 标准版本,低成本",
			ContextLength: 32768,
			Modality:      "text->text",
			Provider:        "doubao",
			SupportTools:    true,
		},
		{
			ID:            "doubao-lite-4k",
			Name:          "Doubao Lite 4K",
			Description:   "豆包 Lite 快速版本,极低成本",
			ContextLength: 4096,
			Modality:      "text->text",
			Provider:        "doubao",
			SupportTools:    true,
		},
		// Doubao Vision 系列
		{
			ID:            "doubao-vision-pro-32k",
			Name:          "Doubao Vision Pro 32K",
			Description:   "豆包视觉 Pro 模型,支持图像理解",
			ContextLength: 32768,
			Modality:      "text+image->text",
			Provider:        "doubao",
			SupportTools:    true,
		},
		{
			ID:            "doubao-vision-lite-32k",
			Name:          "Doubao Vision Lite 32K",
			Description:   "豆包视觉 Lite 模型,低成本图像理解",
			ContextLength: 32768,
			Modality:      "text+image->text",
			Provider:        "doubao",
			SupportTools:    true,
		},
		// Doubao 角色扮演
		{
			ID:            "doubao-character-pro-32k",
			Name:          "Doubao Character Pro 32K",
			Description:   "豆包角色扮演模型,适合 NPC 和虚拟人场景",
			ContextLength: 32768,
			Modality:      "text->text",
			Provider:        "doubao",
			SupportTools:    true,
		},
	},
}

// DoubaoModels Doubao 模型 ID 列表 (兼容旧接口)
var DoubaoModels = DoubaoProvider.GetModelIDs()
