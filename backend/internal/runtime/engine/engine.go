package engine

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/aicenter/aicenter/internal/runtime/planner"
	"github.com/aicenter/aicenter/internal/runtime/tools"
	"github.com/aicenter/aicenter/internal/models"
)

// Decision is the approval outcome supplied by an external caller.
type Decision int

const (
	DecisionApprove Decision = iota
	DecisionReject
	DecisionPending
)

// ToolEvent is emitted to the frontend during execution.
type ToolEvent struct {
	Kind    string `json:"kind"` // plan | tool_run | tool_result | approval_required | final | error
	Payload any    `json:"payload,omitempty"`
}

type Plan struct {
	Text      string            `json:"text"`
	ToolCalls []models.ToolCall `json:"tool_calls"`
}

type Result struct {
	Text     string                  `json:"text"`
	ToolRuns []ToolRun               `json:"tool_runs,omitempty"`
	Final    bool                    `json:"final"`
	Approval *models.ApprovalRequest `json:"approval,omitempty"`
}

type ToolRun struct {
	Name   string          `json:"name"`
	Args   json.RawMessage `json:"args,omitempty"`
	Result *tools.Result   `json:"result"`
}

// Executor coordinates the thought-act loop.
type Executor struct {
	reg     *tools.Registry
	mode    string // "allow_all" | "deny_all" | "manual"
	allow   []string
	deny    []string
	maxIter int
}

func New(reg *tools.Registry, agent *models.Agent) *Executor {
	return &Executor{
		reg:     reg,
		mode:    agent.ToolPermissionMode,
		allow:   agent.Tools,
		deny:    agent.RequireApprovalFor,
		maxIter: agent.MaxIterations,
	}
}

// RunPlanner asks the LLM (provided as a function returning a JSON string) to plan.
func (ex *Executor) RunPlanner(ctx context.Context, system string, lastUser string, llm func(prompt string) (string, error)) (*Plan, error) {
	allowed := ex.reg.Resolve(ex.allow)
	if len(allowed) == 0 {
		// If no tools configured, fall back to all read-only ones.
		allowed = ex.readOnlyAll()
	}
	prompt := planner.BuildPrompt(system, lastUser, allowed)
	raw, err := llm(prompt)
	if err != nil {
		return nil, err
	}
	msg, err := planner.Parse(raw)
	if err != nil {
		return nil, err
	}
	return &Plan{Text: msg.Text, ToolCalls: msg.ToolCalls}, nil
}

func (ex *Executor) readOnlyAll() []*tools.Tool {
	out := make([]*tools.Tool, 0)
	for _, t := range ex.reg.All() {
		if t.ReadOnly {
			out = append(out, t)
		}
	}
	return out
}

// Decide returns (shouldExecute bool, requiresApproval bool, decision Decision).
func (ex *Executor) Decide(call models.ToolCall) (bool, bool, Decision) {
	name := call.Name
	switch ex.mode {
	case "allow_all":
		return true, false, DecisionApprove
	case "deny_all":
		return false, false, DecisionReject
	case "manual":
		// Tools in RequireApprovalFor need approval; read-only ones are auto-approved.
		if planner.IsReadOnly(ex.reg, name) {
			return true, false, DecisionApprove
		}
		if contains(ex.deny, name) {
			return false, true, DecisionPending
		}
		return true, false, DecisionApprove
	}
	return false, false, DecisionReject
}

// ExecuteTool runs a tool call (assumes already approved).
func (ex *Executor) ExecuteTool(ctx context.Context, call models.ToolCall, dryRun bool) *tools.Result {
	args := make(map[string]any)
	_ = json.Unmarshal(call.Args, &args)
	return ex.reg.Call(ctx, call.Name, args, dryRun)
}

// BuildApprovalRequest creates the approval row for a tool call that needs review.
func (ex *Executor) BuildApprovalRequest(user, toolName string, args json.RawMessage, dryRun *tools.Result) *models.ApprovalRequest {
	return &models.ApprovalRequest{
		RequestType:  "tool_approval",
		Status:       models.ApprovalPending,
		RequestedBy:  user,
		ToolName:     toolName,
		ToolArgs:     args,
		RiskLevel:    riskFor(ex.reg, toolName),
		DryRunResult: toRaw(dryRun),
	}
}

func riskFor(reg *tools.Registry, name string) string {
	t, ok := reg.Get(name)
	if !ok {
		return "high"
	}
	return t.RiskLevel
}

func toRaw(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

func redact(s string) string {
	parts := strings.Fields(s)
	return strings.Join(parts, " ")
}
