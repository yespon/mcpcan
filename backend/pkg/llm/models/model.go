package models

// 图片 MIME 类型常量
const (
	MimeJPEG = "image/jpeg"
	MimePNG  = "image/png"
	MimeWEBP = "image/webp"
	MimeGIF  = "image/gif"
	MimeHEIC = "image/heic"
	MimeBMP  = "image/bmp"
)

// 文档 MIME 类型常量
const (
	MimePDF  = "application/pdf"
	MimeTXT  = "text/plain"
	MimeMD   = "text/markdown"
	MimeDOCX = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	MimeDOC  = "application/msword"
	MimeXLSX = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	MimeXLS  = "application/vnd.ms-excel"
	MimePPTX = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	MimePPT  = "application/vnd.ms-powerpoint"
	MimeCSV  = "text/csv"
)

// 常用 MIME 类型组合（便于 providers_gen.go 引用）
var (
	// CommonImageTypes 通用图片类型 (jpeg/png/webp/gif)
	CommonImageTypes = []string{MimeJPEG, MimePNG, MimeWEBP, MimeGIF}
	// ExtendedImageTypes 扩展图片类型 (含 heic)
	ExtendedImageTypes = []string{MimeJPEG, MimePNG, MimeWEBP, MimeGIF, MimeHEIC}
	// DoubaoImageTypes 豆包支持的图片类型 (含 bmp)
	DoubaoImageTypes = []string{MimeJPEG, MimePNG, MimeWEBP, MimeGIF, MimeBMP}
	// BasicDocTypes 基础文档类型 (pdf/txt)
	BasicDocTypes = []string{MimePDF, MimeTXT}
	// ExtendedDocTypes 扩展文档类型 (pdf/txt/md/docx/xlsx/csv)
	ExtendedDocTypes = []string{MimePDF, MimeTXT, MimeMD, MimeDOCX, MimeXLSX, MimeCSV}
	// DoubaoDocTypes 豆包支持的完整文档类型
	DoubaoDocTypes = []string{MimePDF, MimeTXT, MimeCSV, MimeDOCX, MimeDOC, MimeXLSX, MimeXLS, MimePPTX, MimePPT, MimeMD}
)

// MCP 测试工具建议限制（比厂商官方更保守）
// 设计原则：作为测试工具关注的是能力验证而非大文件传输，
// 限制更严可减少 token 消耗、降低测试复杂度
const (
	// MCPMaxImageSize MCP 工具推荐图片大小上限 (5MB，官方通常 10-20MB)
	MCPMaxImageSize int64 = 5 * 1024 * 1024
	// MCPMaxImageCount MCP 工具推荐单次最多图片数 (1 张，足够测试视觉能力)
	MCPMaxImageCount = 1
	// MCPMaxDocumentSize MCP 工具推荐文档大小上限 (5MB，官方通常 10-20MB)
	MCPMaxDocumentSize int64 = 5 * 1024 * 1024
	// MCPMaxDocumentCount MCP 工具推荐单次最多文档数 (1 个，足够测试文档解析)
	MCPMaxDocumentCount = 1
)

// ModelInfo 模型详细信息
type ModelInfo struct {
	ID                  string   `json:"id"`                  // 模型 ID (API 调用名称)
	Name                string   `json:"name"`                // 显示名称
	Description         string   `json:"description"`         // 模型描述
	ContextLength       int      `json:"contextLength"`       // 上下文长度（token）
	Provider            string   `json:"provider"`            // 提供商 ID
	SupportThinking     bool     `json:"supportThinking"`     // 是否支持深度思考 (o1, R1 等推理模型)
	SupportTools        bool     `json:"supportTools"`        // 是否支持工具调用 (Function Calling)
	SupportSystemPrompt bool     `json:"supportSystemPrompt"` // 是否支持自定义系统提示词
	SupportTemperature  bool     `json:"supportTemperature"`  // 是否支持调整 Temperature (推理模型通常不支持)

	// 图片输入能力
	SupportsVision  bool     `json:"supportsVision"`  // 是否支持图片输入（多模态视觉）
	ImageMimeTypes  []string `json:"imageMimeTypes"`  // 支持的图片 MIME 类型列表，如 ["image/jpeg","image/png"]
	MaxImageSize    int64    `json:"maxImageSize"`    // 单张图片最大字节数（0=不限）
	MaxImageCount   int      `json:"maxImageCount"`   // 单次请求最多图片数量（0=不限）

	// 文档附件能力
	SupportsDocument  bool     `json:"supportsDocument"`  // 是否支持文档/文件附件
	DocumentMimeTypes []string `json:"documentMimeTypes"` // 支持的文档 MIME 类型列表，如 ["application/pdf","text/plain"]
	MaxDocumentSize   int64    `json:"maxDocumentSize"`   // 单份文档最大字节数（0=不限）
	MaxDocumentCount  int      `json:"maxDocumentCount"`  // 单次请求最多文档数量（0=不限）
}

// IsMultimodal returns true if the model supports image input
func (m ModelInfo) IsMultimodal() bool {
	return m.SupportsVision
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
