package models

import "encoding/json"

// Agent is a named AI persona with a bound LLM, system prompt, and tool set.
type Agent struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Description        string   `json:"description,omitempty"`
	ModelID            string   `json:"model_id"`
	SystemPrompt       string   `json:"system_prompt,omitempty"`
	Temperature        float64  `json:"temperature"`
	MaxTokens          int      `json:"max_tokens"`
	MaxIterations      int      `json:"max_iterations"`
	Tools              []string `json:"tools"`
	ToolPermissionMode string   `json:"tool_permission_mode"` // "allow_all" | "deny_all" | "manual"
	RequireApprovalFor []string `json:"require_approval_for"`
	IsEnabled          bool     `json:"is_enabled"`
	CreatedBy          string   `json:"created_by,omitempty"`
	CreatedAt          string   `json:"created_at"`
	UpdatedAt          string   `json:"updated_at"`
}

// AgentSession is a single conversation run for an agent against an optional target server.
type AgentSession struct {
	ID             string `json:"id"`
	AgentID        string `json:"agent_id"`
	UserID         string `json:"user_id"`
	ServerID       string `json:"server_id,omitempty"`
	Title          string `json:"title"`
	Status         string `json:"status"` // active | completed | failed | paused
	ContextSummary string `json:"context_summary,omitempty"`
	TokenInput     int    `json:"token_input"`
	TokenOutput    int    `json:"token_output"`
	StartedAt      string `json:"started_at"`
	EndedAt        string `json:"ended_at,omitempty"`
	CreatedAt      string `json:"created_at"`
}

// AgentMessage records one turn (user text, assistant text, or a tool call/result) within a session.
type AgentMessage struct {
	ID         string          `json:"id"`
	SessionID  string          `json:"session_id"`
	Role       string          `json:"role"` // user | assistant | tool
	Content    string          `json:"content,omitempty"`
	ToolCallID string          `json:"tool_call_id,omitempty"`
	ToolName   string          `json:"tool_name,omitempty"`
	ToolArgs   json.RawMessage `json:"tool_args,omitempty"`
	ToolResult json.RawMessage `json:"tool_result,omitempty"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
	CreatedAt  string          `json:"created_at"`
}

// ToolCall describes a single tool invocation requested by the agent LLM.
type ToolCall struct {
	Name string          `json:"name"`
	Args json.RawMessage `json:"args,omitempty"`
}

// ApprovalStatus enumerates the states of an approval request.
type ApprovalStatus string

const (
	ApprovalPending  ApprovalStatus = "pending"
	ApprovalApproved ApprovalStatus = "approved"
	ApprovalRejected ApprovalStatus = "rejected"
	ApprovalExpired  ApprovalStatus = "expired"
)

// ApprovalRequest captures a tool invocation that needs human approval.
type ApprovalRequest struct {
	ID           string          `json:"id"`
	RequestType  string          `json:"request_type"`
	Status       ApprovalStatus  `json:"status"`
	RequestedBy  string          `json:"requested_by,omitempty"`
	ToolName     string          `json:"tool_name"`
	ToolArgs     json.RawMessage `json:"tool_args,omitempty"`
	RiskLevel    string          `json:"risk_level"`
	DryRunResult json.RawMessage `json:"dry_run_result,omitempty"`
	ApprovedBy   string          `json:"approved_by,omitempty"`
	CreatedAt    string          `json:"created_at"`
}
