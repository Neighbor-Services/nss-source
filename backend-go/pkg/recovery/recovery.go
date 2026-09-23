package recovery

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"backend-go/pkg/logger"

	"github.com/gin-gonic/gin"
)

// SafeRun executes a task safely, catching any panics and logging them via structured slog.
// It avoids crashing the entire process and serves as an in-house resilient alternative to external APMs.
func SafeRun(taskName string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			slog.Error("CRITICAL: Panic caught in background task",
				slog.String("task", taskName),
				slog.Any("panic", r),
				slog.String("stack", logger.MaskPII(stack)),
			)
		}
	}()

	fn()
}

// SafeRunWithContext executes a task with context, catching panics gracefully.
func SafeRunWithContext(ctx context.Context, taskName string, fn func(ctx context.Context)) {
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			slog.ErrorContext(ctx, "CRITICAL: Panic caught in background task with context",
				slog.String("task", taskName),
				slog.Any("panic", r),
				slog.String("stack", logger.MaskPII(stack)),
			)
		}
	}()

	fn(ctx)
}

// Middleware returns a Gin panic recovery middleware with structured JSON logging and PII masking.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				stack := string(debug.Stack())
				path := c.Request.URL.Path
				method := c.Request.Method
				clientIP := c.ClientIP()

				slog.Error("HTTP Server Panic Recovered",
					slog.String("method", method),
					slog.String("path", path),
					slog.String("client_ip", clientIP),
					slog.Any("panic", r),
					slog.String("stack", logger.MaskPII(stack)),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "An internal server error occurred. Our engineering team has been alerted.",
					"error":   fmt.Sprintf("%v", r),
				})
			}
		}()
		c.Next()
	}
}
