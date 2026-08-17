package router

import (
	"database/sql"
	"mobicode/apps/server/internal/config"
	"mobicode/apps/server/internal/controller"
	"mobicode/apps/server/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// New builds the Gin router and registers application routes.
func New(cfg config.Config, logger *zap.Logger, db *sql.DB, tasks *controller.TaskController) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(middleware.RequestIDMiddleware(), middleware.Recovery(logger), middleware.SecurityHeaders(), middleware.AccessLog(logger), func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, cfg.Server.MaxBodyBytes)
		c.Next()
	})
	r.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"status": "ok"}, "error": nil, "request_id": middleware.RequestID(c)})
	})
	r.GET("/health/ready", func(c *gin.Context) {
		if err := db.PingContext(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"data": nil, "error": gin.H{"code": "dependency_unavailable", "message": "database unavailable", "fields": []string{}}, "request_id": middleware.RequestID(c)})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"status": "ok"}, "error": nil, "request_id": middleware.RequestID(c)})
	})
	api := r.Group("/api/v1/tasks")
	api.POST("", tasks.Create)
	api.GET("", tasks.List)
	api.GET("/:id", tasks.Get)
	api.PATCH("/:id", tasks.Update)
	api.DELETE("/:id", tasks.Delete)
	return r
}

var Module = fx.Module("router", fx.Provide(New))
