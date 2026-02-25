package models

// OllamaProvider Ollama 本地推理提供商信息
// Ollama 是本地部署服务,模型列表取决于用户安装
var OllamaProvider = ProviderInfo{
	ID:          "ollama",
	Name:        "Ollama",
	BaseURL:     "http://localhost:11434/v1",
	RegisterURL: "",
	DocsURL:     "https://ollama.com/library",
	Models: []ModelInfo{},
}

// OllamaModels Ollama 模型 ID 列表
var OllamaModels = OllamaProvider.GetModelIDs()
