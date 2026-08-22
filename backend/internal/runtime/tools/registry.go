package tools

import (
	"context"
	"encoding/json"

	"github.com/aicenter/aicenter/internal/repository"
)

// Result carries a tool execution outcome.
type Result struct {
	Ok      bool           `json:"ok"`
	Status  string         `json:"status,omitempty"` // ok | error | dry_run
	Message string         `json:"message,omitempty"`
	Payload map[string]any `json:"payload,omitempty"`
	DryRun  map[string]any `json:"dry_run,omitempty"`
}

func (r *Result) MarshalArgs() json.RawMessage {
	b, _ := json.Marshal(r.Payload)
	return b
}

// Tool is a named, callable capability exposed to an agent.
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ArgsSchema  json.RawMessage
	ReadOnly    bool   `json:"read_only"`
	Destructive bool   `json:"destructive"`
	RiskLevel   string `json:"risk_level"` // low | medium | high
	Call        func(ctx context.Context, args map[string]any) *Result
}

// Registry is the set of tools an agent is allowed to call.
type Registry struct {
	tools map[string]*Tool
}

func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]*Tool)}
}

func (r *Registry) Register(t *Tool) {
	r.tools[t.Name] = t
}

func (r *Registry) Get(name string) (*Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

func (r *Registry) All() []*Tool {
	out := make([]*Tool, 0, len(r.tools))
	for _, t := range r.tools {
		out = append(out, t)
	}
	return out
}

func (r *Registry) Names() []string {
	out := make([]string, 0, len(r.tools))
	for n := range r.tools {
		out = append(out, n)
	}
	return out
}

// Resolve returns the Tools whose names are present in the allow list.
func (r *Registry) Resolve(allowList []string) []*Tool {
	out := make([]*Tool, 0, len(allowList))
	for _, n := range allowList {
		if t, ok := r.tools[n]; ok {
			out = append(out, t)
		}
	}
	return out
}

// Call dispatches a tool call, optionally in dry-run mode.
func (r *Registry) Call(ctx context.Context, name string, args map[string]any, dryRun bool) *Result {
	t, ok := r.Get(name)
	if !ok {
		return &Result{Ok: false, Status: "error", Message: "tool not found: " + name}
	}
	res := t.Call(ctx, args)
	if dryRun {
		res.DryRun = res.Payload
		res.Payload = nil
		res.Status = "dry_run"
	}
	return res
}

// AuditEntry from repo (used by runtime).
func RecordAudit(repo *repository.AuditRepository, entry *repository.AuditEntry) error {
	return repo.Record(entry)
}
