package router

import (
	"database/sql"
	"net/http"

	"mobicode/apps/server/internal/config"
	"mobicode/apps/server/internal/middleware"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// New builds the Gin router and registers application routes.
func New(cfg config.Config, logger *zap.Logger, db *sql.DB, graphqlHandler http.Handler) *gin.Engine {
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
	r.POST("/graphql", gin.WrapH(graphqlHandler))
	if cfg.Environment != "production" {
		r.GET("/graphql", gin.WrapH(playground.Handler("GraphQL Playground", "/graphql")))
	}
	return r
}

var Module = fx.Module("router", fx.Provide(New))
