package handler

import (
	"fmt"
	"net/http"

	"audio-mixer/internal/config"
	"audio-mixer/internal/service"

	"github.com/gin-gonic/gin"
)

// StreamRadioHandler handles the /api/radio endpoint.
func StreamRadioHandler(c *gin.Context) {
	// Load configuration.
	cfg := config.LoadConfig()

	// Prepare a slice of song file paths.
	// You can expand this to dynamically update the queue.
	songQueue := []string{cfg.MP3FilePath1, cfg.MP3FilePath2}

	// Start the continuous radio stream.
	if err := service.StreamRadio(c, songQueue); err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error streaming radio: %v", err))
	}
}
