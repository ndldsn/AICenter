package ai

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type anthropicClient struct {
	baseURL string
	apiKey  string
}

var ErrAnthropicNotImplemented = errors.New("anthropic provider not yet implemented")

func NewAnthropicClient(cfg Config) Client {
	return &anthropicClient{baseURL: cfg.BaseURL, apiKey: cfg.APIKey}
}

func (c *anthropicClient) ListModels(ctx context.Context) ([]Model, error) {
	return nil, ErrAnthropicNotImplemented
}

func (c *anthropicClient) ChatCompletion(ctx context.Context, req ChatRequest) error {
	// Placeholder so that the handler chain exists end-to-end but fails loudly
	// if a user enables an Anthropic provider before we wire the real API.
	time.Sleep(200 * time.Millisecond)
	return fmt.Errorf("%w", ErrAnthropicNotImplemented)
}

// DefaultProviderConcurrency caps the number of simultaneous outbound calls to
// any single provider. This is a security/DoS guard: it prevents credential
// abuse and 429 storms from bursts, and keeps one hot provider from starving
// the backend.
const DefaultProviderConcurrency = 4

// Factory builds the right client for the given provider type.
type Factory struct{}

func NewFactory() *Factory { return &Factory{} }

func (f *Factory) Build(p ProviderType, cfg Config) Client {
	var c Client
	switch p {
	case ProviderOpenAICompatible:
		c = NewOpenAICompatClient(cfg)
	case ProviderAnthropic:
		c = NewAnthropicClient(cfg)
	case ProviderMock:
		c = &MockClient{}
	default:
		// Gemini and any unknown type fall back to mock-ish behaviour so
		// seed data does not break the app.
		c = &fallbackClient{cfg: cfg, kind: p}
	}
	// Wrap every client with a per-provider concurrency limiter.
	return NewLimited(c, DefaultProviderConcurrency)
}

type fallbackClient struct {
	cfg  Config
	kind ProviderType
}

func (f *fallbackClient) ListModels(ctx context.Context) ([]Model, error) {
	return []Model{{ID: "placeholder", ModelID: "placeholder", Name: "placeholder", IsEnabled: true}}, nil
}

func (f *fallbackClient) ChatCompletion(ctx context.Context, req ChatRequest) error {
	return fmt.Errorf("provider type %q not yet supported", f.kind)
}
