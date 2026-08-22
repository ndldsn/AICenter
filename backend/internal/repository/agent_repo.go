package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/aicenter/aicenter/internal/models"
	"github.com/google/uuid"
)

var ErrAgentNotFound = errors.New("agent not found")
var ErrSessionNotFound = errors.New("session not found")

// AgentRepository persists agents rows.
type AgentRepository struct {
	db *sql.DB
}

func NewAgentRepository(db *sql.DB) *AgentRepository {
	return &AgentRepository{db: db}
}

func (r *AgentRepository) Create(a *models.Agent) error {
	a.ID = uuid.New().String()
	now := time.Now().Format("2006-01-02 15:04:05")
	toolsJSON, _ := json.Marshal(a.Tools)
	approvalJSON, _ := json.Marshal(a.RequireApprovalFor)
	_, err := r.db.Exec(`
		INSERT INTO agents (id, name, description, model_id, system_prompt, temperature, max_tokens, max_iterations, tools, tool_permission_mode, require_approval_for, is_enabled, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.Name, a.Description, a.ModelID, a.SystemPrompt,
		a.Temperature, a.MaxTokens, a.MaxIterations,
		string(toolsJSON), a.ToolPermissionMode, string(approvalJSON),
		a.IsEnabled, a.CreatedBy, now, now)
	return err
}

func (r *AgentRepository) Get(id string) (*models.Agent, error) {
	var a models.Agent
	var toolsJSON, approvalJSON string
	var enabled int
	var createdBy sql.NullString
	err := r.db.QueryRow(`
		SELECT id, name, description, model_id, system_prompt, temperature, max_tokens, max_iterations, tools, tool_permission_mode, require_approval_for, is_enabled, created_by, created_at, updated_at
		FROM agents WHERE id=?`, id).Scan(
		&a.ID, &a.Name, &a.Description, &a.ModelID, &a.SystemPrompt,
		&a.Temperature, &a.MaxTokens, &a.MaxIterations,
		&toolsJSON, &a.ToolPermissionMode, &approvalJSON,
		&enabled, &createdBy, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAgentNotFound
		}
		return nil, err
	}
	a.IsEnabled = enabled != 0
	if createdBy.Valid {
		a.CreatedBy = createdBy.String
	}
	_ = json.Unmarshal([]byte(toolsJSON), &a.Tools)
	_ = json.Unmarshal([]byte(approvalJSON), &a.RequireApprovalFor)
	return &a, nil
}

func (r *AgentRepository) List(enabled *bool) ([]models.Agent, error) {
	var where string
	args := []interface{}{}
	if enabled != nil {
		where = " WHERE is_enabled = ?"
		args = append(args, b2i(*enabled))
	}
	query := "SELECT id, name, description, model_id, system_prompt, temperature, max_tokens, max_iterations, tools, tool_permission_mode, require_approval_for, is_enabled, created_by, created_at, updated_at FROM agents" + where + " ORDER BY created_at DESC"
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAgents(rows)
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (r *AgentRepository) Update(a *models.Agent) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	toolsJSON, _ := json.Marshal(a.Tools)
	approvalJSON, _ := json.Marshal(a.RequireApprovalFor)
	_, err := r.db.Exec(`
		UPDATE agents SET name=?, description=?, model_id=?, system_prompt=?, temperature=?, max_tokens=?, max_iterations=?, tools=?, tool_permission_mode=?, require_approval_for=?, is_enabled=?, updated_at=?
		WHERE id=?`,
		a.Name, a.Description, a.ModelID, a.SystemPrompt,
		a.Temperature, a.MaxTokens, a.MaxIterations,
		string(toolsJSON), a.ToolPermissionMode, string(approvalJSON),
		a.IsEnabled, now, a.ID)
	return err
}

func (r *AgentRepository) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM agents WHERE id=?", id)
	return err
}

func scanAgents(rows *sql.Rows) ([]models.Agent, error) {
	var out []models.Agent
	for rows.Next() {
		var a models.Agent
		var toolsJSON, approvalJSON string
		var enabled int
		var createdBy sql.NullString
		if err := rows.Scan(
			&a.ID, &a.Name, &a.Description, &a.ModelID, &a.SystemPrompt,
			&a.Temperature, &a.MaxTokens, &a.MaxIterations,
			&toolsJSON, &a.ToolPermissionMode, &approvalJSON,
			&enabled, &createdBy, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		a.IsEnabled = enabled != 0
		if createdBy.Valid {
			a.CreatedBy = createdBy.String
		}
		_ = json.Unmarshal([]byte(toolsJSON), &a.Tools)
		_ = json.Unmarshal([]byte(approvalJSON), &a.RequireApprovalFor)
		out = append(out, a)
	}
	return out, nil
}

// AgentSessionRepository persists agent_sessions.
type AgentSessionRepository struct{ db *sql.DB }

func NewAgentSessionRepository(db *sql.DB) *AgentSessionRepository {
	return &AgentSessionRepository{db: db}
}

func (r *AgentSessionRepository) Create(s *models.AgentSession) error {
	s.ID = uuid.New().String()
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := r.db.Exec(`
		INSERT INTO agent_sessions (id, agent_id, user_id, server_id, title, status, token_input, token_output, started_at, created_at)
		VALUES (?, ?, ?, ?, ?, 'active', 0, 0, ?, ?)`,
		s.ID, s.AgentID, s.UserID, s.ServerID, s.Title, now, now)
	return err
}

func (r *AgentSessionRepository) Get(id string) (*models.AgentSession, error) {
	var s models.AgentSession
	var serverID, endedAt, contextSummary sql.NullString
	err := r.db.QueryRow(`
		SELECT id, agent_id, user_id, server_id, title, status, context_summary, token_input, token_output, started_at, ended_at, created_at
		FROM agent_sessions WHERE id=?`, id).Scan(
		&s.ID, &s.AgentID, &s.UserID, &serverID, &s.Title, &s.Status,
		&contextSummary, &s.TokenInput, &s.TokenOutput,
		&s.StartedAt, &endedAt, &s.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	if serverID.Valid {
		s.ServerID = serverID.String
	}
	if endedAt.Valid {
		s.EndedAt = endedAt.String
	}
	if contextSummary.Valid {
		s.ContextSummary = contextSummary.String
	}
	return &s, nil
}

func (r *AgentSessionRepository) List(agentID, status string) ([]models.AgentSession, error) {
	var args []interface{}
	parts := []string{}
	if agentID != "" {
		parts = append(parts, "agent_id=?")
		args = append(args, agentID)
	}
	if status != "" {
		parts = append(parts, "status=?")
		args = append(args, status)
	}
	where := ""
	if len(parts) > 0 {
		where = " WHERE " + strings.Join(parts, " AND ")
	}
	rows, err := r.db.Query(`
		SELECT id, agent_id, user_id, server_id, title, status, context_summary, token_input, token_output, started_at, ended_at, created_at
		FROM agent_sessions`+where+` ORDER BY started_at DESC LIMIT 50`,
		args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSessions(rows)
}

func (r *AgentSessionRepository) UpdateStatus(id, status string) error {
	_, err := r.db.Exec("UPDATE agent_sessions SET status=? WHERE id=?", status, id)
	return err
}

func scanSessions(rows *sql.Rows) ([]models.AgentSession, error) {
	var out []models.AgentSession
	for rows.Next() {
		var s models.AgentSession
		var serverID, endedAt, contextSummary sql.NullString
		if err := rows.Scan(
			&s.ID, &s.AgentID, &s.UserID, &serverID, &s.Title, &s.Status,
			&contextSummary, &s.TokenInput, &s.TokenOutput,
			&s.StartedAt, &endedAt, &s.CreatedAt); err != nil {
			return nil, err
		}
		if serverID.Valid {
			s.ServerID = serverID.String
		}
		if endedAt.Valid {
			s.EndedAt = endedAt.String
		}
		if contextSummary.Valid {
			s.ContextSummary = contextSummary.String
		}
		out = append(out, s)
	}
	return out, nil
}

// AgentMessageRepository persists messages.
type AgentMessageRepository struct{ db *sql.DB }

func NewAgentMessageRepository(db *sql.DB) *AgentMessageRepository {
	return &AgentMessageRepository{db: db}
}

func (r *AgentMessageRepository) Append(m *models.AgentMessage) error {
	m.ID = uuid.New().String()
	var toolArgs, toolResult, meta string
	if m.ToolArgs != nil {
		toolArgs = string(m.ToolArgs)
	}
	if m.ToolResult != nil {
		toolResult = string(m.ToolResult)
	}
	if m.Metadata != nil {
		meta = string(m.Metadata)
	}
	_, err := r.db.Exec(`
		INSERT INTO agent_messages (id, session_id, role, content, tool_call_id, tool_name, tool_args, tool_result, metadata, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		m.ID, m.SessionID, m.Role, m.Content,
		m.ToolCallID, m.ToolName, toolArgs, toolResult, meta)
	return err
}

func (r *AgentMessageRepository) ListBySession(sessionID string, limit int) ([]models.AgentMessage, error) {
	rows, err := r.db.Query(`
		SELECT id, session_id, role, content, tool_call_id, tool_name, tool_args, tool_result, metadata, created_at
		FROM agent_messages WHERE session_id=? ORDER BY created_at ASC
		LIMIT ?`, sessionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.AgentMessage
	for rows.Next() {
		var m models.AgentMessage
		var content, toolCallID, toolName, toolArgs, toolResult, meta sql.NullString
		if err := rows.Scan(
			&m.ID, &m.SessionID, &m.Role, &content,
			&toolCallID, &toolName, &toolArgs, &toolResult, &meta, &m.CreatedAt); err != nil {
			return nil, err
		}
		if content.Valid {
			m.Content = content.String
		}
		if toolCallID.Valid {
			m.ToolCallID = toolCallID.String
		}
		if toolName.Valid {
			m.ToolName = toolName.String
		}
		if toolArgs.Valid {
			m.ToolArgs = []byte(toolArgs.String)
		}
		if toolResult.Valid {
			m.ToolResult = []byte(toolResult.String)
		}
		if meta.Valid {
			m.Metadata = []byte(meta.String)
		}
		out = append(out, m)
	}
	return out, nil
}
