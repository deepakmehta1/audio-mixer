package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds the application settings.
type Config struct {
	Port          string
	MP3FilePath   string
	YoutubeAPIKey string
}

// LoadConfig loads configuration settings from a .env file and environment variables.
// It returns a Config instance with defaults if variables are missing.
func LoadConfig() Config {
	// Attempt to load .env file; log a message if it's not found.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default environment variables")
	}

	config := Config{
		Port:          os.Getenv("PORT"),
		MP3FilePath:   os.Getenv("MP3FilePath"),
		YoutubeAPIKey: os.Getenv("YOUTUBE_API_KEY"),
	}
	log.Println("config ", config)
	return config
}
