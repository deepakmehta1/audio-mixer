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
// and broadcasts their data to the global broadcaster.
func StartStreaming() {
	go func() {
		for {
			// Pop the first song from the global queue.
			queueMutex.Lock()
			if len(songQueue) == 0 {
				queueMutex.Unlock()
				log.Println("Queue is empty, waiting for songs...")
				time.Sleep(1 * time.Second)
				continue
			}
			// Remove the first song from the queue.
			path := songQueue[0]
			songQueue = songQueue[1:]
			queueMutex.Unlock()

			log.Printf("Playing song: %s", path)
			f, err := os.Open(path)
			if err != nil {
				log.Printf("Error opening %s: %v", path, err)
				continue
			}

			chunkChan := make(chan []byte)
			go func(file *os.File, out chan<- []byte, filePath string) {
				defer close(out)
				buf := make([]byte, 512) // chunk size for frequent skip checks
				for {
					n, err := file.Read(buf)
					if n > 0 {
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
					GlobalBroadcaster.Broadcast(chunk)
				}
			}
			f.Close()
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
