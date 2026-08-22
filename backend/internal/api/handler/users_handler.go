package handler

import (
	"strconv"

	"github.com/aicenter/aicenter/internal/api/middleware"
	"github.com/aicenter/aicenter/internal/api/response"
	"github.com/aicenter/aicenter/internal/service"
	"github.com/gin-gonic/gin"
)

type UsersHandler struct {
	svc *service.UserService
}

func NewUsersHandler(svc *service.UserService) *UsersHandler {
	return &UsersHandler{svc: svc}
}

func (h *UsersHandler) RegisterRoutes(r *gin.RouterGroup) {
	users := r.Group("/users", middleware.RequirePermission("users.manage"))
	{
		users.GET("", h.List)
		users.GET("/:id", h.Get)
		users.PUT("/:id", h.Update)
		users.PATCH("/:id/role", h.ChangeRole)
		users.DELETE("/:id", h.Delete)
	}
}

func (h *UsersHandler) List(c *gin.Context) {
	q := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	res, err := h.svc.List(q, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{
		"items":     res.Users,
		"total":     res.Total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *UsersHandler) Get(c *gin.Context) {
	id := c.Param("id")
	u, err := h.svc.GetByID(id)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}
	response.Success(c, u)
}

func (h *UsersHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req service.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.Username != nil && len(*req.Username) < 3 {
		response.BadRequest(c, "username must be at least 3 characters")
		return
	}
	u, err := h.svc.Update(id, req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, u)
}

func (h *UsersHandler) ChangeRole(c *gin.Context) {
	id := c.Param("id")
	var req service.ChangeRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.Role == "" {
		response.BadRequest(c, "role is required")
		return
	}
	u, err := h.svc.ChangeRole(id, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, u)
}

func (h *UsersHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(id); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "user deleted", "id": id})
}