package models

// OllamaProvider Ollama 本地推理提供商信息
// Ollama 是本地部署服务,模型列表取决于用户安装
var OllamaProvider = ProviderInfo{
	ID:          "ollama",
	Name:        "Ollama",
	BaseURL:     "http://localhost:11434/v1",
	RegisterURL: "",
	DocsURL:     "https://ollama.com/library",
	Models: []ModelInfo{
		// 常见支持 Function Calling 的模型
		{
			ID:            "llama3.3",
			Name:          "Llama 3.3",
			Description:   "Meta Llama 3.3,支持工具调用",
			ContextLength: 128000,
			Modality:      "text->text",
			Provider:      "ollama",
		},
		{
			ID:            "qwen2.5",
			Name:          "Qwen 2.5",
			Description:   "Qwen 2.5 本地版,支持工具调用",
			ContextLength: 32000,
			Modality:      "text->text",
			Provider:      "ollama",
		},
		{
			ID:            "mistral",
			Name:          "Mistral",
			Description:   "Mistral 本地版",
			ContextLength: 32000,
			Modality:      "text->text",
			Provider:      "ollama",
		},
	},
}

// OllamaModels Ollama 模型 ID 列表
var OllamaModels = OllamaProvider.GetModelIDs()
