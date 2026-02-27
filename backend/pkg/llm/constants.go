package llm

import (
	"github.com/kymo-mcp/mcpcan/pkg/llm/models"
)

// Re-export Model Lists for backward compatibility
// These lists are populated from the central registry in pkg/llm/models
var (
	DeepSeekModels  = models.DeepSeekProvider.GetModelIDs()
	OpenAIModels    = models.OpenAIProvider.GetModelIDs()
	AnthropicModels = models.AnthropicProvider.GetModelIDs()
	GoogleModels    = models.GoogleProvider.GetModelIDs()
	DoubaoModels    = models.DoubaoProvider.GetModelIDs()
	QwenModels      = models.QwenProvider.GetModelIDs()
	ZhipuModels     = models.ZhipuProvider.GetModelIDs()
	// 新增厂商
	BaiduModels     = models.BaiduProvider.GetModelIDs()
	HunyuanModels   = models.HunyuanProvider.GetModelIDs()
	SparkModels     = models.SparkProvider.GetModelIDs()
	MiniMaxModels   = models.MiniMaxProvider.GetModelIDs()
	Yi01AIModels    = models.Yi01AIProvider.GetModelIDs()
	MoonshotModels  = models.MoonshotProvider.GetModelIDs()
	MistralModels   = models.MistralProvider.GetModelIDs()
	XAIModels       = models.XAIProvider.GetModelIDs()
	CohereModels    = models.CohereProvider.GetModelIDs()
)

// GetAllModels returns all supported model IDs
func GetAllModelIDs() []string {
	var ids []string
	all := models.GetAllModels()
	for _, m := range all {
		ids = append(ids, m.ID)
	}
	return ids
}

// GetModelInfo returns the full model info including capabilities
func GetModelInfo(modelID string) *models.ModelInfo {
	return models.GetModelByID(modelID)
}

// IsVisionSupported checks if the model supports multimodal input (images).
func IsVisionSupported(modelID string) bool {
	info := models.GetModelByID(modelID)
	if info != nil {
		return info.IsMultimodal()
	}
	return false
}

// IsThinkingSupported checks if the model supports deep reasoning (e.g. o1, R1)
func IsThinkingSupported(modelID string) bool {
	info := models.GetModelByID(modelID)
	if info != nil {
		return info.SupportThinking
	}
	return false
}

// IsToolsSupported checks if the model supports function calling
func IsToolsSupported(modelID string) bool {
	info := models.GetModelByID(modelID)
	if info != nil {
		return info.SupportTools
	}
	return false
}

// IsSystemPromptSupported checks if the model supports custom system prompt
func IsSystemPromptSupported(modelID string) bool {
	info := models.GetModelByID(modelID)
	if info != nil {
		return info.SupportSystemPrompt
	}
	return true // Default to true for backward compatibility
}

// IsTemperatureSupported checks if temperature can be adjusted for this model
func IsTemperatureSupported(modelID string) bool {
	info := models.GetModelByID(modelID)
	if info != nil {
		return info.SupportTemperature
	}
	return true // Default to true for backward compatibility
}
