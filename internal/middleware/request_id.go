package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

const requestIDKey = "request_id"

type requestIDContextKey struct{}

// RequestIDMiddleware assigns a request identifier to each request.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			bytes := make([]byte, 16)
			if _, err := rand.Read(bytes); err == nil {
				id = hex.EncodeToString(bytes)
			} else {
				id = "unknown"
			}
		}
		c.Set(requestIDKey, id)
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), requestIDContextKey{}, id))
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

// RequestIDFromContext returns the request identifier from a standard context.
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDContextKey{}).(string)
	return id
}

// RequestID returns the request identifier from Gin context.
func RequestID(c *gin.Context) string {
	value, ok := c.Get(requestIDKey)
	if !ok {
		return ""
	}
	id, _ := value.(string)
	return id
}
