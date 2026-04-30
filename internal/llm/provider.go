package llm

import (
	"context"
	"fmt"
)

// Provider is the interface for LLM providers with streaming support.
type Provider interface {
	Stream(ctx context.Context, messages []Message, opts GenerateOptions) (<-chan StreamEvent, error)
	Name() string
}

// NewProvider creates a provider by name with the given API key and optional base URL.
func NewProvider(providerName, apiKey, baseURL string) (Provider, error) {
	switch providerName {
	case "anthropic":
		return NewAnthropicProvider(apiKey, baseURL), nil
	case "openai":
		return NewOpenAIProvider(apiKey), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}
}
