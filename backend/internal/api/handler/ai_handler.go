package handler

import (
	"github.com/aicenter/aicenter/internal/ai"
	"github.com/aicenter/aicenter/internal/models"
	"github.com/aicenter/aicenter/internal/service"
	"github.com/gin-gonic/gin"
)

type AIHandler struct {
	svc *service.AIService
}

func NewAIHandler(svc *service.AIService) *AIHandler {
	return &AIHandler{svc: svc}
}

// Provider CRUD

func (h *AIHandler) ListProviders(c *gin.Context) {
	list, err := h.svc.ListProviders()
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 0, "message": "success", "data": list})
}

func (h *AIHandler) GetProvider(c *gin.Context) {
	p, err := h.svc.GetProvider(c.Param("id"))
	if err != nil {
		c.JSON(404, gin.H{"code": 404, "message": "provider not found"})
		return
	}
	c.JSON(200, gin.H{"code": 0, "message": "success", "data": p})
}

func (h *AIHandler) CreateProvider(c *gin.Context) {
	var req RequestCreateProvider
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "invalid request body"})
		return
	}
	if req.Name == "" {
		c.JSON(400, gin.H{"code": 400, "message": "name is required"})
		return
	}
	p, err := h.svc.CreateProvider(req.AIProvider)
	if err != nil {
		c.JSON(400, gin.H{"code": 400, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 0, "message": "created", "data": p})
}

func (h *AIHandler) UpdateProvider(c *gin.Context) {
	var req RequestUpdateProvider
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "invalid request body"})
		return
	}
	p, err := h.svc.UpdateProvider(c.Param("id"), req.AIProvider)
	if err != nil {
		if err == service.ErrDecryptionFailed {
			c.JSON(500, gin.H{"code": 500, "message": "provider key decryption failed (did the encryption key change?)"})
			return
		}
		if err.Error() == "provider not found" || err.Error() == "record not found" {
			c.JSON(404, gin.H{"code": 404, "message": "provider not found"})
			return
		}
		c.JSON(400, gin.H{"code": 400, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 0, "message": "updated", "data": p})
}

func (h *AIHandler) DeleteProvider(c *gin.Context) {
	if err := h.svc.DeleteProvider(c.Param("id")); err != nil {
		c.JSON(404, gin.H{"code": 404, "message": "provider not found"})
		return
	}
	c.JSON(200, gin.H{"code": 0, "message": "deleted"})
}

// Models

func (h *AIHandler) ListModels(c *gin.Context) {
	providerID := c.Param("provider_id")
	list, err := h.svc.ListModels(providerID)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 0, "message": "success", "data": list})
}

// Chat — streaming SSE

func (h *AIHandler) Chat(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(200)

	var req RequestChat
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c.Writer, "invalid request body")
		return
	}
	if req.ProviderID == "" || req.ModelID == "" || len(req.Messages) == 0 {
		writeError(c.Writer, "provider_id, model_id, and messages are required")
		return
	}

	streamer := ai.NewStreamEvents(c.Writer)
	if err := h.svc.Chat(req.ProviderID, req.ModelID, req.Messages, streamer); err != nil {
		writeError(c.Writer, err.Error())
		return
	}
	writeDone(c.Writer)
}

type RequestCreateProvider struct {
	models.AIProvider
}

type RequestUpdateProvider struct {
	models.AIProvider
}

type RequestChat struct {
	ProviderID string       `json:"provider_id"`
	ModelID    string       `json:"model_id"`
	Messages   []ai.Message `json:"messages"`
}

func writeError(w gin.ResponseWriter, msg string) {
	_, _ = w.Write([]byte("event: error\ndata: " + msg + "\n\n"))
}

func writeDone(w gin.ResponseWriter) {
	_, _ = w.Write([]byte("event: done\ndata: [DONE]\n\n"))
}
