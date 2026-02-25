package models

// ZhipuProvider 智谱 GLM 提供商信息
// 数据来源：https://docs.bigmodel.cn/cn/guide/models/text/glm-5
var ZhipuProvider = ProviderInfo{
	ID:          "zhipu",
	Name:        "智谱 AI (Zhipu GLM)",
	BaseURL:     "https://open.bigmodel.cn/api/paas/v4",
	RegisterURL: "https://bigmodel.cn/usercenter/apikeys",
	DocsURL:     "https://docs.bigmodel.cn",
	Models: []ModelInfo{
		// GLM-5 (2026-02 最新旗舰)
		{
			ID:              "glm-5",
			Name:            "GLM-5",
			Description:     "智谱新一代旗舰，MoE 架构 744B 参数，Agentic Engineering | 200k 上下文 (2026-02)",
			ContextLength:   200000,
			Modality:        "text->text",
			Provider:        "zhipu",
			SupportThinking: true,
			SupportTools:    true,
		},
		// GLM-4.7 (2025-12)
		{
			ID:              "glm-4.7",
			Name:            "GLM-4.7",
			Description:     "GLM-4.7 编程与复杂推理旗舰，支持深度思考 | 200k 上下文 (2025-12)",
			ContextLength:   200000,
			Modality:        "text->text",
			Provider:        "zhipu",
			SupportThinking: true,
			SupportTools:    true,
		},
		{
			ID:           "glm-4.7-flash",
			Name:         "GLM-4.7 Flash",
			Description:  "GLM-4.7 快速版，低延迟高效推理 | 200k 上下文",
			ContextLength: 200000,
			Modality:     "text->text",
			Provider:     "zhipu",
			SupportTools: true,
		},
		// GLM-4.6 (2025-09)
		{
			ID:           "glm-4.6",
			Name:         "GLM-4.6",
			Description:  "GLM-4.6 前沿 MoE 模型，355B 参数，MIT 许可开源 | 200k 上下文 (2025-09)",
			ContextLength: 200000,
			Modality:     "text->text",
			Provider:     "zhipu",
			SupportTools: true,
		},
		// GLM-4.5 (2025-07)
		{
			ID:           "glm-4.5",
			Name:         "GLM-4.5",
			Description:  "GLM-4.5 工具调用与代码专项，MoE 355B | 128k 上下文 (2025-07)",
			ContextLength: 131072,
			Modality:     "text->text",
			Provider:     "zhipu",
			SupportTools: true,
		},
		{
			ID:           "glm-4.5-air",
			Name:         "GLM-4.5 Air",
			Description:  "GLM-4.5 轻量版，106B 参数高效版本 | 128k 上下文",
			ContextLength: 131072,
			Modality:     "text->text",
			Provider:     "zhipu",
			SupportTools: true,
		},
		// GLM-4 系列 (稳定版)
		{
			ID:           "glm-4-plus",
			Name:         "GLM-4 Plus",
			Description:  "GLM-4 增强版，复杂任务综合能力强",
			ContextLength: 128000,
			Modality:     "text->text",
			Provider:     "zhipu",
			SupportTools: true,
		},
		{
			ID:           "glm-4",
			Name:         "GLM-4",
			Description:  "GLM-4 标准版，能力全面均衡",
			ContextLength: 128000,
			Modality:     "text->text",
			Provider:     "zhipu",
			SupportTools: true,
		},
		{
			ID:           "glm-4-long",
			Name:         "GLM-4 Long",
			Description:  "GLM-4 超长上下文版，适合长文本分析 | 1M 上下文",
			ContextLength: 1000000,
			Modality:     "text->text",
			Provider:     "zhipu",
			SupportTools: true,
		},
		{
			ID:           "glm-4-flash",
			Name:         "GLM-4 Flash",
			Description:  "GLM-4 快速版，低延迟低成本",
			ContextLength: 128000,
			Modality:     "text->text",
			Provider:     "zhipu",
			SupportTools: true,
		},
		// GLM-4V 视觉系列
		{
			ID:           "glm-4v-plus",
			Name:         "GLM-4V Plus",
			Description:  "GLM-4V 增强视觉版，图文理解最强",
			ContextLength: 8192,
			Modality:     "text+image->text",
			Provider:     "zhipu",
			SupportTools: true,
		},
		{
			ID:       "glm-4v",
			Name:     "GLM-4V",
			Description: "GLM-4 视觉模型，支持图像理解",
			ContextLength: 8192,
			Modality: "text+image->text",
			Provider: "zhipu",
		},
	},
}

// ZhipuModels Zhipu 模型 ID 列表 (兼容旧接口)
var ZhipuModels = ZhipuProvider.GetModelIDs()
