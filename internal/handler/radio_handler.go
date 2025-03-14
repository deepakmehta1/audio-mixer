package handler

import (
	"net/http"
	"os"

	"audio-mixer/internal/config"
	"audio-mixer/internal/service"
	"log"

	"github.com/gin-gonic/gin"
)

// StreamRadioHandler handles GET /api/radio.
// It serves the HLS playlist (index.m3u8) so that VLC can play the stream.
func StreamRadioHandler(c *gin.Context) {
	// Check if the HLS manifest exists.
	if _, err := os.Stat("./hls/index.m3u8"); os.IsNotExist(err) {
		c.String(http.StatusNotFound, "HLS stream not ready")
		return
	}
	// Option: serve the manifest file.
	c.File("./hls/index.m3u8")
}

// SkipRadioHandler handles POST /api/radio/skip.
// It sends a signal to skip the current song.
func SkipRadioHandler(c *gin.Context) {
	service.SkipCurrentSong()
	c.String(http.StatusOK, "Skip signal sent.")
}

func GetPriorityQueueHandler(c *gin.Context) {
	priorityQ := service.GetPriorityQueue()
	c.JSON(http.StatusOK, gin.H{"queue": priorityQ})
}

// AddPrioritySongHandler handles POST /api/radio/priority.
// It accepts an MP3 file via multipart form data, saves it to "files", and adds it to the priority queue.
func AddPrioritySongHandler(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "MP3 file is required (form field 'file')"})
		return
	}
	if err := os.MkdirAll("files", os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create files folder"})
		return
	}
	savePath := "files/" + file.Filename
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save MP3 file"})
		return
	}
	if err := service.AddPrioritySong(savePath); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Priority song uploaded and added to queue", "path": savePath})
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
		"message": "YouTube conversion job enqueued. Song will be added to regular queue upon completion",
	})
}
