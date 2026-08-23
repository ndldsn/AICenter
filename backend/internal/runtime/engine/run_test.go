package engine

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/aicenter/aicenter/internal/models"
	"github.com/aicenter/aicenter/internal/runtime/tools"
)

// fakeLLM returns a deterministic sequence of planner replies, then a final
// no-tool answer. Each reply can depend on the number of prior calls so we can
// assert the loop advances based on the observations it accumulated.
func newFakeLLM(replies []string) (func(string) (string, error), *int32) {
	var calls int32
	fn := func(prompt string) (string, error) {
		i := atomic.AddInt32(&calls, 1)
		idx := int(i) - 1
		if idx >= len(replies) {
			idx = len(replies) - 1
		}
		return replies[idx], nil
	}
	return fn, &calls
}

func testRegistry() *tools.Registry {
	reg := tools.NewRegistry()
	reg.Register(&tools.Tool{
		Name: "list_servers", Description: "List managed servers", ReadOnly: true, RiskLevel: "low",
		Call: func(ctx context.Context, args map[string]any) *tools.Result {
			return &tools.Result{Ok: true, Status: "ok", Message: "srv-1 (web-01, online)"}
		},
	})
	reg.Register(&tools.Tool{
		Name: "restart_service", Description: "Restart a service", ReadOnly: false, RiskLevel: "high",
		Call: func(ctx context.Context, args map[string]any) *tools.Result {
			return &tools.Result{Ok: true, Status: "ok", Message: "restarted"}
		},
	})
	return reg
}

func TestRun_MultiTurn_UsesObservations(t *testing.T) {
	reg := testRegistry()
	agent := &models.Agent{
		ToolPermissionMode: "allow_all", // auto-execute, no approval needed
		Tools:              []string{"list_servers", "restart_service"},
		MaxIterations:      5,
	}

	// Turn 1: plan calls list_servers. Turn 2: sees the result, calls restart_service.
	// Turn 3: no tool calls -> done.
	llm, calls := newFakeLLM([]string{
		`{"text":"Let me list servers.","tool_calls":[{"name":"list_servers","args":{}}]}`,
		`{"text":"Now restart web-01.","tool_calls":[{"name":"restart_service","args":{"host":"web-01"}}]}`,
		`{"text":"Done. web-01 restarted.","tool_calls":[]}`,
	})

	exec := New(reg, agent)
	var steps int
	final, err := exec.Run(context.Background(), "sys", "fix web-01", llm, RunOptions{
		MaxIterations: agent.MaxIterations,
		OnStep: func(step RunStep) error {
			steps++
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}

	if *calls != 3 {
		t.Fatalf("expected 3 planner calls (plan->observe->plan->observe->done), got %d", *calls)
	}
	if steps != 3 {
		t.Fatalf("expected 3 OnStep callbacks, got %d", steps)
	}
	if final != `"Done. web-01 restarted."` {
		t.Fatalf("unexpected final text: %q", final)
	}
}

func TestRun_StopsAtIterationCap(t *testing.T) {
	reg := testRegistry()
	agent := &models.Agent{ToolPermissionMode: "allow_all", Tools: []string{"list_servers"}, MaxIterations: 3}

	// Always returns a tool call -> must stop after MaxIterations to avoid infinite loop.
	llm, calls := newFakeLLM([]string{
		`{"text":"again","tool_calls":[{"name":"list_servers","args":{}}]}`,
	})
	exec := New(reg, agent)
	_, err := exec.Run(context.Background(), "sys", "loop", llm, RunOptions{MaxIterations: agent.MaxIterations})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if *calls != 3 {
		t.Fatalf("expected loop to stop at cap=3, got %d calls", *calls)
	}
}

func TestRun_ApprovalPausesLoop(t *testing.T) {
	reg := testRegistry()
	// manual mode + restart_service listed in RequireApprovalFor => needs approval.
	agent := &models.Agent{ToolPermissionMode: "manual", Tools: []string{"restart_service"}, RequireApprovalFor: []string{"restart_service"}, MaxIterations: 5}

	llm, _ := newFakeLLM([]string{
		`{"text":"restart it","tool_calls":[{"name":"restart_service","args":{"host":"web-01"}}]}`,
	})
	exec := New(reg, agent)
	var gotApproval bool
	_, err := exec.Run(context.Background(), "sys", "restart web-01", llm, RunOptions{
		MaxIterations: agent.MaxIterations,
		OnStep: func(step RunStep) error {
			if step.Approval != nil {
				gotApproval = true
			}
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if !gotApproval {
		t.Fatal("expected the loop to pause and surface an approval request for a manual-mode write tool")
	}
}

// Ensure observations are actually serialized into the prompt on later turns.
func TestRun_ObservationInjectedIntoPrompt(t *testing.T) {
	reg := testRegistry()
	agent := &models.Agent{ToolPermissionMode: "allow_all", Tools: []string{"list_servers", "restart_service"}, MaxIterations: 5}

	var seenPrompts []string
	llm := func(prompt string) (string, error) {
		seenPrompts = append(seenPrompts, prompt)
		if len(seenPrompts) == 1 {
			return `{"text":"list","tool_calls":[{"name":"list_servers","args":{}}]}`, nil
		}
		return `{"text":"done","tool_calls":[]}`, nil
	}
	exec := New(reg, agent)
	_, err := exec.Run(context.Background(), "sys", "go", llm, RunOptions{MaxIterations: agent.MaxIterations})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if len(seenPrompts) < 2 {
		t.Fatalf("expected >=2 prompts, got %d", len(seenPrompts))
	}
	// The second prompt must contain the prior tool result text.
	if !containsStr(seenPrompts[1], "srv-1") {
		t.Fatalf("second-turn prompt missing prior observation; got:\n%s", seenPrompts[1])
	}
}

func containsStr(haystack, needle string) bool {
	return len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0
}

func indexOf(haystack, needle string) int {
	// tiny helper to avoid importing strings just for Contains in test
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}

