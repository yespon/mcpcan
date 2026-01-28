package models

// ModelInfo 模型详细信息
type ModelInfo struct {
	ID            string `json:"id"`            // 模型 ID (API 调用名称)
	Name          string `json:"name"`          // 显示名称
	Description   string `json:"description"`   // 模型描述
	ContextLength int    `json:"contextLength"` // 上下文长度
	Modality      string `json:"modality"`      // 支持模态 (text->text, text+image->text)
	Provider      string `json:"provider"`      // 提供商 ID
}

// ProviderInfo 提供商信息
type ProviderInfo struct {
	ID          string      `json:"id"`          // 提供商 ID
	Name        string      `json:"name"`        // 显示名称
	BaseURL     string      `json:"baseUrl"`     // API 基础地址
	RegisterURL string      `json:"registerUrl"` // API Key 注册地址
	DocsURL     string      `json:"docsUrl"`     // 文档地址
	Models      []ModelInfo `json:"models"`      // 支持的模型列表
}

// GetModelIDs 获取提供商的所有模型 ID 列表
func (p *ProviderInfo) GetModelIDs() []string {
	ids := make([]string, len(p.Models))
	for i, m := range p.Models {
		ids[i] = m.ID
	}
	return ids
}
