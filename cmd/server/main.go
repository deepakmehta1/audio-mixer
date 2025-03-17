package main

import (
	"log"
	"time"

	"audio-mixer/internal/config"
	"audio-mixer/internal/handler"
	"audio-mixer/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	// Load the regular queue from the local static file.
	if err := service.LoadRegularQueue("files/songs.json"); err != nil {
		log.Printf("Warning: could not load regular queue from local file: %v", err)
	} else {
		log.Printf("Regular queue loaded from files/songs.json")
	}

	// Schedule a refresh from S3; new songs will be appended to the regular queue.
	bucketName := "tingo-regular-queue"
	prefix := "songs/"
	service.ScheduleS3QueueRefresh(bucketName, prefix, 5*time.Hour)

	// Start continuous HLS streaming.
	service.StartStreaming()
	// Start the YouTube conversion worker.
	service.StartYTWorker()

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	router.Static("/hls", "./hls")

	api := router.Group("/api")
	{
		api.GET("/radio", handler.StreamRadioHandler)
		api.POST("/radio/skip", handler.SkipRadioHandler)
		api.GET("/radio/queue", handler.GetPriorityQueueHandler)
		api.POST("/radio/queue", handler.AddPrioritySongHandler)
		api.POST("/radio/youtube", handler.AddYouTubeSongHandler)
	}

	router.Run(":" + cfg.Port)
}
