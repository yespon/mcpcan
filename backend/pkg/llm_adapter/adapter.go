package llm_adapter

// llm_adapter/adapter.go
// Adapter 层入口：因 types.go 已通过 type alias 直接复用 pkg/llm 层类型，
// 这里不再需要任何字段转换，直接透传请求和响应。

import (
	orig_llm "github.com/kymo-mcp/mcpcan/pkg/llm"
)

// NewProvider 创建一个新的 LLM Provider 实例
func NewProvider(typ ProviderType, config ProviderConfig) (Provider, error) {
	p, err := orig_llm.NewProvider(typ, config)
	if err != nil {
		return nil, err
	}
	return p, nil
}
