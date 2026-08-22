package ai

import (
	"context"
	"errors"
)

// ErrRateLimited is returned when a provider's concurrency cap is saturated.
var ErrRateLimited = errors.New("provider concurrency limit reached; try again shortly")

// Limited wraps a Client with a per-provider concurrency limiter (semaphore).
// This hard-caps outbound calls to a provider so a burst cannot exhaust the
// provider quota or abuse stored credentials, and surfaces bounded back-pressure
// instead of an unbounded backlog.
type Limited struct {
	Client
	sem chan struct{}
}

// NewLimited wraps c with at most `max` concurrent calls. If max <= 0 no
// limiting is applied (returns c unchanged).
func NewLimited(c Client, max int) Client {
	if max <= 0 || c == nil {
		return c
	}
	return &Limited{Client: c, sem: make(chan struct{}, max)}
}

func (l *Limited) ChatCompletion(ctx context.Context, req ChatRequest) error {
	select {
	case l.sem <- struct{}{}:
		defer func() { <-l.sem }()
		return l.Client.ChatCompletion(ctx, req)
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (l *Limited) ListModels(ctx context.Context) ([]Model, error) {
	select {
	case l.sem <- struct{}{}:
		defer func() { <-l.sem }()
		return l.Client.ListModels(ctx)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
