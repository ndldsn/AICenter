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

// RunStep is emitted for every turn so callers (REST/SSE) can persist, notify,
// and stream progress without the engine knowing about storage or HTTP.
type RunStep struct {
	Plan      *Plan
	ToolRuns  []ToolRun
	Approval  *models.ApprovalRequest
	FinalText string
	Done      bool
}

// RunOptions configures the multi-turn loop.
type RunOptions struct {
	// MaxIterations caps planner calls; 0 falls back to the agent's configured cap.
	MaxIterations int
	// OnStep is called after each planner turn (with its tool execution results
	// and any approval request). Returning a non-nil error aborts the loop.
	OnStep func(step RunStep) error
}

// Run executes the full thought-act loop: plan → execute → observe → replan,
// repeating until the LLM emits no tool calls, an approval is required, or the
// iteration cap is reached. It is shared by the REST and SSE handlers so both
// get true multi-turn reasoning rather than a single plan+one tool call.
func (ex *Executor) Run(ctx context.Context, system, userQuery string, llm func(prompt string) (string, error), opts RunOptions) (string, error) {
	maxIter := opts.MaxIterations
	if maxIter <= 0 {
		maxIter = ex.maxIter
	}
	if maxIter <= 0 {
		maxIter = 8
	}

	observations := make([]planner.Observation, 0)
	finalText := ""

	for iter := 0; iter < maxIter; iter++ {
		if ctx.Err() != nil {
			return finalText, ctx.Err()
		}

		allowed := ex.reg.Resolve(ex.allow)
		if len(allowed) == 0 {
			allowed = ex.readOnlyAll()
		}
		prompt := planner.BuildPrompt(system, userQuery, allowed, observations)
		raw, err := llm(prompt)
		if err != nil {
			return finalText, err
		}
		msg, err := planner.Parse(raw)
		if err != nil {
			return finalText, err
		}

		plan := &Plan{Text: msg.Text, ToolCalls: msg.ToolCalls}
		if plan.Text != "" {
			finalText = plan.Text
		}

		// No tool calls -> the agent considers the task done.
		if len(plan.ToolCalls) == 0 {
			step := RunStep{Plan: plan, FinalText: finalText, Done: true}
			if opts.OnStep != nil {
				if err := opts.OnStep(step); err != nil {
					return finalText, err
				}
			}
			return finalText, nil
		}

		// Execute each tool call in the plan, collecting observations.
		turnRuns := make([]ToolRun, 0, len(plan.ToolCalls))
		stepApproval := (*models.ApprovalRequest)(nil)
		pausedForApproval := false

		for i, call := range plan.ToolCalls {
			execute, needsApproval, _ := ex.Decide(call)
			if !execute {
				if needsApproval {
					dry := ex.ExecuteTool(ctx, call, true)
					approv := ex.BuildApprovalRequest("", call.Name, call.Args, dry)
					run := ToolRun{Name: call.Name, Args: call.Args, Result: &tools.Result{Ok: false, Status: "pending_approval"}}
					turnRuns = append(turnRuns, run)
					observations = append(observations, planner.Observation{
						ToolName: call.Name,
						Args:     map[string]any{"_denied": "pending_approval"},
						Result:   "pending human approval",
					})
					stepApproval = approv
					pausedForApproval = true
					_ = i
					continue
				}
				run := ToolRun{Name: call.Name, Args: call.Args, Result: &tools.Result{Ok: false, Status: "denied"}}
				turnRuns = append(turnRuns, run)
				observations = append(observations, planner.Observation{
					ToolName: call.Name,
					Args:     map[string]any{"_denied": true},
					Result:   "denied by policy",
				})
				_ = i
				continue
			}

			res := ex.ExecuteTool(ctx, call, false)
			turnRuns = append(turnRuns, ToolRun{Name: call.Name, Args: call.Args, Result: res})
			observations = append(observations, planner.Observation{
				ToolName: call.Name,
				Args:     argsOf(call.Args),
				Result:   resultSummary(res),
			})
		}

		// If any tool call needs approval, pause the loop and surface it.
		if pausedForApproval {
			step := RunStep{Plan: plan, ToolRuns: turnRuns, Approval: stepApproval, Done: false}
			if opts.OnStep != nil {
				if err := opts.OnStep(step); err != nil {
					return finalText, err
				}
			}
			return finalText, nil
		}

		step := RunStep{Plan: plan, ToolRuns: turnRuns, Done: false}
		if opts.OnStep != nil {
			if err := opts.OnStep(step); err != nil {
				return finalText, err
			}
		}
	}

	// Hit the iteration cap: return whatever we have as the final answer.
	return finalText, nil
}

func argsOf(raw json.RawMessage) map[string]any {
	m := make(map[string]any)
	_ = json.Unmarshal(raw, &m)
	return m
}

func resultSummary(r *tools.Result) string {
	if r == nil {
		return "no result"
	}
	if r.Message != "" {
		return r.Message
	}
	if r.Status != "" {
		return r.Status
	}
	return "ok"
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
