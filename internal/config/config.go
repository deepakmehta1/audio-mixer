package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	YoutubeAPIKey      string
	HLSBaseURL         string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSRegion          string
}

// getEnv returns the value for a given environment variable or a fallback if not set.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() Config {
	return Config{
		Port:               getEnv("PORT", "8080"),
		YoutubeAPIKey:      getEnv("YOUTUBE_API_KEY", ""),
		HLSBaseURL:         getEnv("HLS_BASE_URL", "http://localhost:8080/hls/"),
		AWSAccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
		AWSRegion:          getEnv("AWS_REGION", "us-west-2"),
	}
}

var GlobalConfig Config

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	GlobalConfig = LoadConfig()
}
