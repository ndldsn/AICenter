package handler

import (
	"strconv"

	"github.com/aicenter/aicenter/internal/service"
	"github.com/gin-gonic/gin"
)

// ServerHandler handles server API requests
type ServerHandler struct {
	svc *service.ServerService
}

// NewServerHandler creates a new server handler
func NewServerHandler() *ServerHandler {
	return &ServerHandler{svc: service.NewServerService()}
}

// RegisterRoutes registers server routes
func (h *ServerHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	servers := r.Group("/servers")
	servers.Use(authMiddleware)
	{
		servers.GET("", h.List)
		servers.POST("", h.Create)
		servers.GET("/:id", h.Get)
		servers.PUT("/:id", h.Update)
		servers.DELETE("/:id", h.Delete)
		servers.POST("/:id/connect", h.TestConnection)
		servers.GET("/:id/metrics", h.GetMetrics)
		servers.POST("/:id/agent-token", h.GenerateAgentToken)
	}

	groups := r.Group("/server-groups")
	groups.Use(authMiddleware)
	{
		groups.GET("", h.ListGroups)
		groups.POST("", h.CreateGroup)
		groups.DELETE("/:id", h.DeleteGroup)
	}
}

// List handles GET /api/v1/servers
func (h *ServerHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	servers, total, err := h.svc.ListServers(page, limit)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"code": 0,
		"message": "success",
		"data": gin.H{
			"items": servers,
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

// Create handles POST /api/v1/servers
func (h *ServerHandler) Create(c *gin.Context) {
	var req service.CreateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": err.Error()})
		return
	}

	server, err := h.svc.CreateServer(&req)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(201, gin.H{"code": 0, "message": "created", "data": server})
}

// Get handles GET /api/v1/servers/:id
func (h *ServerHandler) Get(c *gin.Context) {
	id := c.Param("id")
	server, err := h.svc.GetServer(id)
	if err != nil {
		c.JSON(404, gin.H{"code": 404, "message": "server not found"})
		return
	}

	c.JSON(200, gin.H{"code": 0, "message": "success", "data": server})
}

// Update handles PUT /api/v1/servers/:id
func (h *ServerHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req service.UpdateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": err.Error()})
		return
	}

	if err := h.svc.UpdateServer(id, &req); err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"code": 0, "message": "updated"})
}

// Delete handles DELETE /api/v1/servers/:id
func (h *ServerHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteServer(id); err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"code": 0, "message": "deleted"})
}

// TestConnection handles POST /api/v1/servers/:id/connect
func (h *ServerHandler) TestConnection(c *gin.Context) {
	id := c.Param("id")

	// Check if this is a test for new server data (id == "test")
	if id == "test" {
		var req service.CreateServerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"code": 400, "message": err.Error()})
			return
		}
		result, err := h.svc.TestNewConnection(&req)
		if err != nil {
			c.JSON(500, gin.H{"code": 500, "message": err.Error()})
			return
		}
		c.JSON(200, gin.H{"code": 0, "message": "success", "data": result})
		return
	}

	result, err := h.svc.TestConnection(id)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"code": 0, "message": "success", "data": result})
}

// GetMetrics handles GET /api/v1/servers/:id/metrics
func (h *ServerHandler) GetMetrics(c *gin.Context) {
	id := c.Param("id")
	// TODO: Get real metrics from server
	c.JSON(200, gin.H{
		"code": 0,
		"message": "success",
		"data": gin.H{
			"server_id": id,
			"metrics": []interface{}{},
		},
	})
}

// GenerateAgentToken handles POST /api/v1/servers/:id/agent-token
func (h *ServerHandler) GenerateAgentToken(c *gin.Context) {
	id := c.Param("id")
	token, err := h.svc.GenerateAgentToken(id)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"code": 0, "message": "success", "data": gin.H{"token": token}})
}

// ListGroups handles GET /api/v1/server-groups
func (h *ServerHandler) ListGroups(c *gin.Context) {
	groups, err := h.svc.GetServerGroups()
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"code": 0, "message": "success", "data": groups})
}

// CreateGroup handles POST /api/v1/server-groups
func (h *ServerHandler) CreateGroup(c *gin.Context) {
	var req struct {
		Name        string  `json:"name" binding:"required"`
		Description string  `json:"description"`
		ParentID    *string `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": err.Error()})
		return
	}

	group, err := h.svc.CreateServerGroup(req.Name, req.Description, req.ParentID)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(201, gin.H{"code": 0, "message": "created", "data": group})
}

// DeleteGroup handles DELETE /api/v1/server-groups/:id
func (h *ServerHandler) DeleteGroup(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteServerGroup(id); err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"code": 0, "message": "deleted"})
}
