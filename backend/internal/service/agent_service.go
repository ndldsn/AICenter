package service

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/aicenter/aicenter/internal/runtime/approval"
	"github.com/aicenter/aicenter/internal/runtime/engine"
	"github.com/aicenter/aicenter/internal/runtime/tools"
	"github.com/aicenter/aicenter/internal/models"
	"github.com/aicenter/aicenter/internal/repository"
)

// AgentService orchestrates agents, sessions, messages, approvals and audit.
type AgentService struct {
	mu           sync.Mutex
	agentRepo    *repository.AgentRepository
	sessionRepo  *repository.AgentSessionRepository
	msgRepo      *repository.AgentMessageRepository
	approvalRepo *repository.ApprovalRepository
	auditRepo    *repository.AuditRepository
	toolReg      *tools.Registry
	approvalQ    *approval.Queue
	llmCall      func(modelID, prompt string) (string, error)
	// notify is an optional Phase 7 hook for approval notifications.
	notify func(eventType, title, severity, message string, data map[string]string)
}

func NewAgentService(
	agentRepo *repository.AgentRepository,
	sessionRepo *repository.AgentSessionRepository,
	msgRepo *repository.AgentMessageRepository,
	approvalRepo *repository.ApprovalRepository,
	auditRepo *repository.AuditRepository,
	toolReg *tools.Registry,
	llmCall func(modelID, prompt string) (string, error),
) *AgentService {
	return &AgentService{
		agentRepo:    agentRepo,
		sessionRepo:  sessionRepo,
		msgRepo:      msgRepo,
		approvalRepo: approvalRepo,
		auditRepo:    auditRepo,
		toolReg:      toolReg,
		approvalQ:    approval.NewQueue(),
		llmCall:      llmCall,
	}
}

// SetNotifier wires the Phase 7 notification hook (called on approval events).
func (s *AgentService) SetNotifier(fn func(eventType, title, severity, message string, data map[string]string)) {
	s.notify = fn
}

func (s *AgentService) ListAgents(enabled *bool) ([]models.Agent, error) {
	return s.agentRepo.List(enabled)
}
func (s *AgentService) GetAgent(id string) (*models.Agent, error) { return s.agentRepo.Get(id) }
func (s *AgentService) CreateAgent(a *models.Agent) error         { return s.agentRepo.Create(a) }
func (s *AgentService) UpdateAgent(a *models.Agent) error         { return s.agentRepo.Update(a) }
func (s *AgentService) DeleteAgent(id string) error               { return s.agentRepo.Delete(id) }

func (s *AgentService) CreateSession(req *CreateSessionReq) (*models.AgentSession, error) {
	_, err := s.agentRepo.Get(req.AgentID)
	if err != nil {
		return nil, err
	}
	sess := &models.AgentSession{
		AgentID:  req.AgentID,
		UserID:   req.UserID,
		ServerID: req.ServerID,
		Title:    titleFrom(req.Query),
	}
	if err := s.sessionRepo.Create(sess); err != nil {
		return nil, err
	}
	_ = s.msgRepo.Append(&models.AgentMessage{
		SessionID: sess.ID, Role: "user", Content: req.Query,
	})
	return sess, nil
}

func (s *AgentService) ListSessions(agentID, status string) ([]models.AgentSession, error) {
	return s.sessionRepo.List(agentID, status)
}

func (s *AgentService) GetSession(id string) (*models.AgentSession, error) {
	return s.sessionRepo.Get(id)
}

func (s *AgentService) GetSessionMessages(id string) ([]models.AgentMessage, error) {
	return s.msgRepo.ListBySession(id, 200)
}

func (s *AgentService) SendToSession(sessionID string, userMessage string, userID string) (SessionResult, error) {
	sess, err := s.sessionRepo.Get(sessionID)
	if err != nil {
		return SessionResult{}, err
	}
	agent, err := s.agentRepo.Get(sess.AgentID)
	if err != nil {
		return SessionResult{}, err
	}
	_ = s.msgRepo.Append(&models.AgentMessage{
		SessionID: sessionID, Role: "user", Content: userMessage,
	})

	exec := engine.New(s.toolReg, agent)
	result := SessionResult{SessionID: sessionID, AgentID: agent.ID, UserID: userID}

	plannerFn := func(prompt string) (string, error) {
		return s.llmCall(agent.ModelID, prompt)
	}

	_, err = exec.Run(context.Background(), agent.SystemPrompt, userMessage, plannerFn, engine.RunOptions{
		MaxIterations: agent.MaxIterations,
		OnStep: func(step engine.RunStep) error {
			// Persist the planner's text as an assistant message each turn.
			_ = s.msgRepo.Append(&models.AgentMessage{
				SessionID: sessionID, Role: "assistant", Content: step.Plan.Text,
				Metadata: jsonMarshal(map[string]any{"phase": "plan", "done": step.Done}),
			})
			if step.Plan.Text != "" {
				result.PlannedText = step.Plan.Text
			}

			for i, run := range step.ToolRuns {
				result.ToolRuns = append(result.ToolRuns, ToolRun{
					Name: run.Name, Args: run.Args, Result: run.Result,
				})
				_ = s.msgRepo.Append(&models.AgentMessage{
					SessionID: sessionID, Role: "tool", ToolName: run.Name,
					ToolArgs: run.Args, ToolResult: jsonMarshal(run.Result),
					Metadata: jsonMarshal(map[string]any{"turn": i}),
				})
				if run.Result != nil && run.Result.Ok {
					_ = s.auditRepo.Record(&repository.AuditEntry{
						Username:       userID,
						Action:         "tool_executed",
						ResourceType:   "tool",
						ResourceID:     run.Name,
						ResourceName:   run.Name,
						AgentSessionID: sessionID,
						StatusCode:     200,
					})
				}
			}

			// A tool needs human approval: create the request, notify, and stop.
			if step.Approval != nil {
				step.Approval.RequestedBy = userID
				if err := s.approvalRepo.Create(step.Approval); err != nil {
					return err
				}
				result.Approval = step.Approval
				result.RequiresApproval = true
				_ = s.msgRepo.Append(&models.AgentMessage{
					SessionID: sessionID, Role: "tool", ToolName: step.Approval.ToolName,
					ToolArgs:   step.Approval.ToolArgs,
					ToolResult: jsonMarshal(map[string]any{
						"status":      "pending_approval",
						"approval_id": step.Approval.ID,
					}),
				})
				if s.notify != nil {
					s.notify("approval.requested", "需要审批: "+step.Approval.ToolName, "warning",
						"Agent 请求执行工具 "+step.Approval.ToolName+"，等待审批。",
						map[string]string{
							"approval_id": step.Approval.ID,
							"tool":        step.Approval.ToolName,
							"session_id":  sessionID,
							"user_id":     userID,
						})
				}
			}
			return nil
		},
	})
	if err != nil {
		sess.Status = "failed"
		_ = s.sessionRepo.UpdateStatus(sessionID, "failed")
		return SessionResult{}, err
	}

	if result.RequiresApproval {
		sess.Status = "paused"
		_ = s.sessionRepo.UpdateStatus(sessionID, "paused")
	} else {
		sess.Status = "completed"
		_ = s.sessionRepo.UpdateStatus(sessionID, "completed")
	}
	return result, nil
}

