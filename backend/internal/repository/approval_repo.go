package repository

import (
	"encoding/json"
	"errors"
	"strings"

	"database/sql"

	"github.com/aicenter/aicenter/internal/models"
	"github.com/google/uuid"
)

var ErrApprovalNotFound = errors.New("approval request not found")

type ApprovalRepository struct{ db *sql.DB }

func NewApprovalRepository(db *sql.DB) *ApprovalRepository {
	return &ApprovalRepository{db: db}
}

func (r *ApprovalRepository) Create(req *models.ApprovalRequest) error {
	req.ID = uuid.New().String()
	var argsJSON string
	if req.ToolArgs != nil {
		argsJSON = string(req.ToolArgs)
	}
	var dryRun string
	if req.DryRunResult != nil {
		dryRun = string(req.DryRunResult)
	}
	_, err := r.db.Exec(`
		INSERT INTO approval_requests (id, request_type, status, requested_by, tool_name, tool_args, risk_level, dry_run_result, created_at)
		VALUES (?, 'tool_approval', 'pending', ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		req.ID, req.RequestedBy, req.ToolName, argsJSON, req.RiskLevel, dryRun)
	return err
}

func (r *ApprovalRepository) Get(id string) (*models.ApprovalRequest, error) {
	var a models.ApprovalRequest
	var requestedBy, toolArgs, dryRun, approvedBy sql.NullString
	err := r.db.QueryRow(`
		SELECT id, request_type, status, requested_by, tool_name, tool_args, risk_level, dry_run_result, approved_by, created_at
		FROM approval_requests WHERE id=?`, id).Scan(
		&a.ID, &a.RequestType, &a.Status, &requestedBy, &a.ToolName,
		&toolArgs, &a.RiskLevel, &dryRun, &approvedBy, &a.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrApprovalNotFound
		}
		return nil, err
	}
	if requestedBy.Valid {
		a.RequestedBy = requestedBy.String
	}
	if toolArgs.Valid {
		a.ToolArgs = []byte(toolArgs.String)
	}
	if dryRun.Valid {
		a.DryRunResult = []byte(dryRun.String)
	}
	if approvedBy.Valid {
		a.ApprovedBy = approvedBy.String
	}
	return &a, nil
}

func (r *ApprovalRepository) List(status string) ([]models.ApprovalRequest, error) {
	var where string
	if status != "" {
		where = " WHERE status=" + status
	}
	rows, err := r.db.Query(`
		SELECT id, request_type, status, requested_by, tool_name, tool_args, risk_level, dry_run_result, approved_by, created_at
		FROM approval_requests` + where + ` ORDER BY created_at DESC LIMIT 100`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanApprovals(rows)
}

func scanApprovals(rows *sql.Rows) ([]models.ApprovalRequest, error) {
	var out []models.ApprovalRequest
	for rows.Next() {
		var a models.ApprovalRequest
		var requestedBy, toolArgs, dryRun, approvedBy sql.NullString
		if err := rows.Scan(
			&a.ID, &a.RequestType, &a.Status, &requestedBy, &a.ToolName,
			&toolArgs, &a.RiskLevel, &dryRun, &approvedBy, &a.CreatedAt); err != nil {
			return nil, err
		}
		if requestedBy.Valid {
			a.RequestedBy = requestedBy.String
		}
		if toolArgs.Valid {
			a.ToolArgs = []byte(toolArgs.String)
		}
		if dryRun.Valid {
			a.DryRunResult = []byte(dryRun.String)
		}
		if approvedBy.Valid {
			a.ApprovedBy = approvedBy.String
		}
		out = append(out, a)
	}
	return out, nil
}

func (r *ApprovalRepository) Resolve(id, status string, approvedBy string) error {
	args := []interface{}{status, id}
	if approvedBy != "" {
		_, err := r.db.Exec("UPDATE approval_requests SET status=?, approved_by=?, approved_at=CURRENT_TIMESTAMP WHERE id=?",
			status, approvedBy, id)
		return err
	}
	_, err := r.db.Exec("UPDATE approval_requests SET status=? WHERE id=?", args[0], args[1])
	_ = args
	return err
}

// AuditRepository records audit events linked to agent activity.
type AuditRepository struct{ db *sql.DB }

func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Record(entry *AuditEntry) error {
	id := uuid.New().String()
	_, err := r.db.Exec(`
		INSERT INTO audit_logs (id, user_id, username, action, resource_type, resource_id, resource_name, method, path, status_code, ip_address, user_agent, request_body, response_body, before_state, after_state, duration_ms, error_message, server_id, agent_session_id, approval_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		id, entry.UserID, entry.Username, entry.Action, entry.ResourceType,
		entry.ResourceID, entry.ResourceName, entry.Method, entry.Path,
		entry.StatusCode, entry.IP, entry.UserAgent, entry.RequestBody, entry.ResponseBody,
		entry.BeforeState, entry.AfterState, entry.DurationMs, entry.Error,
		"", entry.AgentSessionID, entry.ApprovalID)
	return err
}

func (r *AuditRepository) List(limit, offset int) ([]AuditEntry, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query("SELECT id, username, action, resource_type, resource_id, resource_name, method, path, status_code, agent_session_id, approval_id, created_at FROM audit_logs ORDER BY created_at DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AuditEntry
	for rows.Next() {
		var e AuditEntry
		var recordID, sID, aID sql.NullString
		if err := rows.Scan(
			&recordID, &e.Username, &e.Action, &e.ResourceType, &e.ResourceID, &e.ResourceName,
			&e.Method, &e.Path, &e.StatusCode, &sID, &aID, &e.CreatedAt); err != nil {
			return nil, err
		}
		if sID.Valid {
			e.AgentSessionID = sID.String
		}
		if aID.Valid {
			e.ApprovalID = aID.String
		}
		_ = recordID
		out = append(out, e)
	}
	return out, nil
}

type AuditEntry struct {
	Username       string `json:"username"`
	UserID         string `json:"user_id,omitempty"`
	Role           string `json:"role,omitempty"`
	Action         string `json:"action"`
	ResourceType   string `json:"resource_type"`
	ResourceID     string `json:"resource_id"`
	ResourceName   string `json:"resource_name,omitempty"`
	Method         string `json:"method,omitempty"`
	Path           string `json:"path,omitempty"`
	StatusCode     int    `json:"status_code,omitempty"`
	IP             string `json:"ip_address,omitempty"`
	UserAgent      string `json:"user_agent,omitempty"`
	RequestBody    string `json:"request_body,omitempty"`
	ResponseBody   string `json:"response_body,omitempty"`
	BeforeState    string `json:"before_state,omitempty"`
	AfterState     string `json:"after_state,omitempty"`
	DurationMs     int    `json:"duration_ms,omitempty"`
	Error          string `json:"error_message,omitempty"`
	AgentSessionID string `json:"agent_session_id,omitempty"`
	ApprovalID     string `json:"approval_id,omitempty"`
	CreatedAt      string `json:"created_at"`
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func jsonMarshal(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// jsonMerge combines two JSON object strings, preferring second's non-null values.
func jsonMerge(a, b string) string {
	if a == "" || a == "{}" {
		return b
	}
	if b == "" || b == "{}" {
		return a
	}
	return a + "," + strings.TrimSuffix(b, "}") + "}"
}
