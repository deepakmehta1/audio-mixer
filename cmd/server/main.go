package main

import (
	"audio-mixer/internal/config"
	"audio-mixer/internal/handler"
	"audio-mixer/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration.
	cfg := config.LoadConfig()

	// (Optional) Initialize the default queue with a song.
	// This adds song1.mp3 to the global queue.
	service.AddSong(cfg.MP3FilePath1)
	// You can add more songs as needed:
	// service.AddSong(cfg.MP3FilePath2)

	// Start the continuous streaming (global broadcaster).
	service.StartStreaming()

	// Set up the Gin server.
	router := gin.Default()
	api := router.Group("/api")
	{
		api.GET("/radio", handler.StreamRadioHandler)
		api.POST("/radio/skip", handler.SkipRadioHandler)
		api.POST("/radio/queue", handler.AddSongHandler)
		api.GET("/radio/queue", handler.GetQueueHandler)
	}
	router.Run("0.0.0.0:" + cfg.Port)
}
