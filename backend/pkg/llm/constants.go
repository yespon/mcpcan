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

// GetModelInfo returns the full model info including capabilities (multimodal, thinking, tools)
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
