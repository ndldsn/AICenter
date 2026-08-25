package planner

import (
	"testing"

	"github.com/aicenter/aicenter/internal/runtime/tools"
)

// --- Parse ---

func TestParse_PlainText(t *testing.T) {
	msg, err := Parse("Hello, I am an assistant.")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if msg.Text != "Hello, I am an assistant." {
		t.Fatalf("expected text, got %q", msg.Text)
	}
	if len(msg.ToolCalls) != 0 {
		t.Fatalf("expected no tool calls, got %d", len(msg.ToolCalls))
	}
}

func TestParse_JsonWithTextOnly(t *testing.T) {
	msg, err := Parse(`{"text":"Just thinking..."}`)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if msg.Text != `"Just thinking..."` {
		t.Fatalf("unexpected text: %q", msg.Text)
	}
	if len(msg.ToolCalls) != 0 {
		t.Fatal("expected no tool calls")
	}
}

func TestParse_JsonWithContentKey(t *testing.T) {
	msg, err := Parse(`{"content":"Using content key"}`)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if msg.Text != `"Using content key"` {
		t.Fatalf("unexpected text: %q", msg.Text)
	}
}

func TestParse_JsonWithToolCalls(t *testing.T) {
	raw := `{"text":"Let me check.","tool_calls":[{"name":"list_servers","args":{}}]}`
	msg, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(msg.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(msg.ToolCalls))
	}
	if msg.ToolCalls[0].Name != "list_servers" {
		t.Fatalf("expected tool name list_servers, got %s", msg.ToolCalls[0].Name)
	}
}

func TestParse_JsonWithMultipleToolCalls(t *testing.T) {
	raw := `{"text":"Doing two things.","tool_calls":[{"name":"list_servers","args":{}},{"name":"read_file","args":{"path":"/etc/hosts"}}]}`
	msg, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(msg.ToolCalls) != 2 {
		t.Fatalf("expected 2 tool calls, got %d", len(msg.ToolCalls))
	}
	if msg.ToolCalls[1].Name != "read_file" {
		t.Fatalf("expected second tool read_file, got %s", msg.ToolCalls[1].Name)
	}
}

func TestParse_EmptyToolCalls(t *testing.T) {
	raw := `{"text":"Done.","tool_calls":[]}`
	msg, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(msg.ToolCalls) != 0 {
		t.Fatal("expected empty tool calls")
	}
}

func TestParse_InvalidToolCalls(t *testing.T) {
	raw := `{"text":"Oops","tool_calls":"not-an-array"}`
	_, err := Parse(raw)
	if err == nil {
		t.Fatal("expected error for invalid tool_calls format")
	}
}

func TestParse_EmptyString(t *testing.T) {
	msg, err := Parse("")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if msg.Text != "" {
		t.Fatalf("expected empty text, got %q", msg.Text)
	}
}

// --- BuildPrompt ---

func TestBuildPrompt_NoObservations(t *testing.T) {
	reg := tools.NewRegistry()
	reg.Register(&tools.Tool{Name: "list_servers", Description: "List all servers", ReadOnly: true})
	reg.Register(&tools.Tool{Name: "restart", Description: "Restart a service", ReadOnly: false})

	prompt := BuildPrompt("You are an agent.", "Check servers", reg.All(), nil)
	if prompt == "" {
		t.Fatal("prompt must not be empty")
	}
	// Must contain system message
	if !containsSubstr(prompt, "You are an agent.") {
		t.Fatal("prompt must contain system message")
	}
	// Must contain user query
	if !containsSubstr(prompt, "Check servers") {
		t.Fatal("prompt must contain user query")
	}
	// Must list available tools
	if !containsSubstr(prompt, "list_servers") || !containsSubstr(prompt, "restart") {
		t.Fatal("prompt must list available tools")
	}
	// Must contain JSON instruction
	if !containsSubstr(prompt, `"text"`) {
		t.Fatal("prompt must instruct JSON response format")
	}
}

func TestBuildPrompt_WithObservations(t *testing.T) {
	reg := tools.NewRegistry()
	reg.Register(&tools.Tool{Name: "list_servers", Description: "List all servers", ReadOnly: true})

	obs := []Observation{
		{ToolName: "list_servers", Args: map[string]any{}, Result: "srv-1 (web, online)"},
	}
	prompt := BuildPrompt("System", "go", reg.All(), obs)
	if !containsSubstr(prompt, "TOOL RESULTS SO FAR") {
		t.Fatal("prompt must contain tool results section when observations exist")
	}
	if !containsSubstr(prompt, "srv-1") {
		t.Fatal("prompt must contain observation result")
	}
	if !containsSubstr(prompt, "Use these results") {
		t.Fatal("prompt must instruct to use prior results")
	}
}

func TestBuildPrompt_MultipleObservations(t *testing.T) {
	reg := tools.NewRegistry()
	reg.Register(&tools.Tool{Name: "read_file", Description: "Read a file", ReadOnly: true})

	obs := []Observation{
		{ToolName: "read_file", Args: map[string]any{"path": "/a"}, Result: "content-a"},
		{ToolName: "read_file", Args: map[string]any{"path": "/b"}, Result: "content-b"},
	}
	prompt := BuildPrompt("Sys", "check", reg.All(), obs)
	if !containsSubstr(prompt, "1. read_file") {
		t.Fatal("prompt must number observations")
	}
	if !containsSubstr(prompt, "2. read_file") {
		t.Fatal("prompt must number second observation")
	}
}

// --- IsReadOnly ---

func TestIsReadOnly_ExistingReadOnlyTool(t *testing.T) {
	reg := tools.NewRegistry()
	reg.Register(&tools.Tool{Name: "list_servers", Description: "List servers", ReadOnly: true})
	if !IsReadOnly(reg, "list_servers") {
		t.Fatal("list_servers should be read-only")
	}
}

func TestIsReadOnly_WriteTool(t *testing.T) {
	reg := tools.NewRegistry()
	reg.Register(&tools.Tool{Name: "restart", Description: "Restart", ReadOnly: false})
	if IsReadOnly(reg, "restart") {
		t.Fatal("restart should not be read-only")
	}
}

func TestIsReadOnly_UnknownTool(t *testing.T) {
	reg := tools.NewRegistry()
	if IsReadOnly(reg, "nonexistent") {
		t.Fatal("unknown tool should not be considered read-only")
	}
}

// helper
func containsSubstr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}