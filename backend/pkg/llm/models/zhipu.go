package models

// ZhipuProvider 智谱 GLM 提供商信息
var ZhipuProvider = ProviderInfo{
	ID:          "zhipu",
	Name:        "智谱 AI (Zhipu GLM)",
	BaseURL:     "https://open.bigmodel.cn/api/paas/v4",
	RegisterURL: "https://bigmodel.cn/usercenter/apikeys",
	DocsURL:     "https://bigmodel.cn/dev/api",
	Models: []ModelInfo{
		// GLM-4.5 系列 - 最新旗舰
		{
			ID:            "glm-4.5",
			Name:          "GLM-4.5",
			Description:   "智谱最新旗舰模型,MoE 架构,支持推理模式",
			ContextLength: 131072,
			Modality:      "text->text",
			Provider:      "zhipu",
		},
		{
			ID:            "glm-4.5-air",
			Name:          "GLM-4.5 Air",
			Description:   "GLM-4.5 轻量版,MoE 架构,高效推理",
			ContextLength: 131072,
			Modality:      "text->text",
			Provider:      "zhipu",
		},
		// GLM-4.7 系列
		{
			ID:            "glm-4.7-flash",
			Name:          "GLM-4.7 Flash",
			Description:   "GLM 快速版本,优化代码和 Agent 能力",
			ContextLength: 200000,
			Modality:      "text->text",
			Provider:      "zhipu",
		},
		// GLM-4 系列
		{
			ID:            "glm-4-32b",
			Name:          "GLM-4 32B",
			Description:   "GLM-4 32B 基础模型,性价比优秀",
			ContextLength: 128000,
			Modality:      "text->text",
			Provider:      "zhipu",
		},
		{
			ID:            "glm-4",
			Name:          "GLM-4",
			Description:   "GLM-4 标准版,能力全面",
			ContextLength: 128000,
			Modality:      "text->text",
			Provider:      "zhipu",
		},
		{
			ID:            "glm-4-plus",
			Name:          "GLM-4 Plus",
			Description:   "GLM-4 增强版,复杂任务更强",
			ContextLength: 128000,
			Modality:      "text->text",
			Provider:      "zhipu",
		},
		{
			ID:            "glm-4-long",
			Name:          "GLM-4 Long",
			Description:   "GLM-4 长文本版,支持超长上下文",
			ContextLength: 1000000,
			Modality:      "text->text",
			Provider:      "zhipu",
		},
		{
			ID:            "glm-4-flash",
			Name:          "GLM-4 Flash",
			Description:   "GLM-4 快速版,低延迟低成本",
			ContextLength: 128000,
			Modality:      "text->text",
			Provider:      "zhipu",
		},
		{
			ID:            "glm-4-air",
			Name:          "GLM-4 Air",
			Description:   "GLM-4 轻量版,免费使用",
			ContextLength: 128000,
			Modality:      "text->text",
			Provider:      "zhipu",
		},
		{
			ID:            "glm-4-airx",
			Name:          "GLM-4 AirX",
			Description:   "GLM-4 AirX 版,更快响应",
			ContextLength: 8192,
			Modality:      "text->text",
			Provider:      "zhipu",
		},
		// GLM-4V 视觉系列
		{
			ID:            "glm-4v",
			Name:          "GLM-4V",
			Description:   "GLM-4 视觉模型,支持图像理解",
			ContextLength: 8192,
			Modality:      "text+image->text",
			Provider:      "zhipu",
		},
		{
			ID:            "glm-4v-plus",
			Name:          "GLM-4V Plus",
			Description:   "GLM-4V 增强版,视觉能力更强",
			ContextLength: 8192,
			Modality:      "text+image->text",
			Provider:      "zhipu",
		},
		// 代码模型
		{
			ID:            "codegeex-4",
			Name:          "CodeGeeX 4",
			Description:   "智谱代码模型,专注代码生成",
			ContextLength: 128000,
			Modality:      "text->text",
			Provider:      "zhipu",
		},
	},
}

// ZhipuModels Zhipu 模型 ID 列表 (兼容旧接口)
var ZhipuModels = ZhipuProvider.GetModelIDs()
