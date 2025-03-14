package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	YoutubeAPIKey string
	HLSBaseURL    string
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() Config {
	return Config{
		Port:          getEnv("PORT", "8080"),
		YoutubeAPIKey: getEnv("YOUTUBE_API_KEY", ""),
		HLSBaseURL:    getEnv("HLS_BASE_URL", "http://localhost:8080/hls/"),
	}
}

var GlobalConfig Config

func init() {
	// Load .env file if it exists.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	GlobalConfig = LoadConfig()
}
