package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"mobicode/apps/server/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	if err := router.Run(":" + cfg.Server.Port); err != nil {
		panic(err)
	}
}
