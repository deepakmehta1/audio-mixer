package handler

import (
	"fmt"
	"net/http"

	"audio-mixer/internal/service"

	"github.com/gin-gonic/gin"
)

// StreamRadioHandler handles GET /api/radio.
// Clients subscribe to the live broadcast stream.
func StreamRadioHandler(c *gin.Context) {
	if err := service.StreamRadio(c); err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error streaming radio: %v", err))
	}
}

// SkipRadioHandler handles POST /api/radio/skip.
// It sends a signal to immediately skip the current song.
func SkipRadioHandler(c *gin.Context) {
	service.SkipCurrentSong()
	c.String(http.StatusOK, "Skip signal sent.")
}

// AddSongHandler handles POST /api/radio/queue.
// It expects a JSON body with a "path" field.
func AddSongHandler(c *gin.Context) {
	var req struct {
		Path string `json:"path"`
	}
	if err := c.BindJSON(&req); err != nil || req.Path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request. Provide a valid song path."})
		return
	}
	service.AddSong(req.Path)
	c.JSON(http.StatusOK, gin.H{"message": "Song added", "queue": service.GetQueue()})
}

// GetQueueHandler handles GET /api/radio/queue.
// It returns the current song queue.
func GetQueueHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"queue": service.GetQueue()})
}
