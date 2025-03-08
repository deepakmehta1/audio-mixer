package main

import (
	"audio-mixer/internal/config"
	"audio-mixer/internal/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration.
	cfg := config.LoadConfig()

	// Create a Gin router.
	router := gin.Default()

	// Set up API routes.
	api := router.Group("/api")
	{
		api.GET("/radio", handler.StreamRadioHandler)
	}

	// Run the server.
	router.Run(":" + cfg.Port)
}
