package main

import (
	"audio-mixer/internal/config"
	"audio-mixer/internal/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()
	router := gin.Default()

	api := router.Group("/api")
	{
		api.GET("/radio", handler.StreamRadioHandler)
		api.POST("/radio/skip", handler.SkipRadioHandler)
		api.POST("/radio/queue", handler.AddSongHandler)
		api.GET("/radio/queue", handler.GetQueueHandler)
	}

	router.Run(":" + cfg.Port)
}
