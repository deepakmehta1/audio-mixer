package service

import (
	"log"
	"os"
	"os/exec"
	"time"

	"audio-mixer/internal/config"

	"github.com/gin-gonic/gin"
)

// Global skip channel (buffered so that sends are non-blocking)
var skipChan = make(chan struct{}, 1)

// SkipCurrentSong sends a signal to skip the current song.
func SkipCurrentSong() {
	select {
	case skipChan <- struct{}{}:
		log.Println("Skip signal sent.")
	default:
		log.Println("Skip signal already pending.")
	}
}

// StartStreaming starts a background goroutine that continuously plays songs
// from the queues (priority first, then regular circular queue) and generates HLS segments.
func StartStreaming() {
	go func() {
		for {
			path := NextSong()
			if path == "" {
				log.Println("No songs in either queue, waiting for songs...")
				time.Sleep(1 * time.Second)
				continue
			}

			log.Printf("Playing song (HLS): %s", path)
			// Clear and prepare the HLS folder.
			os.RemoveAll("./hls")
			os.MkdirAll("./hls", 0755)

			// Build an FFmpeg command to generate HLS segments.
			// Use the base URL from configuration.
			cmd := exec.Command("ffmpeg",
				"-re",
				"-i", path,
				"-c:a", "aac",
				"-b:a", "192k",
				"-hls_time", "4",
				"-hls_list_size", "0",
				"-force_key_frames", "expr:gte(t,n_forced*2)",
				"-hls_segment_filename", "./hls/hls_%03d.ts",
				"-hls_base_url", config.GlobalConfig.HLSBaseURL,
				"./hls/index.m3u8",
			)

			if err := cmd.Start(); err != nil {
				log.Printf("Error starting ffmpeg for %s: %v", path, err)
				continue
			}

			done := make(chan error, 1)
			go func() {
				done <- cmd.Wait()
			}()

		waitLoop:
			for {
				select {
				case <-skipChan:
					log.Printf("Skip signal received for %s. Killing ffmpeg...", path)
					cmd.Process.Kill()
					break waitLoop
				case err := <-done:
					if err != nil {
						log.Printf("ffmpeg ended with error for song %s: %v", path, err)
					} else {
						log.Printf("Finished song (HLS): %s", path)
					}
					break waitLoop
				}
			}
			// After ffmpeg is done, move on to the next song.
		}
	}()
}

// StreamRadio is the HTTP handler used by new subscribers.
// It serves the HLS manifest (index.m3u8) so that clients can play the HLS stream.
func StreamRadio(c *gin.Context) error {
	if _, err := os.Stat("./hls/index.m3u8"); os.IsNotExist(err) {
		c.String(404, "HLS stream not ready")
		return nil
	}
	c.File("./hls/index.m3u8")
	return nil
}
