package handler

import (
	"context"

	"github.com/aicenter/aicenter/internal/api/response"
	"github.com/aicenter/aicenter/internal/service"
	"github.com/gin-gonic/gin"
)

// BatchHandler handles multi-server batch operations.
type BatchHandler struct {
	svc *service.BatchService
}

func NewBatchHandler(svc *service.BatchService) *BatchHandler {
	return &BatchHandler{svc: svc}
}

// RegisterRoutes registers batch routes under /api/v1/servers/batch/...
func (h *BatchHandler) RegisterRoutes(group *gin.RouterGroup, auth gin.HandlerFunc) {
	servers := group.Group("/servers")
	batch := servers.Group("/batch")
	batch.Use(auth)
	{
		batch.POST("/command", h.BatchCommand)
	}
}

// BatchCommand POST /api/v1/servers/batch/command
// Runs a single command across many servers concurrently.
func (h *BatchHandler) BatchCommand(c *gin.Context) {
	var req service.BatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid body: "+err.Error())
		return
	}
	if req.Command == "" {
		response.BadRequest(c, "command is required")
		return
	}
	if !c.IsAborted() {
		req.ServerIDs = dedupe(req.ServerIDs)
	}
	results := h.svc.BatchCommand(context.Background(), &req)
	response.Success(c, gin.H{"items": results, "total": len(results)})
}

func dedupe(ids []string) []string {
	seen := map[string]bool{}
	out := ids[:0]
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}
