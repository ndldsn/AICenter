package tools

import (
	"context"
	"testing"
)

// helper: a simple no-op tool for testing
func dummyTool(name string, readOnly bool) *Tool {
	return &Tool{
		Name:        name,
		Description: name + " tool",
		ReadOnly:    readOnly,
		Call: func(ctx context.Context, args map[string]any) *Result {
			return &Result{Ok: true, Status: "ok", Payload: args}
		},
	}
}

// --- NewRegistry ---

func TestNewRegistry_Empty(t *testing.T) {
	r := NewRegistry()
	if len(r.All()) != 0 {
		t.Fatal("new registry should be empty")
	}
	if len(r.Names()) != 0 {
		t.Fatal("new registry should have no names")
	}
}

// --- Register + Get ---

func TestRegister_AndGet(t *testing.T) {
	r := NewRegistry()
	tool := dummyTool("list_servers", true)
	r.Register(tool)

	got, ok := r.Get("list_servers")
	if !ok {
		t.Fatal("expected to find registered tool")
	}
	if got.Name != "list_servers" {
		t.Fatalf("expected list_servers, got %s", got.Name)
	}
}

func TestGet_NotFound(t *testing.T) {
	r := NewRegistry()
	_, ok := r.Get("nonexistent")
	if ok {
		t.Fatal("should not find unregistered tool")
	}
}

func TestRegister_Overwrite(t *testing.T) {
	r := NewRegistry()
	r.Register(dummyTool("x", true))
	r.Register(dummyTool("x", false)) // overwrite
	got, ok := r.Get("x")
	if !ok {
		t.Fatal("expected to find tool")
	}
	if got.ReadOnly {
		t.Fatal("expected ReadOnly=false after overwrite")
	}
}

// --- All ---

func TestAll_ReturnsAllTools(t *testing.T) {
	r := NewRegistry()
	r.Register(dummyTool("a", true))
	r.Register(dummyTool("b", false))
	all := r.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(all))
	}
}

// --- Names ---

func TestNames_ReturnsAllNames(t *testing.T) {
	r := NewRegistry()
	r.Register(dummyTool("alpha", true))
	r.Register(dummyTool("beta", false))
	names := r.Names()
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}
	// Check both names exist (order not guaranteed)
	found := map[string]bool{}
	for _, n := range names {
		found[n] = true
	}
	if !found["alpha"] || !found["beta"] {
		t.Fatalf("expected alpha and beta, got %v", names)
	}
}

// --- Resolve ---

func TestResolve_FiltersByAllowList(t *testing.T) {
	r := NewRegistry()
	r.Register(dummyTool("a", true))
	r.Register(dummyTool("b", false))
	r.Register(dummyTool("c", true))

	resolved := r.Resolve([]string{"a", "c"})
	if len(resolved) != 2 {
		t.Fatalf("expected 2 resolved, got %d", len(resolved))
	}
	for _, tool := range resolved {
		if tool.Name == "b" {
			t.Fatal("b should not be resolved")
		}
	}
}

func TestResolve_EmptyAllowList(t *testing.T) {
	r := NewRegistry()
	r.Register(dummyTool("a", true))
	resolved := r.Resolve([]string{})
	if len(resolved) != 0 {
		t.Fatalf("expected 0 resolved, got %d", len(resolved))
	}
}

func TestResolve_UnknownNamesIgnored(t *testing.T) {
	r := NewRegistry()
	r.Register(dummyTool("a", true))
	resolved := r.Resolve([]string{"a", "nonexistent"})
	if len(resolved) != 1 {
		t.Fatalf("expected 1 resolved, got %d", len(resolved))
	}
}

// --- Call ---

func TestCall_ExistingTool(t *testing.T) {
	r := NewRegistry()
	r.Register(dummyTool("echo", true))
	res := r.Call(context.Background(), "echo", map[string]any{"key": "val"}, false)
	if !res.Ok {
		t.Fatal("expected Ok=true")
	}
	if res.Status != "ok" {
		t.Fatalf("expected status ok, got %s", res.Status)
	}
	if res.Payload["key"] != "val" {
		t.Fatalf("expected payload key=val, got %v", res.Payload)
	}
}

func TestCall_ToolNotFound(t *testing.T) {
	r := NewRegistry()
	res := r.Call(context.Background(), "missing", nil, false)
	if res.Ok {
		t.Fatal("expected Ok=false for missing tool")
	}
	if res.Status != "error" {
		t.Fatalf("expected status error, got %s", res.Status)
	}
}

func TestCall_DryRun(t *testing.T) {
	r := NewRegistry()
	r.Register(dummyTool("write_file", false))
	res := r.Call(context.Background(), "write_file", map[string]any{"path": "/tmp/x"}, true)
	if res.Status != "dry_run" {
		t.Fatalf("expected status dry_run, got %s", res.Status)
	}
	if res.DryRun == nil {
		t.Fatal("expected DryRun to be populated")
	}
	if res.Payload != nil {
		t.Fatal("expected Payload to be nil in dry-run mode")
	}
}

// --- Result.MarshalArgs ---

func TestResult_MarshalArgs(t *testing.T) {
	r := &Result{Payload: map[string]any{"foo": "bar"}}
	raw := r.MarshalArgs()
	if len(raw) == 0 {
		t.Fatal("expected non-empty raw message")
	}
}