package llm

import (
	"context"
)

// ProviderType defines the type of LLM provider
type ProviderType string

const (
	ProviderOpenAI    ProviderType = "openai"
	ProviderDeepSeek  ProviderType = "deepseek"
	ProviderMoonshot  ProviderType = "moonshot"
	ProviderDoubao    ProviderType = "doubao"    // Doubao (Volcengine)
	ProviderAnthropic ProviderType = "anthropic" // Future support
	ProviderGemini    ProviderType = "gemini"    // Future support
)

// ProviderConfig holds configuration for LLM provider
type ProviderConfig struct {
	BaseURL string
	APIKey  string
}

// Message represents a chat message
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// ToolCall represents a tool call request from LLM
type ToolCall struct {
	Index    int              `json:"index,omitempty"` // Added for streaming aggregation
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction represents the function call details in a tool call (arguments are string JSON)
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Tool represents a tool definition
type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// Function represents the function call details
type Function struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters,omitempty"` // JSON Schema
}

// ChatRequest represents a request to the LLM
type ChatRequest struct {
	Model       string
	Messages    []Message
	Tools       []Tool // Structured tool definition
	Temperature float32
	MaxTokens   int
	Stream      bool
}

// ToolOutput represents the result of a tool execution
type ToolOutput struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Result string `json:"result"`
}

// StreamResponse represents a chunk of response from streaming API
type StreamResponse struct {
	Content     string
	ToolCalls   []ToolCall
	ToolOutputs []ToolOutput // Added for notifying execution results
	Usage       *Usage
	Error       error
}

// Usage represents token usage statistics
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// Provider defines the interface for LLM providers
type Provider interface {
	// StreamChat sends a chat request and returns a stream of responses
	StreamChat(ctx context.Context, req ChatRequest) (<-chan StreamResponse, error)
}
