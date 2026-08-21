package ai

import (
	"context"
	"strings"
	"testing"
)

type captureStreamer struct {
	chunks []string
	usage  map[string]int
}

func (s *captureStreamer) PushDelta(chunk string) error {
	s.chunks = append(s.chunks, chunk)
	return nil
}

func (s *captureStreamer) PushUsage(prompt, completion int) error {
	s.usage = map[string]int{"prompt": prompt, "completion": completion}
	return nil
}

func (s *captureStreamer) Flush() error { return nil }

func TestMockClient_ChatCompletion(t *testing.T) {
	c := &MockClient{}
	streamer := &captureStreamer{}
	if err := c.ChatCompletion(context.Background(), ChatRequest{
		Model:    "gpt-4o",
		Messages: []Message{{Role: "user", Content: "hi"}},
		Stream:   true,
		Streamer: streamer,
	}); err != nil {
		t.Fatalf("ChatCompletion: %v", err)
	}
	if len(streamer.chunks) == 0 {
		t.Fatal("expected streamed chunks")
	}
	joined := strings.Join(streamer.chunks, "")
	if !strings.Contains(joined, "Hello from the mock provider") {
		t.Fatalf("unexpected stream content: %q", joined)
	}
	if streamer.usage == nil {
		t.Fatal("expected usage report")
	}
}

func TestMockClient_ListModels(t *testing.T) {
	c := &MockClient{}
	models, err := c.ListModels(context.Background())
	if err != nil {
		t.Fatalf("ListModels: %v", err)
	}
	if len(models) != 2 {
		t.Fatalf("expected 2 mock models, got %d", len(models))
	}
}
