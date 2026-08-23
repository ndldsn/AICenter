package ai

import (
	"context"
	"encoding/json"
	"io"
)

// ProviderType is the API dialect a provider speaks.
type ProviderType string

const (
	ProviderOpenAICompatible ProviderType = "openai-compatible"
	ProviderAnthropic        ProviderType = "anthropic"
	ProviderGemini           ProviderType = "gemini"
	ProviderOllama           ProviderType = "ollama"
	ProviderDeepSeek         ProviderType = "deepseek"
	ProviderMock             ProviderType = "mock" // development e2e only
)

// Config is the provider dial-in (URL + key).
type Config struct {
	BaseURL string
	APIKey  string
}

// ChatRequest is a single chat-completion invocation.
type ChatRequest struct {
	Model     string
	Messages  []Message
	Stream    bool
	MaxTokens int
	Streamer  SSEStreamer
}

// Message is a chat message.
type Message struct {
	Role    string // "system", "user", "assistant"
	Content string
}

// Model mirrors models.AIModel (kept here so ai package doesn't depend on
// models).
type Model struct {
	ID             string `json:"id"`
	ProviderID     string `json:"provider_id"`
	Name           string `json:"name"`
	ModelID        string `json:"model_id"`
	ModelType      string `json:"model_type"`
	MaxTokens      int    `json:"max_tokens"`
	SupportsStream bool   `json:"supports_stream"`
	SupportsTools  bool   `json:"supports_tools"`
	IsEnabled      bool   `json:"is_enabled"`
	IsDefault      bool   `json:"is_default"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// Client is the provider-neutral abstraction every handler/service calls into.
type Client interface {
	// ListModels returns the models exposed by this provider (usually via
	// GET /v1/models or similar).
	ListModels(ctx context.Context) ([]Model, error)

	// ChatCompletion performs a chat completion. When Stream is true the
	// implementation should push delta chunks via the streamer (SSE).
	ChatCompletion(ctx context.Context, req ChatRequest) error
}

// SSEStreamer is the write side of the SSE bridge between a provider and the
// HTTP response. Implementations write `data:` lines; nil means buffering.
type SSEStreamer interface {
	PushDelta(chunk string) error
	PushUsage(promptTokens, completionTokens int) error
	Flush() error
}

// StreamEvents wraps an io.Writer with an SSE helper so the chat handler can
// call streamer.PushDelta(chunk) instead of re-formatting "data: " strings.
type StreamEvents struct {
	w   io.Writer
	buf []byte
}

func NewStreamEvents(w io.Writer) *StreamEvents {
	return &StreamEvents{w: w}
}

func (s *StreamEvents) PushDelta(chunk string) error {
	delta := map[string]string{"content": chunk}
	payload, err := json.Marshal(delta)
	if err != nil {
		return err
	}
	_, err = s.w.Write([]byte("data: " + string(payload) + "\n\n"))
	return err
}

func (s *StreamEvents) PushUsage(promptTokens, completionTokens int) error {
	usage := map[string]int{
		"prompt_tokens":     promptTokens,
		"completion_tokens": completionTokens,
	}
	usageJSON, err := json.Marshal(usage)
	if err != nil {
		return err
	}
	_, err = s.w.Write([]byte("data: {\"usage\":" + string(usageJSON) + "}\n\n"))
	return err
}

func (s *StreamEvents) Flush() error {
	if fw, ok := s.w.(interface{ Flush() error }); ok {
		return fw.Flush()
	}
	return nil
}
