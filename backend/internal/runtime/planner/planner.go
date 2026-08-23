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

// Observation is a single tool execution result fed back to the LLM so it can
// reason about what happened and decide the next step (the "observe" half of
// the thought-act loop).
type Observation struct {
	ToolName string         `json:"tool_name"`
	Args     map[string]any `json:"args,omitempty"`
	Result   string         `json:"result"`
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
// Prior tool observations (if any) are appended so the model can reason about
// what already happened and decide the next step instead of repeating itself.
func BuildPrompt(system string, lastUser string, tools_ []*tools.Tool, observations []Observation) string {
	var buf strings.Builder
	buf.WriteString(system)
	buf.WriteString("\n\n## AVAILABLE TOOLS\n\n")
	for _, t := range tools_ {
		buf.WriteString(fmt.Sprintf("- %s: %s\n", t.Name, t.Description))
	}

	if len(observations) > 0 {
		buf.WriteString("\n## TOOL RESULTS SO FAR\n\n")
		for i, o := range observations {
			argsJSON, _ := json.Marshal(o.Args)
			buf.WriteString(fmt.Sprintf("%d. %s(args=%s) -> %s\n", i+1, o.ToolName, string(argsJSON), o.Result))
		}
		buf.WriteString("\nUse these results to continue. If the task is complete, respond with an empty tool_calls list and a final text answer.\n")
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
