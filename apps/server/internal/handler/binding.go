package handler

import (
	"mobicode/apps/server/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

// BindAndValidate decodes and validates a JSON request body.
func BindAndValidate(c *gin.Context, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		c.JSON(http.StatusUnprocessableEntity, envelope{Data: nil, Error: errorBody{Code: "validation", Message: "request validation failed", Fields: []string{err.Error()}}, RequestID: middleware.RequestID(c)})
		return false
	}
	return true
}
