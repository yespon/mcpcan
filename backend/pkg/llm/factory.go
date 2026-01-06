package llm

import (
	"fmt"
)

// Factory defines the function signature for creating a provider
type Factory func(config ProviderConfig) Provider

var (
	// providers registry
	providers = make(map[ProviderType]Factory)
)

// RegisterProvider registers a new provider factory
func RegisterProvider(typ ProviderType, factory Factory) {
	providers[typ] = factory
}

// NewProvider creates a new LLM provider instance
func NewProvider(typ ProviderType, config ProviderConfig) (Provider, error) {
	factory, ok := providers[typ]
	if !ok {
		return nil, fmt.Errorf("unsupported provider type: %s", typ)
	}
	return factory(config), nil
}
