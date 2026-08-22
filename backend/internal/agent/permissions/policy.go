package permissions

import (
	"github.com/aicenter/aicenter/internal/agent/tools"
)

// Policy evaluates whether an agent may invoke a tool under the configured mode.
type Policy struct {
	Mode          string // "allow_all" | "deny_all" | "manual"
	Allow         []string
	Deny          []string
	RequireManual []string
}

// Evaluate returns: (execute, requireApproval, decision).
func (p *Policy) Evaluate(toolName string, reg *tools.Registry) (bool, bool, string) {
	switch p.Mode {
	case "allow_all":
		return true, false, "approved"
	case "deny_all":
		return false, false, "denied"
	case "manual":
		if p.isReadOnly(toolName, reg) {
			return true, false, "approved"
		}
		if contains(p.Deny, toolName) || contains(p.RequireManual, toolName) {
			return false, true, "pending"
		}
		return true, false, "approved"
	}
	return false, false, "denied"
}

func (p *Policy) isReadOnly(name string, reg *tools.Registry) bool {
	t, ok := reg.Get(name)
	if !ok {
		return false
	}
	return t.ReadOnly
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
