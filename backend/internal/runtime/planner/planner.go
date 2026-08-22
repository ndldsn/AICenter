package planner

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aicenter/aicenter/internal/runtime/tools"
	"github.com/aicenter/aicenter/internal/models"
)

// ToolMessage is a single assistant response chunk that may contain text and/or tool calls.
type ToolMessage struct {
	Text      string
	ToolCalls []models.ToolCall
}

// Parse parses the assistant's raw JSON response into a ToolMessage.
// Accepted shapes (best-effort):
//   - {"text":"...","tool_calls":[{"name":"x","args":{...}}]}
//   - {"content":"...","tool_calls":[...]}
//   - {"text":"..."}  (no tools)
func Parse(raw string) (*ToolMessage, error) {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &obj); err != nil {
		// Not JSON: treat as a plain text answer with no tool calls.
		return &ToolMessage{Text: strings.TrimSpace(raw)}, nil
	}

	msg := &ToolMessage{}
	if t, ok := obj["text"]; ok {
		msg.Text = string(t)
	} else if c, ok := obj["content"]; ok {
		msg.Text = string(c)
	}

	rawCalls := obj["tool_calls"]
	if rawCalls == nil {
		return msg, nil
	}

	var calls []struct {
		Name string          `json:"name"`
		Args json.RawMessage `json:"args"`
	}
	if err := json.Unmarshal(rawCalls, &calls); err != nil {
		return nil, fmt.Errorf("invalid tool_calls: %w", err)
	}
	msg.ToolCalls = make([]models.ToolCall, len(calls))
	for i, c := range calls {
		msg.ToolCalls[i] = models.ToolCall{Name: c.Name, Args: c.Args}
	}
	return msg, nil
}

// BuildPrompt returns the system + last user message as a formatted prompt
// suitable for a text-only LLM (the LLM then emits JSON with tool_calls).
func BuildPrompt(system string, lastUser string, tools_ []*tools.Tool) string {
	var buf strings.Builder
	buf.WriteString(system)
	buf.WriteString("\n\n## AVAILABLE TOOLS\n\n")
	for _, t := range tools_ {
		buf.WriteString(fmt.Sprintf("- %s: %s\n", t.Name, t.Description))
	}
	buf.WriteString(fmt.Sprintf("\n## USER QUERY\n%s\n", lastUser))
	buf.WriteString("\nRespond as JSON with {\"text\": \"...\", \"tool_calls\": [{\"name\": ..., \"args\": {...}}]}. If no tool is needed, set tool_calls to [].")
	return buf.String()
}

// IsReadOnly reports whether the named tool is declared read-only.
func IsReadOnly(reg *tools.Registry, name string) bool {
	t, ok := reg.Get(name)
	if !ok {
		return false
	}
	return t.ReadOnly
}
