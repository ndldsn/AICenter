package router

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/aicenter/aicenter/internal/runtime/tools"
	"github.com/aicenter/aicenter/internal/ai"
	"github.com/aicenter/aicenter/internal/api/handler"
	"github.com/aicenter/aicenter/internal/api/middleware"
	"github.com/aicenter/aicenter/internal/config"
	"github.com/aicenter/aicenter/internal/models"
	"github.com/aicenter/aicenter/internal/monitor"
	"github.com/aicenter/aicenter/internal/notifier"
	"github.com/aicenter/aicenter/internal/pkg/crypto"
	"github.com/aicenter/aicenter/internal/pkg/logger"
	"github.com/aicenter/aicenter/internal/repository"
	"github.com/aicenter/aicenter/internal/service"
	"github.com/aicenter/aicenter/internal/terminal"
	"github.com/aicenter/aicenter/internal/websocket"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Setup configures the Gin router
func Setup(cfg *config.Config, db *sql.DB, hub *websocket.Hub, log *zap.Logger) *gin.Engine {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Global middleware
	r.Use(middleware.Logger(log))
	r.Use(middleware.Recovery(log))
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())

	// API v1
	v1 := r.Group("/api/v1")
	// AuditMiddleware is inserted at the v1 group level so it covers
	// both public (login/register) and protected routes. It is
	// intentionally best-effort and never rejects a request.
	v1.Use(middleware.AuditMiddleware(db))

	// Health check (no auth, no audit)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "aicenter"})
	})

	// WebSocket endpoint (authenticated via ?token= query param)
	r.GET("/ws", func(c *gin.Context) {
		websocket.ServeWs(hub, cfg.Auth.Secret, c.Writer, c.Request)
	})

	// Protected routes: real JWT bearer auth.
	prot := v1.Group("/", middleware.JWTAuth(cfg.Auth.Secret))

	// Public auth routes (login/register/refresh). /auth/me is wired below under
	// protected routes so it can see the JWT-injected userID.
	authHandler := handler.NewAuthHandler(
		service.NewAuthService(
			repository.NewUserRepository(db),
			cfg,
		),
	)
	authHandler.RegisterPublicRoutes(v1)

	{
		authHandler.RegisterProtectedRoutes(prot)

		// Dashboard
		prot.GET("/dashboard", handleDashboard)

		// Servers - using real handler
		serverHandler := handler.NewServerHandler()
		serverHandler.RegisterRoutes(prot, middleware.JWTAuth(cfg.Auth.Secret))

		// Docker (Phase 3 - real handler)
		dockerHandler := handler.NewDockerHandler(service.NewDockerService(hub))
		dockerHandler.RegisterRoutes(prot, middleware.JWTAuth(cfg.Auth.Secret))

		// AI Providers & Models (Phase 4 - real handler)
		aiService := service.NewAIService(
			repository.NewAIProviderRepository(db),
			repository.NewAIModelRepository(db),
		)
		aiHandler := handler.NewAIHandler(aiService)
		prot.GET("/ai/providers", aiHandler.ListProviders)
		prot.GET("/ai/providers/:id", aiHandler.GetProvider)
		prot.POST("/ai/providers", aiHandler.CreateProvider)
		prot.PUT("/ai/providers/:id", aiHandler.UpdateProvider)
		prot.DELETE("/ai/providers/:id", aiHandler.DeleteProvider)
		prot.GET("/ai/models/:provider_id", aiHandler.ListModels)
		prot.POST("/ai/chat", aiHandler.Chat)

		// Agents
		agentService := service.NewAgentService(
			repository.NewAgentRepository(db),
			repository.NewAgentSessionRepository(db),
			repository.NewAgentMessageRepository(db),
			repository.NewApprovalRepository(db),
			repository.NewAuditRepository(db),
			tools.DefaultTools(),
			func(modelID, prompt string) (string, error) {
				return callLLM(aiService, modelID, prompt)
			},
		)
		agentHandler := handler.NewAgentHandler(agentService).WithAI(aiService)
		prot.GET("/agents", agentHandler.ListAgents)
		prot.POST("/agents", agentHandler.CreateAgent)
		prot.GET("/agents/:id", agentHandler.GetAgent)
		prot.PUT("/agents/:id", agentHandler.UpdateAgent)
		prot.DELETE("/agents/:id", agentHandler.DeleteAgent)
		prot.POST("/agents/:id/sessions", agentHandler.CreateSession)
		prot.GET("/agents/sessions", agentHandler.ListSessions)
		prot.GET("/agents/sessions/:id", agentHandler.GetSession)
		prot.POST("/agents/sessions/:id/messages", agentHandler.SendMessage)
		prot.POST("/agents/sessions/:id/run", agentHandler.ChatRun)

		prot.GET("/audit-logs", agentHandler.ListAudit)

		// Approvals
		prot.GET("/approvals", agentHandler.ListApprovals)
		prot.GET("/approvals/:id", agentHandler.GetApproval)
		prot.POST("/approvals/:id/approve", agentHandler.Approve)
		prot.POST("/approvals/:id/reject", agentHandler.Reject)

		// Tasks
		prot.GET("/tasks", handleListTasks)
		prot.POST("/tasks", handleCreateTask)
		prot.GET("/tasks/:id", handleGetTask)

		// Monitor (Phase 6 - real handler)
		monitorRepo := repository.NewMonitorRepository(db)
		engine := monitor.NewEngine(monitorRepo, monitor.SynthCollector{}, 15*time.Second)
		engine.Start()
		defer engine.Stop()
		monitorService := service.NewMonitorService(monitorRepo, engine)
		monitorHandler := handler.NewMonitorHandler(monitorService)
		monitorHandler.RegisterRoutes(prot, middleware.JWTAuth(cfg.Auth.Secret))

		// Notifications (Phase 7)
		notifRepo := repository.NewNotificationRepository(db)
		notifDispatcher := notifier.NewDispatcher(notifRepo, log)
		notifService := service.NewNotificationService(notifRepo, notifDispatcher)
		notifHandler := handler.NewNotificationHandler(notifService)
		notifHandler.RegisterRoutes(prot, middleware.JWTAuth(cfg.Auth.Secret))
		// Wire the alert engine to the notification dispatcher so fired alerts
		// trigger notifications (honouring each rule's bound channels).
		engine.SetNotifier(func(eventType, title, severity, message string, data map[string]string, channelIDs []string) {
			_ = notifService.Notify(eventType, title, severity, message, data, channelIDs)
		})
		// Wire agent approvals to the notification dispatcher.
		agentService.SetNotifier(func(eventType, title, severity, message string, data map[string]string) {
			_ = notifService.Notify(eventType, title, severity, message, data, nil)
		})

		// Web Terminal (完善与优化 Phase 7.1)
		terminalMgr := terminal.NewManager(log)
		terminalHandler := handler.NewTerminalHandler(terminalMgr, log, cfg.Auth.Secret)
		prot.POST("/terminal/sessions", terminalHandler.CreateSession)
		prot.GET("/terminal/sessions", terminalHandler.ListSessions)
		prot.POST("/terminal/sessions/:id/close", terminalHandler.CloseSession)
		// WebSocket bridge for an active terminal session.
		r.GET("/ws/terminal", func(c *gin.Context) {
			terminalHandler.Bridge(c.Writer, c.Request)
		})

		// Batch operations (多服务器批量操作)
		batchHandler := handler.NewBatchHandler(service.NewBatchService())
		batchHandler.RegisterRoutes(prot, middleware.JWTAuth(cfg.Auth.Secret))

		// Users (H2 batch 2b: real user CRUD gated on users.manage)
		usersHandler := handler.NewUsersHandler(service.NewUserService(repository.NewUserRepository(db)))
		usersHandler.RegisterRoutes(prot)

		// Roles - real RBAC management endpoints (H2 batch 2).
		rolesHandler := handler.NewRolesHandler(repository.NewRoleRepository(db))
		rolesHandler.RegisterRoutes(prot, middleware.JWTAuth(cfg.Auth.Secret))

		// Settings
		prot.GET("/settings", handleGetSettings)
		prot.PUT("/settings", handleUpdateSettings)
	}

	return r
}

