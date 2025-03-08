package service

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// StreamRadio streams a continuous radio-style audio stream.
// It reads MP3 files from the provided slice and writes their data continuously.
func StreamRadio(c *gin.Context, songPaths []string) error {
	// Create an io.Pipe: the writer will be fed by our song loop,
	// and the reader will be copied into the HTTP response.
	pr, pw := io.Pipe()

	// Start a goroutine to write songs continuously into the pipe.
	go func() {
		defer pw.Close()
		// Loop indefinitely over the songs.
		for {
			for _, path := range songPaths {
				// Open the MP3 file.
				f, err := os.Open(path)
				if err != nil {
					fmt.Printf("Error opening %s: %v\n", path, err)
					continue // Skip to next song if error occurs.
				}

				// Copy the file data to the pipe writer.
				// We’re streaming the MP3 as-is.
				_, err = io.Copy(pw, f)
				if err != nil {
					fmt.Printf("Error copying %s: %v\n", path, err)
					f.Close()
					continue
				}
				f.Close()
				// Optionally, add a short pause between songs:
				// time.Sleep(time.Millisecond * 100)
			}
		}
	}()

	// Set the Content-Type to audio/mpeg so clients treat it as an MP3 stream.
	c.Writer.Header().Set("Content-Type", "audio/mpeg")
	c.Status(http.StatusOK)

	// Copy from the pipe reader to the HTTP response.
	// This loop blocks and writes data as it becomes available.
	buf := make([]byte, 4096)
	for {
		n, err := pr.Read(buf)
		if n > 0 {
			_, writeErr := c.Writer.Write(buf[:n])
			if writeErr != nil {
				return fmt.Errorf("error writing to response: %v", writeErr)
			}
			// Flush to ensure immediate delivery.
			if f, ok := c.Writer.(http.Flusher); ok {
				f.Flush()
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading from pipe: %v", err)
		}
	}
	return nil
}
