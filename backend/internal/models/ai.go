package models

// ProviderType identifies the API dialect a provider speaks.
type ProviderType string

const (
	ProviderOpenAICompatible ProviderType = "openai-compatible" // OpenAI, DeepSeek, Ollama, etc.
	ProviderAnthropic        ProviderType = "anthropic"
	ProviderGemini           ProviderType = "gemini" // reserved; not yet implemented
)

// AIProvider is a configured LLM provider (OpenAI, DeepSeek, Ollama, ...).
type AIProvider struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	DisplayName string       `json:"display_name"`
	BaseURL     string       `json:"base_url"`
	APIKeyEnc   string       `json:"-"`            // encrypted at rest, never serialized
	APIKeyHint  string       `json:"api_key_hint"` // e.g. "sk-...abcd" for the UI
	APIType     ProviderType `json:"api_type"`
	IsEnabled   bool         `json:"is_enabled"`
	IsDefault   bool         `json:"is_default"`
	CreatedAt   string       `json:"created_at"`
	UpdatedAt   string       `json:"updated_at"`
}

// AIModel is a model endpoint exposed by a provider.
type AIModel struct {
	ID             string `json:"id"`
	ProviderID     string `json:"provider_id"`
	Name           string `json:"name"`
	ModelID        string `json:"model_id"`
	ModelType      string `json:"model_type"` // chat | embedding | image
	MaxTokens      int    `json:"max_tokens"`
	SupportsStream bool   `json:"supports_stream"`
	SupportsTools  bool   `json:"supports_tools"`
	IsEnabled      bool   `json:"is_enabled"`
	IsDefault      bool   `json:"is_default"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}
