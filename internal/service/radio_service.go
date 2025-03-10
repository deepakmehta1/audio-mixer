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

// Global skip channel (buffered so that sends are non-blocking)
var skipChan = make(chan struct{}, 1)

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

// SkipCurrentSong sends a signal to skip the current song.
func SkipCurrentSong() {
	select {
	case skipChan <- struct{}{}:
		log.Println("Skip signal sent.")
	default:
		log.Println("Skip signal already pending.")
	}
}

// StartStreaming starts a background goroutine that continuously reads songs from the queue
// and broadcasts their data to the global broadcaster. This runs regardless of client connections.
func StartStreaming() {
	go func() {
		for {
			// Get a snapshot of the current queue.
			queueMutex.Lock()
			currentQueue := make([]string, len(songQueue))
			copy(currentQueue, songQueue)
			queueMutex.Unlock()

			if len(currentQueue) == 0 {
				log.Println("Queue is empty, waiting for songs...")
				time.Sleep(1 * time.Second)
				continue
			}

			// Loop through songs in the current queue.
			for _, path := range currentQueue {
				log.Printf("Playing song: %s", path)
				f, err := os.Open(path)
				if err != nil {
					log.Printf("Error opening %s: %v", path, err)
					continue
				}

				// Create a channel to receive data chunks from the file reader.
				chunkChan := make(chan []byte)
				// Launch a goroutine to read the file in small chunks.
				go func(file *os.File, out chan<- []byte, filePath string) {
					defer close(out)
					buf := make([]byte, 512) // smaller chunk size for frequent skip checks
					for {
						n, err := file.Read(buf)
						if n > 0 {
							// Copy the data before sending.
							data := make([]byte, n)
							copy(data, buf[:n])
							out <- data
						}
						if err == io.EOF {
							break
						}
						if err != nil {
							log.Printf("Error reading from file %s: %v", filePath, err)
							break
						}
						// Sleep briefly to slow the stream rate.
						time.Sleep(15 * time.Millisecond)
					}
				}(f, chunkChan, path)

			readLoop:
				for {
					select {
					case <-skipChan:
						log.Printf("Skip signal received, skipping song: %s", path)
						break readLoop
					case chunk, ok := <-chunkChan:
						if !ok {
							log.Printf("Finished song: %s", path)
							break readLoop
						}
						// Broadcast the chunk.
						GlobalBroadcaster.Broadcast(chunk)
					}
				}
				f.Close()
			}
		}
	}()
}

// StreamRadio is the HTTP handler used by new subscribers.
// It subscribes to the global broadcaster and writes data from the live stream
// to the HTTP response, allowing clients to "tune in" live.
func StreamRadio(c *gin.Context) error {
	sub := GlobalBroadcaster.Subscribe()
	defer GlobalBroadcaster.Unsubscribe(sub)

	c.Writer.Header().Set("Content-Type", "audio/mpeg")
	c.Status(http.StatusOK)

	for {
		select {
		case data, ok := <-sub:
			if !ok {
				return nil
			}
			_, err := c.Writer.Write(data)
			if err != nil {
				log.Printf("Error writing to HTTP response: %v", err)
				return err
			}
			if f, ok := c.Writer.(http.Flusher); ok {
				f.Flush()
			}
		case <-c.Request.Context().Done():
			return nil
		}
	}
}
