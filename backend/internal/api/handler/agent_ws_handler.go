package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/aicenter/aicenter/internal/ai"
	"github.com/aicenter/aicenter/internal/runtime/engine"
	"github.com/aicenter/aicenter/internal/runtime/tools"
	"github.com/aicenter/aicenter/internal/service"
	"github.com/gin-gonic/gin"
)

// ChatRunRequest is the request body for POST /agents/sessions/:id/run.
type ChatRunRequest struct {
	Query string `json:"query" binding:"required"`
}

// ChatEvent is the SSE envelope emitted by ChatRun.
type ChatEvent struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// ChatRun starts an agent run for the given session and streams events back as SSE.
// It is mounted under the protected route group; the caller's user_id is injected
// by JWT middleware.
func (h *AgentHandler) ChatRun(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "session id is required"})
		return
	}

	var req ChatRunRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Query) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "query is required"})
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	flusher, _ := c.Writer.(http.Flusher)

	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if flusher != nil {
					_, _ = c.Writer.Write([]byte(": keep-alive\n\n"))
					flusher.Flush()
				}
			}
		}
	}()

	sess, err := h.svc.GetSession(sessionID)
	if err != nil {
		_ = sseWrite(c.Writer, "error", map[string]string{"message": "session not found: " + err.Error()})
		return
	}
	agent, err := h.svc.GetAgent(sess.AgentID)
	if err != nil {
		_ = sseWrite(c.Writer, "error", map[string]string{"message": "agent not found: " + err.Error()})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		userID = sess.UserID
	}

	h.svc.AppendMessage(&service.AppendMessageReq{
		SessionID: sessionID,
		Role:      "user",
		Content:   req.Query,
	})

	exec := engine.New(h.svc.ToolRegistry(), agent)
	_ = h.svc.ToolRegistry().Resolve(agent.Tools)

	var mu sync.Mutex
	var toolRuns []engine.ToolRun

	emit := func(evt ChatEvent) error {
		mu.Lock()
		defer mu.Unlock()
		return sseWrite(c.Writer, evt.Type, evt.Data)
	}

	plannerFn := func(prompt string) (string, error) {
		providers, err := h.aiSvc.ListProviders()
		if err != nil || len(providers) == 0 {
			return defaultPlanText(prompt)
		}
		for _, p := range providers {
			if !p.IsEnabled {
				continue
			}
			if ai.ProviderType(p.APIType) != ai.ProviderOpenAICompatible {
				continue
			}
			client, err := h.aiSvc.GetProviderClient(p.ID)
			if err != nil {
				continue
			}
			out, err := syncLLM(client, agent.ModelID, prompt)
			if err == nil && out != "" {
				return out, nil
			}
		}
		return defaultPlanText(prompt)
	}

	plan, err := exec.RunPlanner(ctx, agent.SystemPrompt, req.Query, plannerFn)
	if err != nil {
		_ = emit(ChatEvent{Type: "error", Data: map[string]string{"message": "planner error: " + err.Error()}})
		return
	}

	_ = emit(ChatEvent{Type: "plan", Data: engine.Plan{Text: plan.Text, ToolCalls: plan.ToolCalls}})

	h.svc.AppendMessage(&service.AppendMessageReq{
		SessionID: sessionID,
		Role:      "assistant",
		Content:   plan.Text,
		Metadata:  map[string]any{"phase": "plan"},
	})

	for i, call := range plan.ToolCalls {
		if ctx.Err() != nil {
			_ = emit(ChatEvent{Type: "error", Data: map[string]string{"message": "client disconnected"}})
			return
		}

		ok, needsApproval, _ := exec.Decide(call)

		if needsApproval {
			dry := exec.ExecuteTool(ctx, call, true)
			approv := exec.BuildApprovalRequest(userID, call.Name, call.Args, dry)
			if err := h.svc.CreateApproval(approv); err != nil {
				_ = emit(ChatEvent{Type: "error", Data: map[string]string{"message": "approval create failed: " + err.Error()}})
				return
			}
			_ = emit(ChatEvent{Type: "approval_required", Data: approv})

			toolRuns = append(toolRuns, engine.ToolRun{Name: call.Name, Args: call.Args, Result: &tools.Result{Ok: false, Status: "pending_approval"}})
			h.svc.AppendMessage(&service.AppendMessageReq{
				SessionID: sessionID,
				Role:      "tool",
				ToolName:  call.Name,
				ToolArgs:  call.Args,
				ToolResult: map[string]any{
					"status":         "pending_approval",
					"approval_id":    approv.ID,
					"dry_run_result": dry,
				},
				Metadata: map[string]any{"turn": i},
			})
			continue
		}

		if !ok {
			toolRuns = append(toolRuns, engine.ToolRun{Name: call.Name, Args: call.Args, Result: &tools.Result{Ok: false, Status: "denied"}})
			_ = emit(ChatEvent{Type: "tool_result", Data: engine.ToolRun{Name: call.Name, Args: call.Args, Result: &tools.Result{Ok: false, Status: "denied"}}})
			continue
		}

		_ = emit(ChatEvent{Type: "tool_run", Data: map[string]any{"name": call.Name, "args": call.Args}})
		res := exec.ExecuteTool(ctx, call, false)

		h.svc.AppendMessage(&service.AppendMessageReq{
			SessionID: sessionID,
			Role:      "tool",
			ToolName:  call.Name,
			ToolArgs:  call.Args,
			ToolResult: map[string]any{
				"status":   res.Status,
				"ok":       res.Ok,
				"message":  res.Message,
				"payload":  res.Payload,
			},
			Metadata: map[string]any{"turn": i},
		})

		run := engine.ToolRun{Name: call.Name, Args: call.Args, Result: res}
		toolRuns = append(toolRuns, run)
		_ = emit(ChatEvent{Type: "tool_result", Data: run})
	}

	finalText := plan.Text
	if finalText == "" {
		finalText = "done"
	}
	_ = emit(ChatEvent{Type: "final", Data: engine.Result{Text: finalText, ToolRuns: toolRuns, Final: true}})

	_ = sseWrite(c.Writer, "done", "[DONE]")
}

