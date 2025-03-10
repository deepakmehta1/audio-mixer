package handler

import (
	"fmt"
	"net/http"
	"os"

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
// It now accepts an MP3 file via multipart form upload,
// stores it in the "files/" folder, and then adds the saved file path to the queue.
func AddSongHandler(c *gin.Context) {
	// Retrieve the file from the "file" form field.
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "MP3 file is required (form field 'file')"})
		return
	}

	// Optionally, you can check the file extension here.
	// For example, ensure file.Filename ends with ".mp3"

	// Ensure the "files" folder exists.
	if err := os.MkdirAll("files", os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create files folder"})
		return
	}

	// Save the uploaded file in the "files/" folder.
	savePath := "files/" + file.Filename
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save MP3 file"})
		return
	}

	// Add the saved file path to the global song queue.
	service.AddSong(savePath)
	c.JSON(http.StatusOK, gin.H{"message": "Song uploaded and added to queue", "path": savePath})
}

// GetQueueHandler handles GET /api/radio/queue.
// It returns the current song queue.
func GetQueueHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"queue": service.GetQueue()})
}
