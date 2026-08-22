package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response is the standard API response envelope
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginatedResponse is the standard paginated response
type PaginatedResponse struct {
	Items interface{} `json:"items"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
}

// ErrorDetail represents a validation error
type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Success returns a successful response
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Created returns a created response
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Code:    0,
		Message: "created",
		Data:    data,
	})
}

// NoContent returns a no content response
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error returns an error response
func Error(c *gin.Context, httpCode int, message string) {
	c.JSON(httpCode, Response{
		Code:    httpCode,
		Message: message,
	})
}

// BadRequest returns a 400 error
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    400,
		Message: message,
	})
}

// Conflict returns a 409 conflict error
func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, Response{
		Code:    409,
		Message: message,
	})
}

// Unauthorized returns a 401 error
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    401,
		Message: message,
	})
}

// Forbidden returns a 403 error
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Code:    403,
		Message: message,
	})
}

// NotFound returns a 404 error
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Code:    404,
		Message: message,
	})
}

// InternalError returns a 500 error
func InternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    500,
		Message: message,
	})
}

// ValidationError returns a 422 error
func ValidationError(c *gin.Context, errors []ErrorDetail) {
	c.JSON(http.StatusUnprocessableEntity, Response{
		Code:    422,
		Message: "Validation failed",
		Data:    errors,
	})
}

// Paginated returns a paginated response
func Paginated(c *gin.Context, items interface{}, total int64, page, limit int) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data: PaginatedResponse{
			Items: items,
			Total: total,
			Page:  page,
			Limit: limit,
		},
	})
}