func sseWrite(w gin.ResponseWriter, event string, data any) error {
	b, _ := json.Marshal(data)
	_, _ = w.Write([]byte("event: " + event + "\ndata: " + string(b) + "\n\n"))
	if f, ok := w.(interface{ Flush() }); ok {
		f.Flush()
	}
	return nil
}

func defaultPlanText(prompt string) (string, error) {
	lower := strings.ToLower(prompt)
	if strings.Contains(lower, "restart") || strings.Contains(lower, "nginx") || strings.Contains(lower, "stop") || strings.Contains(lower, "kill") {
		return `{"text":"I will restart the service.","tool_calls":[{"name":"restart_service","args":{"host":"web-01"}}]}`, nil
	}
	if strings.Contains(lower, "server") || strings.Contains(lower, "list") {
		return `{"text":"I will list the managed servers.","tool_calls":[{"name":"list_servers","args":{"limit":20}}]}`, nil
	}
	if strings.Contains(lower, "model") {
		return `{"text":"Here are the available models.","tool_calls":[{"name":"list_models","args":{}}]}`, nil
	}
	return `{"text":"I can help. What would you like to do?"}`, nil
}

func syncLLM(client ai.Client, modelID, prompt string) (string, error) {
	var buf strings.Builder
	streamer := ai.NewStreamEvents(&buf)
	if err := client.ChatCompletion(context.Background(), ai.ChatRequest{
		Model:    modelID,
		Messages: []ai.Message{{Role: "user", Content: prompt}},
		Stream:   true,
		Streamer: streamer,
	}); err != nil {
		return "", err
	}
	raw := buf.String()
	var out strings.Builder
	for _, line := range strings.Split(raw, "\n") {
		if strings.HasPrefix(line, "data: ") {
			out.WriteString(strings.TrimPrefix(line, "data: "))
		}
	}
	return out.String(), nil
}
