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
		docker.GET("/containers/:id/logs", h.GetContainerLogs)
		docker.POST("/containers/:id/start", h.Start)
		docker.POST("/containers/:id/stop", h.Stop)
		docker.DELETE("/containers/:id", h.Delete)
		docker.GET("/images", h.ListImages)
		docker.POST("/images/pull", h.PullImage)
		docker.DELETE("/images/:id", h.DeleteImage)
		docker.GET("/volumes", h.ListVolumes)
		docker.POST("/volumes", h.CreateVolume)
		docker.DELETE("/volumes/:name", h.DeleteVolume)
		docker.GET("/networks", h.ListNetworks)
		docker.POST("/networks", h.CreateNetwork)
		docker.DELETE("/networks/:id", h.DeleteNetwork)
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

// GetContainerLogs handles GET /api/v1/docker/containers/:id/logs
func (h *DockerHandler) GetContainerLogs(c *gin.Context) {
	tail := 200
	if v := c.Query("tail"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			tail = n
		}
	}
	logs, err := h.svc.ContainerLogs(c.Request.Context(), c.Param("id"), tail)
	if err != nil {
		if err.Error() == "container not found" {
			c.JSON(404, gin.H{"code": 404, "message": "container not found"})
			return
		}
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 0, "message": "success", "data": gin.H{"id": c.Param("id"), "logs": logs}})
}

// ListImages handles GET /api/v1/docker/images
func (h *DockerHandler) ListImages(c *gin.Context) {
	list, err := h.svc.ListImages(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 0, "message": "success", "data": gin.H{"items": list, "total": len(list)}})
}

// PullImage handles POST /api/v1/docker/images/pull
func (h *DockerHandler) PullImage(c *gin.Context) {
	var req struct {
		Repository string `json:"repository"`
		Tag        string `json:"tag"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "invalid request body"})
		return
	}
	if err := h.svc.PullImage(c.Request.Context(), req.Repository, req.Tag); err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 0, "message": "pulled"})
}

// DeleteImage handles DELETE /api/v1/docker/images/:id
func (h *DockerHandler) DeleteImage(c *gin.Context) {
	force := c.Query("force") == "true"
	if err := h.svc.DeleteImage(c.Request.Context(), c.Param("id"), force); err != nil {
		if err.Error() == "container not found" {
			c.JSON(404, gin.H{"code": 404, "message": "image not found"})
			return
		}
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 0, "message": "deleted"})
}

// ListVolumes handles GET /api/v1/docker/volumes
func (h *DockerHandler) ListVolumes(c *gin.Context) {
	list, err := h.svc.ListVolumes(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 0, "message": "success", "data": gin.H{"items": list, "total": len(list)}})
}

// CreateVolume handles POST /api/v1/docker/volumes
func (h *DockerHandler) CreateVolume(c *gin.Context) {
	var req struct {
		Name   string `json:"name"`
		Driver string `json:"driver"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "invalid request body"})
		return
	}
	if req.Name == "" {
		c.JSON(400, gin.H{"code": 400, "message": "name is required"})
		return
	}
	vol, err := h.svc.CreateVolume(c.Request.Context(), req.Name, req.Driver)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 0, "message": "created", "data": vol})
}

// DeleteVolume handles DELETE /api/v1/docker/volumes/:name
func (h *DockerHandler) DeleteVolume(c *gin.Context) {
	force := c.Query("force") == "true"
	if err := h.svc.DeleteVolume(c.Request.Context(), c.Param("name"), force); err != nil {
		if err.Error() == "container not found" {
			c.JSON(404, gin.H{"code": 404, "message": "volume not found"})
			return
		}
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 0, "message": "deleted"})
}

// ListNetworks handles GET /api/v1/docker/networks
func (h *DockerHandler) ListNetworks(c *gin.Context) {
	list, err := h.svc.ListNetworks(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 0, "message": "success", "data": gin.H{"items": list, "total": len(list)}})
}

// CreateNetwork handles POST /api/v1/docker/networks
func (h *DockerHandler) CreateNetwork(c *gin.Context) {
	var req struct {
		Name   string `json:"name"`
		Driver string `json:"driver"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "invalid request body"})
		return
	}
	if req.Name == "" {
		c.JSON(400, gin.H{"code": 400, "message": "name is required"})
		return
	}
	net, err := h.svc.CreateNetwork(c.Request.Context(), req.Name, req.Driver)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 0, "message": "created", "data": net})
}

// DeleteNetwork handles DELETE /api/v1/docker/networks/:id
func (h *DockerHandler) DeleteNetwork(c *gin.Context) {
	if err := h.svc.DeleteNetwork(c.Request.Context(), c.Param("id")); err != nil {
		if err.Error() == "container not found" {
			c.JSON(404, gin.H{"code": 404, "message": "network not found"})
			return
		}
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 0, "message": "deleted"})
}