// Placeholder handlers - these will be replaced with real implementations

func handleDashboard(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Dashboard data - TODO", "stats": gin.H{
		"servers":    0,
		"containers": 0,
		"agents":     0,
		"models":     0,
	}})
}

func handleListDockerHosts(c *gin.Context) {
	c.JSON(200, gin.H{"items": []interface{}{}, "total": 0})
}

func handleListContainers(c *gin.Context) {
	c.JSON(200, gin.H{"items": []interface{}{}, "total": 0})
}

func handleGetContainer(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get container - TODO"})
}

func handleContainerStart(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Container start - TODO"})
}

func handleContainerStop(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Container stop - TODO"})
}

func handleContainerDelete(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Container delete - TODO"})
}

func handleListImages(c *gin.Context) {
	c.JSON(200, gin.H{"items": []interface{}{}, "total": 0})
}

func handleListVolumes(c *gin.Context) {
	c.JSON(200, gin.H{"items": []interface{}{}, "total": 0})
}

func handleListProviders(c *gin.Context) {
	c.JSON(200, gin.H{"items": []interface{}{}, "total": 0})
}

func handleCreateProvider(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Create provider - TODO"})
}

func handleListModels(c *gin.Context) {
	c.JSON(200, gin.H{"items": []interface{}{}, "total": 0})
}

