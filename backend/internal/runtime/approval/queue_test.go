package approval

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/aicenter/aicenter/internal/models"
)

// --- NewQueue ---

func TestNewQueue_Empty(t *testing.T) {
	q := NewQueue()
	if q == nil {
		t.Fatal("NewQueue should return non-nil")
	}
}

// --- WaitForApproval + Resolve happy path ---

func TestWaitForApproval_Resolved(t *testing.T) {
	q := NewQueue()
	var status models.ApprovalStatus
	var err error

	done := make(chan struct{})
	go func() {
		status, err = q.WaitForApproval(context.Background(), "req-1")
		close(done)
	}()

	// Give the goroutine time to register
	time.Sleep(50 * time.Millisecond)

	if !q.Resolve("req-1", models.ApprovalApproved) {
		t.Fatal("Resolve should return true for pending request")
	}

	<-done
	if err != nil {
		t.Fatalf("WaitForApproval: %v", err)
	}
	if status != models.ApprovalApproved {
		t.Fatalf("expected approved, got %s", status)
	}
}

// --- Resolve for non-existent request ---

func TestResolve_NotFound(t *testing.T) {
	q := NewQueue()
	if q.Resolve("nonexistent", models.ApprovalApproved) {
		t.Fatal("Resolve should return false for unknown request")
	}
}

// --- WaitForApproval with context cancellation ---

func TestWaitForApproval_ContextCancel(t *testing.T) {
	q := NewQueue()
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		_, _ = q.WaitForApproval(ctx, "req-cancel")
		close(done)
	}()

	// Cancel context
	cancel()

	select {
	case <-done:
		// success: WaitForApproval returned
	case <-time.After(2 * time.Second):
		t.Fatal("WaitForApproval should return after context cancel")
	}
}

// --- Concurrent Resolve ---

func TestResolve_Concurrent(t *testing.T) {
	q := NewQueue()

	// Start multiple waiters
	const n = 5
	results := make(chan models.ApprovalStatus, n)
	for i := 0; i < n; i++ {
		go func(id string) {
			s, _ := q.WaitForApproval(context.Background(), id)
			results <- s
		}(string(rune('A' + i)))
	}

	// Resolve each one
	time.Sleep(50 * time.Millisecond)
	for i := 0; i < n; i++ {
		id := string(rune('A' + i))
		if !q.Resolve(id, models.ApprovalApproved) {
			t.Fatalf("failed to resolve %s", id)
		}
	}

	// Collect results
	for i := 0; i < n; i++ {
		select {
		case s := <-results:
			if s != models.ApprovalApproved {
				t.Fatalf("expected approved, got %s", s)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for approval result")
		}
	}
}

// --- Resolve removes entry (double resolve returns false) ---

func TestResolve_DoubleResolveFails(t *testing.T) {
	q := NewQueue()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		q.WaitForApproval(context.Background(), "req-dup")
	}()

	time.Sleep(50 * time.Millisecond)
	if !q.Resolve("req-dup", models.ApprovalApproved) {
		t.Fatal("first resolve should succeed")
	}
	wg.Wait()
	if q.Resolve("req-dup", models.ApprovalRejected) {
		t.Fatal("second resolve should fail (already removed)")
	}
}

// --- WaitForApproval rejected ---

func TestWaitForApproval_Rejected(t *testing.T) {
	q := NewQueue()
	var status models.ApprovalStatus

	done := make(chan struct{})
	go func() {
		status, _ = q.WaitForApproval(context.Background(), "req-rej")
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	q.Resolve("req-rej", models.ApprovalRejected)

	<-done
	if status != models.ApprovalRejected {
		t.Fatalf("expected rejected, got %s", status)
	}
}