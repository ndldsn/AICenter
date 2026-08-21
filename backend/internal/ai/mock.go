package ai

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// MockClient streams tokens with a short delay so the SSE path can be tested
// end-to-end without a real LLM. Token text is just a deterministic sequence
// ("Hello", " ", "world", ...) so handler tests can assert on substrings.
type MockClient struct {
	mu      sync.Mutex
	history []Message
}

func (m *MockClient) ListModels(ctx context.Context) ([]Model, error) {
	return []Model{
		{ID: "mock-gpt-4o", ProviderID: "mock", Name: "gpt-4o", ModelID: "gpt-4o", ModelType: "chat", SupportsStream: true, IsEnabled: true, IsDefault: true},
		{ID: "mock-claude-3-5", ProviderID: "mock", Name: "claude-3-5-sonnet", ModelID: "claude-3-5-sonnet", ModelType: "chat", SupportsStream: true, IsEnabled: true},
	}, nil
}

func (m *MockClient) ChatCompletion(ctx context.Context, req ChatRequest) error {
	m.mu.Lock()
	m.history = append(m.history, req.Messages...)
	m.mu.Unlock()

	response := "Hello from the mock provider. "
	if len(req.Messages) > 0 {
		response += fmt.Sprintf("Your last message was: %q. ", req.Messages[len(req.Messages)-1].Content)
	}
	response += "Streaming test successful."

	if req.Stream && req.Streamer != nil {
		// Chunk the response into ~5-byte slices to mimic real streaming.
		chunks := 0
		for i := 0; i < len(response); i += 4 {
			end := i + 4
			if end > len(response) {
				end = len(response)
			}
			if err := req.Streamer.PushDelta(response[i:end]); err != nil {
				return err
			}
			if err := req.Streamer.Flush(); err != nil {
				return err
			}
			chunks++
			time.Sleep(20 * time.Millisecond)
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
		}
		return req.Streamer.PushUsage(42, chunks)
	}

	_ = response
	return nil
}

var ErrMockNotReady = errors.New("mock client not ready")
