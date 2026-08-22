package service

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/aicenter/aicenter/internal/agent/approval"
	"github.com/aicenter/aicenter/internal/agent/runtime"
	"github.com/aicenter/aicenter/internal/agent/tools"
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

	exec := runtime.New(s.toolReg, agent)
	result := SessionResult{SessionID: sessionID, AgentID: agent.ID, UserID: userID}

	plannerFn := func(prompt string) (string, error) {
		return s.llmCall(agent.ModelID, prompt)
	}

	plan, err := exec.RunPlanner(context.Background(), agent.SystemPrompt, userMessage, plannerFn)
	if err != nil {
		return SessionResult{}, err
	}

	_ = s.msgRepo.Append(&models.AgentMessage{
		SessionID: sessionID, Role: "assistant", Content: plan.Text,
		Metadata: jsonMarshal(map[string]any{"phase": "plan"}),
	})
	result.PlannedText = plan.Text

	for i, call := range plan.ToolCalls {
		execute, needsApproval, _ := exec.Decide(call)
		if !execute {
			if needsApproval {
				dry := exec.ExecuteTool(context.Background(), call, true)
				approv := exec.BuildApprovalRequest(userID, call.Name, call.Args, dry)
				if err := s.approvalRepo.Create(approv); err != nil {
					return SessionResult{}, err
				}
				result.Approval = approv
				_ = s.msgRepo.Append(&models.AgentMessage{
					SessionID: sessionID, Role: "tool", ToolName: call.Name,
					ToolArgs: call.Args,
					ToolResult: jsonMarshal(map[string]any{
						"status":         "pending_approval",
						"approval_id":    approv.ID,
						"dry_run_result": dry,
					}),
					Metadata: jsonMarshal(map[string]any{"turn": i}),
				})
				result.RequiresApproval = true
				// Phase 7: notify approval.requested
				if s.notify != nil {
					data := map[string]string{
						"approval_id": approv.ID,
						"tool":        call.Name,
						"session_id":  sessionID,
						"user_id":     userID,
					}
					s.notify("approval.requested", "需要审批: "+call.Name, "warning",
						"Agent 请求执行工具 "+call.Name+"，等待审批。", data)
				}
				return result, nil
			}
			_ = s.msgRepo.Append(&models.AgentMessage{
				SessionID: sessionID, Role: "tool", ToolName: call.Name,
				ToolArgs:   call.Args,
				ToolResult: jsonMarshal(map[string]any{"status": "denied"}),
			})
			result.ToolRuns = append(result.ToolRuns, ToolRun{
				Name:   call.Name,
				Args:   call.Args,
				Result: &tools.Result{Ok: false, Status: "denied"},
			})
			continue
		}

		res := exec.ExecuteTool(context.Background(), call, false)
		_ = s.msgRepo.Append(&models.AgentMessage{
			SessionID: sessionID, Role: "tool", ToolName: call.Name,
			ToolArgs: call.Args, ToolResult: jsonMarshal(res),
		})
		_ = s.auditRepo.Record(&repository.AuditEntry{
			Username:       userID,
			Action:         "tool_executed",
			ResourceType:   "tool",
			ResourceID:     call.Name,
			ResourceName:   call.Name,
			AgentSessionID: sessionID,
			StatusCode:     200,
		})
		result.ToolRuns = append(result.ToolRuns, ToolRun{
			Name:   call.Name,
			Args:   call.Args,
			Result: res,
		})
	}

	sess.Status = "completed"
	_ = s.sessionRepo.UpdateStatus(sessionID, "completed")
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
		s.notify("approval.resolved", "审批已通过: "+approvalID, "info", "审批请求 "+approvalID+" 已被批准。",
			map[string]string{"approval_id": approvalID, "result": "approved", "by": approvedBy})
	}
	return nil
}

func (s *AgentService) Reject(approvalID string) error {
	if err := s.approvalRepo.Resolve(approvalID, string(models.ApprovalRejected), ""); err != nil {
		return err
	}
	s.approvalQ.Resolve(approvalID, models.ApprovalRejected)
	if s.notify != nil {
		s.notify("approval.resolved", "审批已拒绝: "+approvalID, "info", "审批请求 "+approvalID+" 已被拒绝。",
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
