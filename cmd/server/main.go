package main

import (
	"log"

	"audio-mixer/internal/config"
	"audio-mixer/internal/handler"
	"audio-mixer/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	// Load the regular queue from the static file.
	if err := service.LoadRegularQueue("files/songs.json"); err != nil {
		log.Printf("Warning: could not load regular queue: %v", err)
	} else {
		log.Printf("Regular queue loaded from files/songs.json")
	}

	// Start continuous HLS streaming.
	service.StartStreaming()
	// Start the YouTube conversion worker.
	service.StartYTWorker()

	// Create the Gin router.
	router := gin.Default()

	// 1) Enable CORS with desired settings.
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 2) Serve HLS segments from the hls folder.
	router.Static("/hls", "./hls")

	// 3) Define your API routes.
	api := router.Group("/api")
	{
		// Clients tune in by requesting the HLS manifest.
		api.GET("/radio", handler.StreamRadioHandler)
		api.POST("/radio/skip", handler.SkipRadioHandler)
		api.GET("/radio/queue", handler.GetPriorityQueueHandler)
		// Priority song uploads.
		api.POST("/radio/queue", handler.AddPrioritySongHandler)
		// YouTube song enqueuing.
		api.POST("/radio/youtube", handler.AddYouTubeSongHandler)
	}

	// 4) Run the server on the configured port.
	router.Run(":" + cfg.Port)
}
