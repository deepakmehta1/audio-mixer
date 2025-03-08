package service

import (
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Global song queue and mutex for safe concurrent access.
var (
	songQueue  []string
	queueMutex sync.Mutex
)

// AddSong adds a new song path to the global song queue.
func AddSong(path string) {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	songQueue = append(songQueue, path)
	log.Printf("Added song to queue: %s", path)
}

// GetQueue returns a copy of the current song queue.
func GetQueue() []string {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	queueCopy := make([]string, len(songQueue))
	copy(queueCopy, songQueue)
	return queueCopy
}

// Global skip channel (buffered so sends are non-blocking) for real-time skip.
var skipChan = make(chan struct{}, 1)

// SkipCurrentSong sends a signal to skip the current song.
func SkipCurrentSong() {
	select {
	case skipChan <- struct{}{}:
		log.Println("Skip signal sent")
	default:
		log.Println("Skip signal already pending")
	}
}

// StreamRadio streams a continuous radio-style audio stream using the dynamic queue.
// It reads songs from the global queue and streams their data in small chunks,
// checking for a skip signal on each chunk read.
func StreamRadio(c *gin.Context, _ []string) error {
	// Create an io.Pipe: writer (pw) receives song data; reader (pr) is copied to the response.
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		for {
			// Get the current dynamic queue.
			currentQueue := GetQueue()
			if len(currentQueue) == 0 {
				log.Println("Queue is empty, waiting for songs...")
				time.Sleep(1 * time.Second)
				continue
			}

			// Loop through songs in the dynamic queue.
			for _, path := range currentQueue {
				log.Printf("Starting song: %s", path)
				f, err := os.Open(path)
				if err != nil {
					log.Printf("Error opening %s: %v", path, err)
					continue // Skip this song if it can't be opened.
				}

				buf := make([]byte, 4096)
			readLoop:
				for {
					// Check for a skip signal.
					select {
					case <-skipChan:
						log.Printf("Skip signal received, skipping song: %s", path)
						break readLoop
					default:
						// No skip signal; continue reading.
					}

					n, err := f.Read(buf)
					if n > 0 {
						if _, wErr := pw.Write(buf[:n]); wErr != nil {
							log.Printf("Error writing to pipe: %v", wErr)
							break readLoop
						}
					}
					if err == io.EOF {
						log.Printf("Finished song: %s", path)
						break // Finished reading current song.
					}
					if err != nil {
						log.Printf("Error reading from file %s: %v", path, err)
						break readLoop
					}

					// Sleep a short time to simulate real-time streaming.
					time.Sleep(50 * time.Millisecond)
				}
				f.Close()
			}
		}
	}()

	// Set the Content-Type header to indicate an MP3 stream.
	c.Writer.Header().Set("Content-Type", "audio/mpeg")
	c.Status(http.StatusOK)

	// Continuously copy data from the pipe to the HTTP response.
	buf := make([]byte, 1)
	for {
		n, err := pr.Read(buf)
		if n > 0 {
			_, writeErr := c.Writer.Write(buf[:n])
			if writeErr != nil {
				log.Printf("Client disconnected (write error): %v", writeErr)
				break // Stop streaming if the client disconnects.
			}
			if f, ok := c.Writer.(http.Flusher); ok {
				f.Flush()
			}
		}
		if err == io.EOF {
			log.Println("Pipe reader reached EOF")
			break
		}
		if err != nil {
			log.Printf("Error reading from pipe: %v", err)
			break
		}
	}

	return nil
}
