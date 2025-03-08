package handler

import (
	"fmt"
	"net/http"

	"audio-mixer/internal/service"

	"github.com/gin-gonic/gin"
)

// StreamRadioHandler handles the GET /api/radio endpoint.
// It streams a continuous radio-style MP3 stream.
func StreamRadioHandler(c *gin.Context) {
	queue := service.GetQueue()
	if len(queue) == 0 {
		c.String(http.StatusNotFound, "No songs in queue.")
		return
	}
	if err := service.StreamRadio(c, queue); err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error streaming radio: %v", err))
	}
}

// SkipRadioHandler handles the POST /api/radio/skip endpoint.
// It sends a skip signal to immediately change the current song.
func SkipRadioHandler(c *gin.Context) {
	service.SkipCurrentSong()
	c.String(http.StatusOK, "Skipped current song.")
}

// AddSongHandler handles the POST /api/radio/queue endpoint.
// It expects a JSON body with a "path" field indicating the MP3 file to add.
func AddSongHandler(c *gin.Context) {
	var req struct {
		Path string `json:"path"`
	}
	if err := c.BindJSON(&req); err != nil || req.Path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request. Provide a valid song path."})
		return
	}
	service.AddSong(req.Path)
	c.JSON(http.StatusOK, gin.H{"message": "Song added", "path": req.Path})
}

// GetQueueHandler handles the GET /api/radio/queue endpoint.
// It returns the current song queue.
func GetQueueHandler(c *gin.Context) {
	queue := service.GetQueue()
	c.JSON(http.StatusOK, gin.H{"queue": queue})
}
