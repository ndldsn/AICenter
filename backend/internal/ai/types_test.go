package ai

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestStreamEvents_PushDeltaEmitsJSON verifies the SSE wire format the frontend
// consumes: each delta is a `data: {"content":"..."}` line terminated by \n\n.
func TestStreamEvents_PushDeltaEmitsJSON(t *testing.T) {
	var sb strings.Builder
	s := NewStreamEvents(&sb)

	if err := s.PushDelta("Hello"); err != nil {
		t.Fatalf("PushDelta: %v", err)
	}
	if err := s.PushDelta(" world"); err != nil {
		t.Fatalf("PushDelta: %v", err)
	}

	out := sb.String()
	lines := strings.Split(strings.TrimSpace(out), "\n\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 SSE frames, got %d: %q", len(lines), out)
	}
	for _, frame := range lines {
		if !strings.HasPrefix(frame, "data: ") {
			t.Fatalf("frame must start with 'data: ': %q", frame)
		}
		payload := strings.TrimPrefix(frame, "data: ")
		var obj struct {
			Content string `json:"content"`
		}
		if err := json.Unmarshal([]byte(payload), &obj); err != nil {
			t.Fatalf("frame payload is not valid JSON: %v (%q)", err, payload)
		}
	}
	joined := strings.ReplaceAll(out, "\n\n", "")
	joined = strings.TrimPrefix(joined, "data: ")
	joined = strings.TrimPrefix(joined, `{"content":"`)
	// sanity: both deltas present
	if !strings.Contains(out, `"content":"Hello"`) || !strings.Contains(out, `"content":" world"`) {
		t.Fatalf("expected both content chunks in output: %q", out)
	}
}

// TestStreamEvents_PushUsage verifies the trailing usage frame shape.
func TestStreamEvents_PushUsage(t *testing.T) {
	var sb strings.Builder
	s := NewStreamEvents(&sb)
	if err := s.PushUsage(42, 22); err != nil {
		t.Fatalf("PushUsage: %v", err)
	}
	out := sb.String()
	payload := strings.TrimPrefix(strings.TrimSpace(out), "data: ")
	var obj struct {
		Usage struct {
			Prompt     int `json:"prompt_tokens"`
			Completion int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal([]byte(payload), &obj); err != nil {
		t.Fatalf("usage payload not valid JSON: %v (%q)", err, payload)
	}
	if obj.Usage.Prompt != 42 || obj.Usage.Completion != 22 {
		t.Fatalf("usage values wrong: %+v", obj.Usage)
	}
}
