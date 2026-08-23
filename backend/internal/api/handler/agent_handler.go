package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/aicenter/aicenter/internal/models"
	"github.com/aicenter/aicenter/internal/service"
)

type AgentHandler struct {
	svc   *service.AgentService
	aiSvc *service.AIService
}

func NewAgentHandler(svc *service.AgentService) *AgentHandler {
	return &AgentHandler{svc: svc}
}

func (h *AgentHandler) WithAI(svc *service.AIService) *AgentHandler {
	h.aiSvc = svc
	return h
}

func (h *AgentHandler) ListAgents(c *gin.Context) {
	enabledRaw := c.Query("enabled")
	var enabled *bool
	if enabledRaw != "" {
		v := enabledRaw == "true" || enabledRaw == "1"
		enabled = &v
	}
	agents, err := h.svc.ListAgents(enabled)
	if err != nil {
		writeErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	write(c, gin.H{"items": agents, "total": len(agents)})
}

func (h *AgentHandler) GetAgent(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		writeErr(c, http.StatusBadRequest, "id is required")
		return
	}
	a, err := h.svc.GetAgent(id)
	if err != nil {
		writeErr(c, http.StatusNotFound, err.Error())
		return
	}
	write(c, a)
}

func (h *AgentHandler) CreateAgent(c *gin.Context) {
	a := &models.Agent{}
	if err := c.ShouldBindJSON(a); err != nil {
		writeErr(c, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	if a.Name == "" || a.ModelID == "" {
		writeErr(c, http.StatusBadRequest, "name and model_id are required")
		return
	}
	if a.Temperature <= 0 {
		a.Temperature = 0.7
	}
	if a.MaxTokens <= 0 {
		a.MaxTokens = 4096
	}
	if a.MaxIterations <= 0 {
		a.MaxIterations = 10
	}
	if a.ToolPermissionMode == "" {
		a.ToolPermissionMode = "manual"
	}
	if a.Tools == nil {
		a.Tools = []string{}
	}
	if a.RequireApprovalFor == nil {
		a.RequireApprovalFor = []string{}
	}
	if a.CreatedBy == "" {
		a.CreatedBy = c.GetString("user_id")
	}
	if err := h.svc.CreateAgent(a); err != nil {
		writeErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	write(c, a)
}

func (h *AgentHandler) UpdateAgent(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		writeErr(c, http.StatusBadRequest, "id is required")
		return
	}
	a := &models.Agent{ID: id}
	if err := c.ShouldBindJSON(a); err != nil {
		writeErr(c, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	if err := h.svc.UpdateAgent(a); err != nil {
		writeErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	a, err := h.svc.GetAgent(id)
	if err != nil {
		writeErr(c, http.StatusNotFound, err.Error())
		return
	}
	write(c, a)
}

func (h *AgentHandler) DeleteAgent(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		writeErr(c, http.StatusBadRequest, "id is required")
		return
	}
	if err := h.svc.DeleteAgent(id); err != nil {
		writeErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	write(c, gin.H{"ok": true})
}

func (h *AgentHandler) CreateSession(c *gin.Context) {
	var req service.CreateSessionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	if req.AgentID == "" {
		writeErr(c, http.StatusBadRequest, "agent_id is required")
		return
	}
	if req.UserID == "" {
		req.UserID = c.GetString("user_id")
	}
	sess, err := h.svc.CreateSession(&req)
	if err != nil {
		writeErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	write(c, sess)
}

func (h *AgentHandler) ListSessions(c *gin.Context) {
	agentID := c.Query("agent_id")
	status := c.Query("status")
	sessions, err := h.svc.ListSessions(agentID, status)
	if err != nil {
		writeErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	write(c, gin.H{"items": sessions, "total": len(sessions)})
}

func (h *AgentHandler) GetSession(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		writeErr(c, http.StatusBadRequest, "id is required")
		return
	}
	sess, err := h.svc.GetSession(id)
	if err != nil {
		writeErr(c, http.StatusNotFound, err.Error())
		return
	}
	messages, err := h.svc.GetSessionMessages(id)
	if err != nil {
		writeErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	write(c, gin.H{"session": sess, "messages": messages})
}

func (h *AgentHandler) SendMessage(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		writeErr(c, http.StatusBadRequest, "session id is required")
		return
	}
	var body struct {
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		writeErr(c, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	if body.Message == "" {
		writeErr(c, http.StatusBadRequest, "message is required")
		return
	}
	userID := c.GetString("user_id")
	result, err := h.svc.SendToSession(sessionID, body.Message, userID)
	if err != nil {
		writeErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	write(c, result)
}

func (h *AgentHandler) ListApprovals(c *gin.Context) {
	status := c.Query("status")
	approvals, err := h.svc.ListApprovals(status)
	if err != nil {
		writeErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	write(c, gin.H{"items": approvals, "total": len(approvals)})
}

func (h *AgentHandler) GetApproval(c *gin.Context) {
	id := c.Param("id")
	a, err := h.svc.GetApproval(id)
	if err != nil {
		writeErr(c, http.StatusNotFound, err.Error())
		return
	}
	write(c, a)
}

func (h *AgentHandler) Approve(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("user_id")
	if err := h.svc.Approve(id, userID); err != nil {
		writeErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	write(c, gin.H{"ok": true})
}

func (h *AgentHandler) Reject(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Reject(id); err != nil {
		writeErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	write(c, gin.H{"ok": true})
}

func (h *AgentHandler) ListAudit(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	if limit <= 0 {
		limit = 50
	}
	audit, err := h.svc.ListAudit(limit, offset)
	if err != nil {
		writeErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	write(c, gin.H{"items": audit, "total": len(audit)})
}

func write(c *gin.Context, v any) {
	c.JSON(200, v)
}

func writeErr(c *gin.Context, status int, msg string) {
	write(c, gin.H{"code": status, "error": msg, "detail": ""})
}

// helper keepalive
var _ = strings.TrimSpace

// helper json
var _ = json.Marshal

// helper fmt
var _ = fmt.Sprintf
