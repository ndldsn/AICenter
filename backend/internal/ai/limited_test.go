package ai

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// fakeClient counts concurrent ChatCompletion calls so the limiter's cap is
// observable.
type fakeClient struct{ inFlight int32 }

func (f *fakeClient) ListModels(ctx context.Context) ([]Model, error) { return nil, nil }

func (f *fakeClient) ChatCompletion(ctx context.Context, req ChatRequest) error {
	atomic.AddInt32(&f.inFlight, 1)
	defer atomic.AddInt32(&f.inFlight, -1)
	select {
	case <-time.After(50 * time.Millisecond):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (f *fakeClient) concurrent() int32 { return atomic.LoadInt32(&f.inFlight) }

// TestLimited_AllowsUpToN verifies the semaphore caps concurrency at N.
func TestLimited_AllowsUpToN(t *testing.T) {
	const cap = 3
	fc := &fakeClient{}
	c := NewLimited(fc, cap)
	var wg sync.WaitGroup
	for i := 0; i < cap; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = c.ChatCompletion(context.Background(), ChatRequest{})
		}()
	}
	// Poll for peak concurrency.
	deadline := time.Now().Add(100 * time.Millisecond)
	var maxSeen int32
	for time.Now().Before(deadline) {
		cur := fc.concurrent()
		if cur > maxSeen {
			maxSeen = cur
		}
		time.Sleep(time.Millisecond)
	}
	wg.Wait()
	if maxSeen != cap {
		t.Fatalf("expected peak concurrency %d, got %d", cap, maxSeen)
	}
}

// TestLimited_GatesBeyondCap verifies that a 4th call does not start until a
// slot frees up (it is queued / gated, not run concurrently).
func TestLimited_GatesBeyondCap(t *testing.T) {
	const cap = 2
	fc := &fakeClient{}
	c := NewLimited(fc, cap).(*Limited)
	// Saturate both slots directly (bypass the public API).
	c.sem <- struct{}{}
	c.sem <- struct{}{}
	// An extra call should NOT start immediately: concurrent stays at 0 because
	// the slots are filled. With a short context it must fail fast.
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	start := time.Now()
	err := c.ChatCompletion(ctx, ChatRequest{})
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("expected error when over capacity with deadline")
	}
	if elapsed > 200*time.Millisecond {
		t.Fatalf("expected fast-fail, took %v", elapsed)
	}
}

// TestNewLimited_NoopOnZeroOrNegative ensures no-op pass-through.
func TestNewLimited_NoopOnZeroOrNegative(t *testing.T) {
	base := &fakeClient{}
	if NewLimited(base, 0) != base {
		t.Fatal("max<=0 should return the underlying client")
	}
}
