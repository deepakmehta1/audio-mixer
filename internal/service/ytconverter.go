package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"audio-mixer/internal/config"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// ytJob holds a YouTube conversion job.
type ytJob struct {
	url string
	cfg config.Config
}

// Global channel for YouTube conversion jobs.
var ytJobChan = make(chan ytJob, 10)

// Global mutex to signal a YouTube conversion is in progress.
var ytConversionLock sync.Mutex

// DownloadResponse represents the JSON response from the external API.
type DownloadResponse struct {
	URL       string `json:"url"`
	ExpiresAt string `json:"expiresAt"`
}

// StartYTWorker starts a background worker that processes YouTube conversion jobs.
func StartYTWorker() {
	go func() {
		for job := range ytJobChan {
			// Lock to signal a conversion is running.
			ytConversionLock.Lock()
			log.Printf("Processing YouTube conversion job: %s", job.url)
			mp3Path, err := ConvertYouTubeToMP3(job.url, job.cfg)
			if err != nil {
				log.Printf("Error converting YouTube media: %v", err)
				ytConversionLock.Unlock()
				continue
			}
			AddSong(mp3Path)
			log.Printf("YouTube conversion finished, added file to queue: %s", mp3Path)
			ytConversionLock.Unlock()
		}
	}()
}

// EnqueueYTJob enqueues a YouTube conversion job along with its configuration.
func EnqueueYTJob(url string, cfg config.Config) {
	ytJobChan <- ytJob{
		url: url,
		cfg: cfg,
	}
}

// WaitForYTConversion waits until any ongoing YouTube conversion finishes.
func WaitForYTConversion() {
	ytConversionLock.Lock()
	ytConversionLock.Unlock()
}

// ConvertYouTubeToMP3 downloads a YouTube video via an external API and converts it to MP3.
// It returns the path to the resulting MP3 file.
func ConvertYouTubeToMP3(youtubeURL string, cfg config.Config) (string, error) {
	timestamp := time.Now().Unix()
	inputFile := fmt.Sprintf("files/yt_media_%d.webm", timestamp)
	outputFile := fmt.Sprintf("files/yt_media_%d.mp3", timestamp)

	// Ensure the "files" folder exists.
	if err := os.MkdirAll("files", os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create files folder: %v", err)
	}

	// Construct the external API URL.
	apiURL := fmt.Sprintf("https://zylalabs.com/api/6264/youtube+search+download+api/8850/download?v=%s", youtubeURL)

	// Create a new HTTP request.
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}
	// Add Authorization header with bearer token if provided.
	if cfg.YoutubeAPIKey != "" {
		req.Header.Add("Authorization", "Bearer "+cfg.YoutubeAPIKey)
	}

	// Execute the request.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error calling download API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download API returned status: %s", resp.Status)
	}

	var dlResp DownloadResponse
	if err := json.NewDecoder(resp.Body).Decode(&dlResp); err != nil {
		return "", fmt.Errorf("error decoding API response: %v", err)
	}

	log.Printf("Download API returned URL: %s (expires at %s)", dlResp.URL, dlResp.ExpiresAt)

	// Download the video using the returned URL.
	out, err := os.Create(inputFile)
	if err != nil {
		return "", fmt.Errorf("error creating file: %v", err)
	}
	defer out.Close()

	downloadResp, err := http.Get(dlResp.URL)
	if err != nil {
		return "", fmt.Errorf("error downloading video from API URL: %v", err)
	}
	defer downloadResp.Body.Close()

	if _, err := io.Copy(out, downloadResp.Body); err != nil {
		return "", fmt.Errorf("error saving video: %v", err)
	}

	// Convert the downloaded file to MP3 using ffmpeg-go.
	err = ffmpeg.Input(inputFile).
		Output(outputFile, ffmpeg.KwArgs{
			"vn":  "",
			"ar":  "44100",
			"ac":  "2",
			"b:a": "192k",
		}).
		OverWriteOutput().
		Run()
	if err != nil {
		return "", fmt.Errorf("error converting file: %v", err)
	}

	log.Printf("Conversion successful: %s", outputFile)
	return outputFile, nil
}
