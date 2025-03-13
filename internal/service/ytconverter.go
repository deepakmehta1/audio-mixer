package service

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/kkdai/youtube/v2"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// Global channel for YouTube conversion jobs.
var ytJobChan = make(chan string, 10)

// StartYTWorker starts a background worker that processes YouTube conversion jobs.
func StartYTWorker() {
	go func() {
		for url := range ytJobChan {
			log.Printf("Processing YouTube conversion job: %s", url)
			mp3Path, err := ConvertYouTubeToMP3(url)
			if err != nil {
				log.Printf("Error converting YouTube media: %v", err)
				continue
			}
			AddSong(mp3Path)
			log.Printf("YouTube conversion finished, added file to queue: %s", mp3Path)
		}
	}()
}

// EnqueueYTJob enqueues a YouTube conversion job.
func EnqueueYTJob(url string) {
	ytJobChan <- url
}

// ConvertYouTubeToMP3 downloads a YouTube video and converts it to MP3.
// It returns the path to the resulting MP3 file.
func ConvertYouTubeToMP3(url string) (string, error) {
	timestamp := time.Now().Unix()
	inputFile := fmt.Sprintf("files/yt_media_%d.webm", timestamp)
	outputFile := fmt.Sprintf("files/yt_media_%d.mp3", timestamp)

	// Ensure the "files" folder exists.
	if err := os.MkdirAll("files", os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create files folder: %v", err)
	}

	client := youtube.Client{}
	video, err := client.GetVideo(url)
	if err != nil {
		return "", fmt.Errorf("error getting video: %v", err)
	}

	// Look for a format with audio channels that has Itag 251.
	formats := video.Formats.WithAudioChannels()
	var format *youtube.Format
	for i, f := range formats {
		if f.ItagNo == 251 {
			format = &formats[i]
			break
		}
	}
	if format == nil {
		if len(formats) == 0 {
			return "", fmt.Errorf("no audio formats available")
		}
		// Fallback: use the first available format with audio.
		format = &formats[0]
	}

	// Create the file to store the downloaded video.
	f, err := os.Create(inputFile)
	if err != nil {
		return "", fmt.Errorf("error creating file: %v", err)
	}
	defer f.Close()

	// Download the video into the file.
	stream, _, err := client.GetStream(video, format)
	if err != nil {
		return "", fmt.Errorf("error getting video stream: %v", err)
	}
	if _, err := io.Copy(f, stream); err != nil {
		return "", fmt.Errorf("error downloading video: %v", err)
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
