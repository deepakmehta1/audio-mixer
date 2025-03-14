package handler

import (
	"fmt"
	"net/http"
	"os"

	"audio-mixer/internal/config"
	"audio-mixer/internal/service"
	"log"

	"github.com/gin-gonic/gin"
)

// StreamRadioHandler handles GET /api/radio.
// Clients "tune in" to the live broadcast.
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
// It accepts an MP3 file via multipart form data, saves it to the "files" folder,
// waits for any ongoing YouTube conversion to finish, and then adds its path to the queue.
func AddSongHandler(c *gin.Context) {
	// Wait for any ongoing YouTube conversion to finish.
	service.WaitForYTConversion()

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "MP3 file is required (form field 'file')"})
		return
	}

	// Ensure the "files" folder exists.
	if err := os.MkdirAll("files", os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create files folder"})
		return
	}

	savePath := "files/" + file.Filename
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save MP3 file"})
		return
	}

	service.AddSong(savePath)
	c.JSON(http.StatusOK, gin.H{"message": "Song uploaded and added to queue", "path": savePath})
}

// GetQueueHandler handles GET /api/radio/queue.
// It returns the current song queue.
func GetQueueHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"queue": service.GetQueue()})
}

// AddYouTubeSongHandler handles POST /api/radio/youtube.
// It expects a JSON body like:
//
//	{"source": "youtube", "url": "https://moody.bozvpn.com/apidownload?v=2Vv-BfVoq4g"}
//
// It enqueues a YouTube conversion job in the background.
func AddYouTubeSongHandler(c *gin.Context) {
	type Request struct {
		Source string `json:"source"`
		URL    string `json:"url"`
	}
	var req Request
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	if req.Source != "youtube" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported source"})
		return
	}

	cfg := config.LoadConfig()
	service.EnqueueYTJob(req.URL, cfg)
	log.Printf("YouTube job enqueued for URL: %s", req.URL)
	c.JSON(http.StatusOK, gin.H{
		"message": "YouTube conversion job enqueued. Song will be added to queue upon completion",
	})
}
