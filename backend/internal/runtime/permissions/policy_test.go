package permissions

import (
	"context"
	"testing"

	"github.com/aicenter/aicenter/internal/runtime/tools"
)

// testRegistry builds a small tool registry for policy tests.
func testPolicyRegistry() *tools.Registry {
	reg := tools.NewRegistry()
	reg.Register(&tools.Tool{
		Name: "list_servers", Description: "List servers", ReadOnly: true, RiskLevel: "low",
		Call: func(ctx context.Context, args map[string]any) *tools.Result {
			return &tools.Result{Ok: true, Status: "ok"}
		},
	})
	reg.Register(&tools.Tool{
		Name: "read_file", Description: "Read a file", ReadOnly: true, RiskLevel: "low",
		Call: func(ctx context.Context, args map[string]any) *tools.Result {
			return &tools.Result{Ok: true, Status: "ok"}
		},
	})
	reg.Register(&tools.Tool{
		Name: "restart_service", Description: "Restart a service", ReadOnly: false, RiskLevel: "high",
		Call: func(ctx context.Context, args map[string]any) *tools.Result {
			return &tools.Result{Ok: true, Status: "ok"}
		},
	})
	reg.Register(&tools.Tool{
		Name: "write_file", Description: "Write a file", ReadOnly: false, RiskLevel: "medium",
		Call: func(ctx context.Context, args map[string]any) *tools.Result {
			return &tools.Result{Ok: true, Status: "ok"}
		},
	})
	return reg
}

// --- allow_all mode ---

func TestPolicy_AllowAll_ReadOnlyTool(t *testing.T) {
	p := &Policy{Mode: "allow_all"}
	reg := testPolicyRegistry()
	execute, requireApproval, decision := p.Evaluate("list_servers", reg)
	if !execute {
		t.Fatal("allow_all should execute read-only tools")
	}
	if requireApproval {
		t.Fatal("allow_all should not require approval for read-only tools")
	}
	if decision != "approved" {
		t.Fatalf("expected approved, got %s", decision)
	}
}

func TestPolicy_AllowAll_WriteTool(t *testing.T) {
	p := &Policy{Mode: "allow_all"}
	reg := testPolicyRegistry()
	execute, requireApproval, decision := p.Evaluate("restart_service", reg)
	if !execute {
		t.Fatal("allow_all should execute write tools")
	}
	if requireApproval {
		t.Fatal("allow_all should not require approval even for write tools")
	}
	if decision != "approved" {
		t.Fatalf("expected approved, got %s", decision)
	}
}

// --- deny_all mode ---

func TestPolicy_DenyAll_ReadOnlyTool(t *testing.T) {
	p := &Policy{Mode: "deny_all"}
	reg := testPolicyRegistry()
	execute, requireApproval, decision := p.Evaluate("list_servers", reg)
	if execute {
		t.Fatal("deny_all should not execute any tool")
	}
	if requireApproval {
		t.Fatal("deny_all should not require approval, just deny")
	}
	if decision != "denied" {
		t.Fatalf("expected denied, got %s", decision)
	}
}

func TestPolicy_DenyAll_WriteTool(t *testing.T) {
	p := &Policy{Mode: "deny_all"}
	reg := testPolicyRegistry()
	execute, _, decision := p.Evaluate("restart_service", reg)
	if execute {
		t.Fatal("deny_all should not execute write tools")
	}
	if decision != "denied" {
		t.Fatalf("expected denied, got %s", decision)
	}
}

// --- manual mode ---

func TestPolicy_Manual_ReadOnlyTool_AutoApproved(t *testing.T) {
	p := &Policy{Mode: "manual"}
	reg := testPolicyRegistry()
	execute, requireApproval, decision := p.Evaluate("list_servers", reg)
	if !execute {
		t.Fatal("manual mode should auto-approve read-only tools")
	}
	if requireApproval {
		t.Fatal("read-only tools should not require approval in manual mode")
	}
	if decision != "approved" {
		t.Fatalf("expected approved, got %s", decision)
	}
}

func TestPolicy_Manual_WriteTool_NotInDenyOrRequire_AutoApproved(t *testing.T) {
	p := &Policy{Mode: "manual"}
	reg := testPolicyRegistry()
	// write_file is not in Deny or RequireManual lists → auto-approved
	execute, requireApproval, decision := p.Evaluate("write_file", reg)
	if !execute {
		t.Fatal("manual mode should auto-approve write tools not in deny/require lists")
	}
	if requireApproval {
		t.Fatal("write tool not in deny/require lists should not need approval")
	}
	if decision != "approved" {
		t.Fatalf("expected approved, got %s", decision)
	}
}

func TestPolicy_Manual_WriteTool_InDeny_RequiresApproval(t *testing.T) {
	p := &Policy{Mode: "manual", Deny: []string{"restart_service"}}
	reg := testPolicyRegistry()
	execute, requireApproval, decision := p.Evaluate("restart_service", reg)
	if execute {
		t.Fatal("tool in Deny list should not be executed")
	}
	if !requireApproval {
		t.Fatal("tool in Deny list should require approval (pending)")
	}
	if decision != "pending" {
		t.Fatalf("expected pending, got %s", decision)
	}
}

func TestPolicy_Manual_WriteTool_InRequireManual_RequiresApproval(t *testing.T) {
	p := &Policy{Mode: "manual", RequireManual: []string{"restart_service"}}
	reg := testPolicyRegistry()
	execute, requireApproval, decision := p.Evaluate("restart_service", reg)
	if execute {
		t.Fatal("tool in RequireManual list should not be auto-executed")
	}
	if !requireApproval {
		t.Fatal("tool in RequireManual list should require approval")
	}
	if decision != "pending" {
		t.Fatalf("expected pending, got %s", decision)
	}
}

// --- unknown mode ---

func TestPolicy_UnknownMode_Denied(t *testing.T) {
	p := &Policy{Mode: "unknown_mode"}
	reg := testPolicyRegistry()
	execute, _, decision := p.Evaluate("list_servers", reg)
	if execute {
		t.Fatal("unknown mode should deny all tools")
	}
	if decision != "denied" {
		t.Fatalf("expected denied, got %s", decision)
	}
}

// --- unknown tool ---

func TestPolicy_Manual_UnknownTool_NotReadOnly(t *testing.T) {
	p := &Policy{Mode: "manual"}
	reg := testPolicyRegistry()
	// Unknown tool is not in registry → isReadOnly returns false
	execute, requireApproval, decision := p.Evaluate("nonexistent_tool", reg)
	if !execute {
		t.Fatal("unknown tool not in Deny/RequireManual should be auto-approved in manual mode")
	}
	if requireApproval {
		t.Fatal("unknown tool not in Deny/RequireManual should not require approval")
	}
	if decision != "approved" {
		t.Fatalf("expected approved, got %s", decision)
	}
}

// --- contains helper ---

func TestContains(t *testing.T) {
	s := []string{"alpha", "beta", "gamma"}
	if !contains(s, "beta") {
		t.Fatal("should find beta")
	}
	if contains(s, "delta") {
		t.Fatal("should not find delta")
	}
	if contains(nil, "anything") {
		t.Fatal("nil slice should not contain anything")
	}
	if contains([]string{}, "anything") {
		t.Fatal("empty slice should not contain anything")
	}
}