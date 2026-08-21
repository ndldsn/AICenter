package handler

import (
	"strconv"

	"github.com/aicenter/aicenter/internal/service"
	"github.com/gin-gonic/gin"
)

// DockerHandler handles Docker API requests (Phase 3).
type DockerHandler struct {
	svc *service.DockerService
}

// NewDockerHandler creates a new docker handler.
func NewDockerHandler() *DockerHandler {
	return &DockerHandler{svc: service.NewDockerService()}
}

// RegisterRoutes registers docker routes under the given group.
func (h *DockerHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	docker := r.Group("/docker")
	docker.Use(authMiddleware)
	{
		docker.GET("/hosts", h.ListHosts)
		docker.GET("/containers", h.ListContainers)
		docker.GET("/containers/:id", h.GetContainer)
		docker.POST("/containers/:id/start", h.Start)
		docker.POST("/containers/:id/stop", h.Stop)
		docker.DELETE("/containers/:id", h.Delete)
	}
}

// ListHosts handles GET /api/v1/docker/hosts
func (h *DockerHandler) ListHosts(c *gin.Context) {
	hosts, err := h.svc.ListHosts(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 0, "message": "success", "data": gin.H{"items": hosts, "total": len(hosts)}})
}

// ListContainers handles GET /api/v1/docker/containers
func (h *DockerHandler) ListContainers(c *gin.Context) {
	all := c.DefaultQuery("all", "true") == "true"
	list, err := h.svc.ListContainers(c.Request.Context(), all)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 0, "message": "success", "data": gin.H{"items": list, "total": len(list)}})
}

// GetContainer handles GET /api/v1/docker/containers/:id
func (h *DockerHandler) GetContainer(c *gin.Context) {
	d, err := h.svc.GetContainer(c.Request.Context(), c.Param("id"))
	if err != nil {
		if err.Error() == "container not found" {
			c.JSON(404, gin.H{"code": 404, "message": "container not found"})
			return
		}
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 0, "message": "success", "data": d})
}

// Start handles POST /api/v1/docker/containers/:id/start
func (h *DockerHandler) Start(c *gin.Context) {
	if err := h.svc.StartContainer(c.Request.Context(), c.Param("id")); err != nil {
		if err.Error() == "container not found" {
			c.JSON(404, gin.H{"code": 404, "message": "container not found"})
			return
		}
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 0, "message": "started"})
}

// Stop handles POST /api/v1/docker/containers/:id/stop
func (h *DockerHandler) Stop(c *gin.Context) {
	timeout := 10
	if v := c.Query("timeout"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			timeout = n
		}
	}
	if err := h.svc.StopContainer(c.Request.Context(), c.Param("id"), timeout); err != nil {
		if err.Error() == "container not found" {
			c.JSON(404, gin.H{"code": 404, "message": "container not found"})
			return
		}
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 0, "message": "stopped"})
}

// Delete handles DELETE /api/v1/docker/containers/:id
func (h *DockerHandler) Delete(c *gin.Context) {
	force := c.Query("force") == "true"
	if err := h.svc.DeleteContainer(c.Request.Context(), c.Param("id"), force); err != nil {
		if err.Error() == "container not found" {
			c.JSON(404, gin.H{"code": 404, "message": "container not found"})
			return
		}
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 0, "message": "deleted"})
}
