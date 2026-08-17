package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Recovery converts panics into logged HTTP failures.
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		logger.Error("panic recovered", zap.Any("panic", recovered), zap.String("request_id", RequestID(c)))
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

// SecurityHeaders adds basic defensive response headers.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Next()
	}
}

// AccessLog records one structured event per request.
func AccessLog(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Info("http request", zap.String("request_id", RequestID(c)), zap.String("method", c.Request.Method), zap.String("route", c.FullPath()), zap.Int("status", c.Writer.Status()), zap.Duration("duration", time.Since(start)))
	}
}

var Module = fx.Module("middleware")
