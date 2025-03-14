package service

import (
	"log"
	"os"
	"os/exec"
	"time"

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
			// Clear the HLS folder.
			os.RemoveAll("./hls")
			os.MkdirAll("./hls", 0755)

			// Build an FFmpeg command to generate HLS segments.
			// -re: read input at native frame rate
			// -hls_time 5: each segment ~5 seconds
			// -hls_list_size 0: list all segments
			// -hls_segment_filename: template for segment filenames
			cmd := exec.Command("ffmpeg",
				"-re",
				"-i", path,
				"-c:a", "aac",
				"-b:a", "192k",
				"-hls_time", "4",
				"-hls_list_size", "0",
				"-force_key_frames", "expr:gte(t,n_forced*2)",
				"-hls_segment_filename", "./hls/hls_%03d.ts",
				"-hls_base_url", "http://localhost:8080/hls/", // <--- Important
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
			// After ffmpeg is done (either naturally or via skip), loop to next song.
		}
	}()
}

// StreamRadio is the HTTP handler used by new subscribers.
// It simply serves the HLS manifest (index.m3u8) from the hls folder.
func StreamRadio(c *gin.Context) error {
	// Check if the HLS manifest exists.
	if _, err := os.Stat("./hls/index.m3u8"); os.IsNotExist(err) {
		c.String(404, "HLS stream not ready")
		return nil
	}
	// Serve the HLS manifest file.
	c.File("./hls/index.m3u8")
	return nil
}
