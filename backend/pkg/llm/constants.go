package llm

// constants.go
// 工具函数：通过 Model ID 查询能力信息。
// 各厂商模型列表（XxxModels）已废弃——请直接使用 models.XxxProvider.GetModelIDs()
// 或 models.GetAllModels()。

import (
	"github.com/kymo-mcp/mcpcan/pkg/llm/models"
)

// GetAllModelIDs 返回所有支持的模型 ID 列表
func GetAllModelIDs() []string {
	var ids []string
	for _, m := range models.GetAllModels() {
		ids = append(ids, m.ID)
	}
	return ids
}

// GetModelInfo 根据 Model ID 返回完整的模型能力信息
func GetModelInfo(modelID string) *models.ModelInfo {
	return models.GetModelByID(modelID)
}

// IsVisionSupported 检查模型是否支持图片输入
func IsVisionSupported(modelID string) bool {
	info := models.GetModelByID(modelID)
	if info != nil {
		return info.IsMultimodal()
	}
	return false
}

// IsThinkingSupported 检查模型是否支持深度推理（o1、R1 等）
func IsThinkingSupported(modelID string) bool {
	info := models.GetModelByID(modelID)
	if info != nil {
		return info.SupportThinking
	}
	return false
}

// IsToolsSupported 检查模型是否支持 Function Calling
func IsToolsSupported(modelID string) bool {
	info := models.GetModelByID(modelID)
	if info != nil {
		return info.SupportTools
	}
	return false
}

// IsSystemPromptSupported 检查模型是否支持自定义 System Prompt
func IsSystemPromptSupported(modelID string) bool {
	info := models.GetModelByID(modelID)
	if info != nil {
		return info.SupportSystemPrompt
	}
	return true // 默认兼容
}

// IsTemperatureSupported 检查模型是否支持 Temperature 调节
func IsTemperatureSupported(modelID string) bool {
	info := models.GetModelByID(modelID)
	if info != nil {
		return info.SupportTemperature
	}
	return true // 默认兼容
}
