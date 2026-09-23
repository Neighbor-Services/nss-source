package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

func JSON(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}

func Success(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Created(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusCreated, message, data)
}

func Error(c *gin.Context, statusCode int, message string, errs ...interface{}) {
	var errDetails interface{}
	if len(errs) > 0 {
		errDetails = errs[0]
	}
	c.JSON(statusCode, APIResponse{
		Success: false,
		Message: message,
		Errors:  errDetails,
	})
}

func BadRequest(c *gin.Context, message string, errs ...interface{}) {
	Error(c, http.StatusBadRequest, message, errs...)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}

func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}
