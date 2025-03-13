package main

import (
	"audio-mixer/internal/config"
	"audio-mixer/internal/handler"
	"audio-mixer/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	// Optionally, initialize the default queue with a local song.
	service.AddSong(cfg.MP3FilePath)
	// service.AddSong(cfg.MP3FilePath2)

	// Start continuous streaming (broadcasting) in background.
	service.StartStreaming()
	// Start the YouTube conversion worker with the loaded configuration.
	service.StartYTWorker()

	router := gin.Default()
	api := router.Group("/api")
	{
		api.GET("/radio", handler.StreamRadioHandler)
		api.POST("/radio/skip", handler.SkipRadioHandler)
		api.POST("/radio/queue", handler.AddSongHandler)
		api.GET("/radio/queue", handler.GetQueueHandler)
		api.POST("/radio/youtube", handler.AddYouTubeSongHandler)
	}
	router.Run(":" + cfg.Port)
}
