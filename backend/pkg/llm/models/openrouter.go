package models

// OpenRouterProvider OpenRouter 提供商信息
// OpenRouter 是模型聚合服务,不预设具体模型列表
var OpenRouterProvider = ProviderInfo{
	ID:          "openrouter",
	Name:        "OpenRouter",
	BaseURL:     "https://openrouter.ai/api/v1",
	RegisterURL: "https://openrouter.ai/keys",
	DocsURL:     "https://openrouter.ai/docs",
	Models:      []ModelInfo{}, // 动态模型,由用户指定
}

// OpenRouterModels OpenRouter 模型 ID 列表
var OpenRouterModels = OpenRouterProvider.GetModelIDs()
