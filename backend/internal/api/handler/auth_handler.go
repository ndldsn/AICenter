package handler

import (
	"github.com/aicenter/aicenter/internal/api/response"
	"github.com/aicenter/aicenter/internal/models"
	"github.com/aicenter/aicenter/internal/repository"
	"github.com/aicenter/aicenter/internal/service"
	"github.com/gin-gonic/gin"
)

// AuthHandler handles public auth endpoints: register / login / refresh / me.
type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// RegisterRoutes exposes the auth endpoints. login/register/refresh are public;
// me requires the JWTAuth middleware to have injected userID.
func (h *AuthHandler) RegisterPublicRoutes(public *gin.RouterGroup) {
	public.POST("/auth/register", h.Register)
	public.POST("/auth/login", h.Login)
	public.POST("/auth/refresh", h.Refresh)
}

func (h *AuthHandler) RegisterProtectedRoutes(protected *gin.RouterGroup) {
	protected.GET("/auth/me", h.Me)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	u, err := h.svc.Register(&req)
	if err != nil {
		if err == repository.ErrUserAlreadyExists {
			response.Conflict(c, "username or email already exists")
		} else {
			response.InternalError(c, "failed to register")
		}
		return
	}
	response.Created(c, u)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := h.svc.Login(&req)
	if err != nil {
		response.Unauthorized(c, "invalid credentials")
		return
	}
	c.JSON(200, response.Response{
		Code:    0,
		Message: "login successful",
		Data: gin.H{
			"access_token":  result.AccessToken,
			"refresh_token": result.RefreshToken,
			"user":          result.User,
		},
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := h.svc.Refresh(body.RefreshToken)
	if err != nil {
		response.Unauthorized(c, "invalid refresh token")
		return
	}
	c.JSON(200, response.Response{
		Code:    0,
		Message: "token refreshed",
		Data: gin.H{
			"access_token":  result.AccessToken,
			"refresh_token": result.RefreshToken,
			"user":          result.User,
		},
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "missing user context")
		return
	}
	u, err := h.svc.Me(userID.(string))
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}
	response.Success(c, u)
}