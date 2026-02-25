package models

// AllProviders 所有支持的提供商列表
var AllProviders = []ProviderInfo{
	// Core Providers
	OpenAIProvider,
	AnthropicProvider,
	DeepSeekProvider,
	GoogleProvider,
	MistralProvider,
	XAIProvider,
	// Chinese Providers
	QwenProvider,
	DoubaoProvider,
	ZhipuProvider,
	MoonshotProvider,
	// Aggregator/Proxy Providers
	OpenRouterProvider,
	OllamaProvider,
}

// GetProviderByID 根据 ID 获取提供商信息
func GetProviderByID(id string) *ProviderInfo {
	for _, p := range AllProviders {
		if p.ID == id {
			return &p
		}
	}
	return nil
}

// GetModelByID 根据模型 ID 获取模型信息
func GetModelByID(modelID string) *ModelInfo {
	for _, p := range AllProviders {
		for _, m := range p.Models {
			if m.ID == modelID {
				return &m
			}
		}
	}
	return nil
}

// GetAllModels 获取所有模型列表
func GetAllModels() []ModelInfo {
	var models []ModelInfo
	for _, p := range AllProviders {
		models = append(models, p.Models...)
	}
	return models
}

// GetProviderForModel 根据模型 ID 获取对应的提供商
func GetProviderForModel(modelID string) *ProviderInfo {
	for _, p := range AllProviders {
		for _, m := range p.Models {
			if m.ID == modelID {
				return &p
			}
		}
	}
	return nil
}
