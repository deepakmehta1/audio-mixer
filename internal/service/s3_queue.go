package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"audio-mixer/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// ScheduleS3QueueRefresh starts a background job that refreshes the regular queue
// from S3 at the specified interval.

// LoadRegularQueueFromS3Listing lists objects from the specified S3 bucket and prefix,
// downloads them locally if they don't exist, and stores their local paths in the regularQueue.
func RefreshRegularQueueFromS3Listing(bucketName, prefix string) error {
	ctx := context.TODO()
	// Load AWS configuration with the region from GlobalConfig.
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(config.GlobalConfig.AWSRegion),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %v", err)
	}

	// Create an S3 client with the loaded configuration.
	s3Client := s3.NewFromConfig(cfg)

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(prefix),
	}

	result, err := s3Client.ListObjectsV2(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to list objects in s3://%s/%s: %v", bucketName, prefix, err)
	}

	// Ensure local folder exists.
	if err := os.MkdirAll("files", os.ModePerm); err != nil {
		return fmt.Errorf("failed to create local files folder: %v", err)
	}

	// Build a map of already-enqueued local file paths.
	queueMutex.Lock()
	existing := make(map[string]bool)
	for _, s := range regularQueue {
		existing[s] = true
	}
	queueMutex.Unlock()

	// Process each object one by one.
	for _, obj := range result.Contents {
		key := *obj.Key
		if !strings.HasSuffix(strings.ToLower(key), ".mp3") {
			continue
		}
		// Extract the base filename.
		parts := strings.Split(key, "/")
		localFilename := parts[len(parts)-1]
		localPath := "files/" + localFilename

		// Skip if the file is already enqueued.
		if existing[localPath] {
			continue
		}

		// Check if file exists locally.
		if _, err := os.Stat(localPath); err == nil {
			log.Printf("Local file exists: %s. Enqueuing.", localPath)
			AddRegularSong(localPath)
		} else {
			// Download from S3.
			getObjInput := &s3.GetObjectInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(key),
			}
			objResp, err := s3Client.GetObject(ctx, getObjInput)
			if err != nil {
				log.Printf("Failed to download s3://%s/%s: %v", bucketName, key, err)
				continue
			}
			outFile, err := os.Create(localPath)
			if err != nil {
				log.Printf("Failed to create local file %s: %v", localPath, err)
				objResp.Body.Close()
				continue
			}
			_, err = io.Copy(outFile, objResp.Body)
			objResp.Body.Close()
			outFile.Close()
			if err != nil {
				log.Printf("Error saving file %s: %v", localPath, err)
				continue
			}
			log.Printf("Downloaded s3://%s/%s to %s", bucketName, key, localPath)
			// Enqueue the new song immediately.
			AddRegularSong(localPath)
		}
	}

	return nil
}

// ScheduleS3QueueRefresh starts a background job that refreshes the regular queue
// from S3 at the specified interval by appending new songs.
func ScheduleS3QueueRefresh(bucketName, prefix string, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			err := RefreshRegularQueueFromS3Listing(bucketName, prefix)
			if err != nil {
				log.Printf("Error refreshing regular queue from S3: %v", err)
			} else {
				log.Printf("Regular queue refreshed from s3://%s/%s", bucketName, prefix)
			}
		}
	}()
}
