package router

import (
	"database/sql"

	"github.com/aicenter/aicenter/internal/api/handler"
	"github.com/aicenter/aicenter/internal/api/middleware"
	"github.com/aicenter/aicenter/internal/config"
	"github.com/aicenter/aicenter/internal/pkg/logger"
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

	// Health check (no auth)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "aicenter"})
	})

	// WebSocket endpoint (no auth for now - TODO: auth via query param)
	r.GET("/ws", func(c *gin.Context) {
		websocket.ServeWs(hub, c.Writer, c.Request)
	})

	// API v1
	v1 := r.Group("/api/v1")

	// Public routes
	auth := v1.Group("/auth")
	{
		auth.POST("/login", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "login endpoint - TODO"})
		})
		auth.POST("/register", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "register endpoint - TODO"})
		})
		auth.POST("/refresh", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "refresh endpoint - TODO"})
		})
	}

	// Protected routes (with mock auth for development)
	protected := v1.Group("/")
	protected.Use(middleware.MockAuth())
	{
		// Dashboard
		protected.GET("/dashboard", handleDashboard)

		// Servers - using real handler
		serverHandler := handler.NewServerHandler()
		serverHandler.RegisterRoutes(protected, middleware.MockAuth())

		// Docker (Phase 3 - real handler)
		dockerHandler := handler.NewDockerHandler()
		dockerHandler.RegisterRoutes(protected, middleware.MockAuth())

		// AI Providers & Models
		protected.GET("/ai/providers", handleListProviders)
		protected.POST("/ai/providers", handleCreateProvider)
		protected.GET("/ai/models", handleListModels)
		protected.POST("/ai/models", handleCreateModel)

		// Agents
		protected.GET("/agents", handleListAgents)
		protected.POST("/agents", handleCreateAgent)
		protected.GET("/agents/:id", handleGetAgent)
		protected.PUT("/agents/:id", handleUpdateAgent)
		protected.DELETE("/agents/:id", handleDeleteAgent)
		protected.POST("/agents/:id/sessions", handleCreateSession)
		protected.GET("/agents/sessions", handleListSessions)
		protected.POST("/agents/sessions/:id/messages", handleSendMessage)

		// Tasks
		protected.GET("/tasks", handleListTasks)
		protected.POST("/tasks", handleCreateTask)
		protected.GET("/tasks/:id", handleGetTask)

		// Monitor
		protected.GET("/monitor/metrics", handleMonitorMetrics)
		protected.GET("/monitor/alerts", handleMonitorAlerts)

		// Audit
		protected.GET("/audit-logs", handleListAuditLogs)

		// Approvals
		protected.GET("/approvals", handleListApprovals)
		protected.POST("/approvals/:id/approve", handleApproveRequest)
		protected.POST("/approvals/:id/reject", handleRejectRequest)

		// Users
		protected.GET("/users", handleListUsers)
		protected.GET("/roles", handleListRoles)

		// Settings
		protected.GET("/settings", handleGetSettings)
		protected.PUT("/settings", handleUpdateSettings)
	}

	return r
}

// Placeholder handlers - these will be replaced with real implementations

func handleDashboard(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Dashboard data - TODO", "stats": gin.H{
		"servers":   0,
		"containers": 0,
		"agents":    0,
		"models":    0,
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

func handleListAgents(c *gin.Context) {
	c.JSON(200, gin.H{"items": []interface{}{}, "total": 0})
}

func handleCreateAgent(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Create agent - TODO"})
}

func handleGetAgent(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get agent - TODO"})
}

func handleUpdateAgent(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Update agent - TODO"})
}

func handleDeleteAgent(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Delete agent - TODO"})
}

func handleCreateSession(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Create session - TODO"})
}

func handleListSessions(c *gin.Context) {
	c.JSON(200, gin.H{"items": []interface{}{}, "total": 0})
}

func handleSendMessage(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Send message - TODO"})
}

func handleListTasks(c *gin.Context) {
	c.JSON(200, gin.H{"items": []interface{}{}, "total": 0})
}

func handleCreateTask(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Create task - TODO"})
}

func handleGetTask(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get task - TODO"})
}

func handleMonitorMetrics(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Monitor metrics - TODO"})
}

func handleMonitorAlerts(c *gin.Context) {
	c.JSON(200, gin.H{"items": []interface{}{}, "total": 0})
}

func handleListAuditLogs(c *gin.Context) {
	c.JSON(200, gin.H{"items": []interface{}{}, "total": 0})
}

func handleListApprovals(c *gin.Context) {
	c.JSON(200, gin.H{"items": []interface{}{}, "total": 0})
}

func handleApproveRequest(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Approve request - TODO"})
}

func handleRejectRequest(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Reject request - TODO"})
}

func handleListUsers(c *gin.Context) {
	c.JSON(200, gin.H{"items": []interface{}{}, "total": 0})
}

func handleListRoles(c *gin.Context) {
	c.JSON(200, gin.H{"items": []interface{}{}, "total": 0})
}

func handleGetSettings(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get settings - TODO"})
}

func handleUpdateSettings(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Update settings - TODO"})
}

var _ = logger.Get
