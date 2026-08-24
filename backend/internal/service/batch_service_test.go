package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/aicenter/aicenter/internal/models"
)

// fakeLister implements ServerLister for unit tests (no DB needed).
type fakeLister struct {
	servers []*models.Server
	err     error
}

func (f *fakeLister) List(offset, limit int) ([]*models.Server, int64, error) {
	if f.err != nil {
		return nil, 0, f.err
	}
	return f.servers, int64(len(f.servers)), nil
}

func TestBatchCommand_EmptyCommand(t *testing.T) {
	s := NewBatchServiceWithStore(&fakeLister{})
	res := s.BatchCommand(context.Background(), &BatchRequest{Command: ""})
	if res != nil {
		t.Fatalf("expected nil for empty command, got %v", res)
	}
}

func TestBatchCommand_LocalExecEcho(t *testing.T) {
	s := NewBatchServiceWithStore(&fakeLister{
		servers: []*models.Server{{
			ID: "s1", Name: "local", Host: "localhost", Port: 22,
			Username: "u", AuthType: "password",
		}},
	})
	res := s.BatchCommand(context.Background(), &BatchRequest{
		Command: "echo BATCH_UNIT_OK", ServerIDs: []string{"s1"}, Timeout: 10,
	})
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	r := res[0]
	if r.Status != "ok" {
		t.Fatalf("status=%q stdout=%q", r.Status, r.Stdout)
	}
	if !strings.Contains(r.Stdout, "BATCH_UNIT_OK") {
		t.Fatalf("expected output echoed, got %q", r.Stdout)
	}
}

func TestBatchCommand_NonZeroExitCode(t *testing.T) {
	s := NewBatchServiceWithStore(&fakeLister{
		servers: []*models.Server{{ID: "s1", Name: "local", Host: "localhost", Port: 22}},
	})
	res := s.BatchCommand(context.Background(), &BatchRequest{
		Command: "exit 7", ServerIDs: []string{"s1"}, Timeout: 10,
	})
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	if res[0].Status != "ok" {
		t.Fatalf("status=%q (non-zero exit is still a clean exit)", res[0].Status)
	}
	if res[0].ExitCode != 7 {
		t.Fatalf("expected exit code 7, got %d", res[0].ExitCode)
	}
}

func TestBatchCommand_TimeoutSurfacesAsFailed(t *testing.T) {
    s := NewBatchServiceWithStore(&fakeLister{
        servers: []*models.Server{{ID: "s1", Name: "local", Host: "localhost", Port: 22}},
    })
    res := s.BatchCommand(context.Background(), &BatchRequest{
        Command: blockCommand(10), ServerIDs: []string{"s1"}, Timeout: 1,
    })
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	r := res[0]
	if r.Status != "failed" {
		t.Fatalf("expected failed on timeout, got %q (dur=%s)", r.Status, r.Duration)
	}
	if !strings.Contains(r.Error, "timeout") {
		t.Fatalf("expected timeout error, got %q", r.Error)
	}
	if r.Duration == "" {
		t.Fatal("expected a duration")
	}
}

func TestBatchCommand_ListErrorReturnsFailure(t *testing.T) {
	s := NewBatchServiceWithStore(&fakeLister{err: errSentinel})
	res := s.BatchCommand(context.Background(), &BatchRequest{Command: "echo hi"})
	if len(res) != 1 || res[0].Status != "failed" {
		t.Fatalf("expected a single failure result, got %v", res)
	}
}

func TestBatchCommand_RunsMultipleServersConcurrently(t *testing.T) {
	s := NewBatchServiceWithStore(&fakeLister{
		servers: []*models.Server{
			{ID: "s1", Name: "a", Host: "localhost", Port: 22},
			{ID: "s2", Name: "b", Host: "localhost", Port: 22},
		},
	})
	start := time.Now()
	res := s.BatchCommand(context.Background(), &BatchRequest{
		Command: "sleep 1", ServerIDs: []string{"s1", "s2"}, Timeout: 10,
	})
	elapsed := time.Since(start)
	if len(res) != 2 {
		t.Fatalf("expected 2 results, got %d", len(res))
	}
	// Both must run concurrently: ~1s not ~2s.
	if elapsed > 1800*time.Millisecond {
		t.Fatalf("expected concurrent execution (~1s), took %v", elapsed)
	}
}

var errSentinel = errString("simulated list error")

type errString string

func (e errString) Error() string { return string(e) }