func handleCreateModel(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Create model - TODO"})
}

// legacy placeholder handlers retained so existing task/monitor routes still resolve.
func handleListTasks(c *gin.Context) {
	c.JSON(200, gin.H{"items": []interface{}{}, "total": 0})
}

func handleCreateTask(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Create task - TODO"})
}

func handleGetTask(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get task - TODO"})
}



func handleListAuditLogs(c *gin.Context) {
	c.JSON(200, gin.H{"items": []interface{}{}, "total": 0})
}

func handleListApprovals(c *gin.Context) {
	c.JSON(200, gin.H{"items": []interface{}{}, "total": 0})
}

// callLLM performs a synchronous text-only LLM call (non-streaming) used by
// the agent planner. It tries an enabled openai-compatible provider and
// falls back to a deterministic placeholder so the agent works on dev boxes
// without a real key.
func callLLM(aiSvc *service.AIService, modelID, prompt string) (string, error) {
	providers, err := aiSvc.ListProviders()
	if err != nil || len(providers) == 0 {
		s, _ := defaultPlanText(prompt)
		return s, nil
	}
	for _, p := range providers {
		if !p.IsEnabled {
			continue
		}
		if ai.ProviderType(p.APIType) != ai.ProviderOpenAICompatible {
			continue
		}
		client, err := buildClientForProvider(&p)
		if err != nil {
			continue
		}
		out, err := syncLLMCall(client, modelID, prompt)
		if err == nil && out != "" {
			return out, nil
		}
	}
	s, _ := defaultPlanText(prompt)
	return s, nil
}

func buildClientForProvider(p *models.AIProvider) (ai.Client, error) {
	key, err := crypto.Decrypt(p.APIKeyEnc)
	if err != nil {
		return nil, err
	}
	cfg := ai.Config{
		BaseURL: strings.TrimRight(p.BaseURL, "/"),
		APIKey:  key,
	}
	return ai.NewFactory().Build(ai.ProviderType(p.APIType), cfg), nil
}

func syncLLMCall(c ai.Client, modelID, prompt string) (string, error) {
	var buf strings.Builder
	streamer := ai.NewStreamEvents(&buf)
	if err := c.ChatCompletion(context.Background(), ai.ChatRequest{
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

// defaultPlanText is a deterministic fallback that produces a valid planner
// JSON response. Used when no provider is reachable on dev machines.
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

func handleApproveRequest(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Approve request - TODO"})
}

func handleRejectRequest(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Reject request - TODO"})
}

func handleGetSettings(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get settings - TODO"})
}

func handleUpdateSettings(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Update settings - TODO"})
}

var _ = logger.Get
