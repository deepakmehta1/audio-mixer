package service

import (
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"audio-mixer/internal/config"

	"github.com/gin-gonic/gin"
)

var skipChan = make(chan struct{}, 1)
var stopCurrentSongChan = make(chan bool, 1)

// SkipCurrentSong signals to stop writing the current file.
func SkipCurrentSong() {
	select {
	case skipChan <- struct{}{}:
		log.Println("Skip signal sent.")
	default:
		log.Println("Skip signal already pending.")
	}
}

// StartStreaming sets up a single FFmpeg pipeline reading from a named pipe.
func StartStreaming() {
	go func() {
		// 1) Create the hls folder and pipe if needed.
		os.RemoveAll("./hls")
		os.MkdirAll("./hls", 0755)

		pipePath := "./hls/radio_input.fifo"

		// On Unix-like systems, create a named pipe if it doesn't exist.
		if _, err := os.Stat(pipePath); os.IsNotExist(err) {
			if err := exec.Command("mkfifo", pipePath).Run(); err != nil {
				log.Fatalf("Failed to create named pipe: %v", err)
			}
		}

		// 2) Start FFmpeg in one continuous process reading from the pipe.
		ffmpegCmd := buildFFmpegCommand(pipePath)
		if err := ffmpegCmd.Start(); err != nil {
			log.Fatalf("Error starting ffmpeg: %v", err)
		}
		log.Println("FFmpeg started with single pipeline reading from pipe...")

		// 3) Goroutine to wait if FFmpeg ever ends (it shouldn't unless error).
		go func() {
			if err := ffmpegCmd.Wait(); err != nil {
				log.Printf("FFmpeg ended with error: %v", err)
			} else {
				log.Println("FFmpeg ended normally.")
			}
		}()

		// 4) Another goroutine to open the pipe for writing and feed songs.
		go feedSongsToPipe(pipePath)
	}()
}

// feedSongsToPipe opens the named pipe for writing, then continuously reads
// songs from the queue, writes them to the pipe, and handles skip signals.
func feedSongsToPipe(pipePath string) {
	for {
		// Open the pipe for writing (blocks until the reading end is open).
		pipeFile, err := os.OpenFile(pipePath, os.O_WRONLY, 0600)
		if err != nil {
			log.Printf("Error opening pipe for writing: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		// feed each track in a loop
		for {
			path := NextSong()
			if path == "" {
				log.Println("No songs in queue, waiting...")
				time.Sleep(1 * time.Second)
				continue
			}
			log.Printf("Feeding song into pipe: %s", path)

			// open the MP3 file
			f, err := os.Open(path)
			if err != nil {
				log.Printf("Error opening file %s: %v", path, err)
				continue
			}

			doneSong := make(chan bool)

			// Write the MP3 data to the pipe in a loop.
			go func() {
				defer close(doneSong)
				defer f.Close()

				buf := make([]byte, 4096)
				for {
					select {
					case <-skipChan:
						// skip signal => stop reading this file
						log.Printf("Skipping current file: %s", path)
						return
					default:
					}

					n, err := f.Read(buf)
					if n > 0 {
						if _, werr := pipeFile.Write(buf[:n]); werr != nil {
							log.Printf("Error writing to pipe: %v", werr)
							return
						}
					}
					if err == io.EOF {
						// finished this song
						log.Printf("Finished feeding song: %s", path)
						return
					}
					if err != nil && err != io.EOF {
						log.Printf("Error reading file %s: %v", path, err)
						return
					}
				}
			}()

			<-doneSong
			// once we finish or skip, move on to next track
		}
		// if pipeFile is closed for any reason, loop tries to reopen it
	}
}

// buildFFmpegCommand constructs the ffmpeg command that reads from the named pipe
// and writes HLS segments + manifest to ./hls/.
func buildFFmpegCommand(pipePath string) *exec.Cmd {
	baseURL := config.GlobalConfig.HLSBaseURL
	cmd := exec.Command("ffmpeg",
		"-re",
		"-i", pipePath,
		"-c:a", "aac",
		"-b:a", "192k",
		"-hls_time", "4",
		"-hls_list_size", "0",
		"-force_key_frames", "expr:gte(t,n_forced*2)",
		"-hls_segment_filename", "./hls/hls_%03d.ts",
		"-hls_base_url", baseURL,
		"./hls/index.m3u8",
	)
	return cmd
}

// StreamRadio is the HTTP handler for GET /api/radio
// which serves the HLS manifest so clients can tune in.
func StreamRadio(c *gin.Context) error {
	if _, err := os.Stat("./hls/index.m3u8"); os.IsNotExist(err) {
		c.String(404, "HLS stream not ready")
		return nil
	}
	c.File("./hls/index.m3u8")
	return nil
}