type CreateSessionReq struct {
	AgentID  string `json:"agent_id"`
	UserID   string `json:"user_id"`
	ServerID string `json:"server_id"`
	Query    string `json:"query"`
}

type SessionResult struct {
	SessionID        string                  `json:"session_id"`
	AgentID          string                  `json:"agent_id"`
	UserID           string                  `json:"user_id"`
	PlannedText      string                  `json:"planned_text"`
	ToolRuns         []ToolRun               `json:"tool_runs"`
	Approval         *models.ApprovalRequest `json:"approval,omitempty"`
	RequiresApproval bool                    `json:"requires_approval"`
}

type ToolRun struct {
	Name   string `json:"name"`
	Args   json.RawMessage
	Result *tools.Result `json:"result"`
}

func (s *AgentService) Approve(approvalID, approvedBy string) error {
	if err := s.approvalRepo.Resolve(approvalID, string(models.ApprovalApproved), approvedBy); err != nil {
		return err
	}
	s.approvalQ.Resolve(approvalID, models.ApprovalApproved)
	if s.notify != nil {
		s.notify("approval.resolved", "审批已通过: "+approvalID, "info",
			"审批请求 "+approvalID+" 已被批准。",
			map[string]string{
				"approval_id": approvalID,
				"result":      "approved",
				"by":          approvedBy,
			})
	}
	return nil
}

func (s *AgentService) Reject(approvalID string) error {
	if err := s.approvalRepo.Resolve(approvalID, string(models.ApprovalRejected), ""); err != nil {
		return err
	}
	s.approvalQ.Resolve(approvalID, models.ApprovalRejected)
	if s.notify != nil {
		s.notify("approval.resolved", "审批已拒绝: "+approvalID, "info",
			"审批请求 "+approvalID+" 已被拒绝。",
			map[string]string{"approval_id": approvalID, "result": "rejected"})
	}
	return nil
}

func (s *AgentService) ListApprovals(status string) ([]models.ApprovalRequest, error) {
	return s.approvalRepo.List(status)
}

func (s *AgentService) GetApproval(id string) (*models.ApprovalRequest, error) {
	return s.approvalRepo.Get(id)
}

func (s *AgentService) ListAudit(limit, offset int) ([]repository.AuditEntry, error) {
	return s.auditRepo.List(limit, offset)
}

func (s *AgentService) ToolRegistry() *tools.Registry { return s.toolReg }

func (s *AgentService) CreateApproval(req *models.ApprovalRequest) error {
	return s.approvalRepo.Create(req)
}

// AppendMessageReq captures the minimal fields needed to append a message.
type AppendMessageReq struct {
	SessionID  string
	Role       string
	Content    string
	ToolName   string
	ToolArgs   json.RawMessage
	ToolResult any
	Metadata   map[string]any
}

func (s *AgentService) AppendMessage(req *AppendMessageReq) {
	if req == nil || strings.TrimSpace(req.SessionID) == "" {
		return
	}
	_ = s.msgRepo.Append(&models.AgentMessage{
		SessionID:  req.SessionID,
		Role:       req.Role,
		Content:    req.Content,
		ToolName:   req.ToolName,
		ToolArgs:   toRaw(req.ToolArgs),
		ToolResult: toRaw(req.ToolResult),
		Metadata:   toRaw(req.Metadata),
	})
}

func toRaw(v any) json.RawMessage {
	if v == nil {
		return nil
	}
	b, _ := json.Marshal(v)
	return b
}

func titleFrom(q string) string {
	if q == "" {
		return "Untitled session"
	}
	parts := strings.Fields(q)
	if len(parts) > 8 {
		parts = parts[:8]
	}
	return strings.Join(parts, " ")
}

func jsonMarshal(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

var _ = time.Second
