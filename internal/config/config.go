package config

// Config holds the application settings.
type Config struct {
	Port         string
	MP3FilePath1 string
	MP3FilePath2 string
}

// LoadConfig returns a Config instance.
func LoadConfig() Config {
	// In a real app you might read from environment variables or a config file.
	return Config{
		Port:         "8080",
		MP3FilePath1: "song1.mp3",
		MP3FilePath2: "song2.mp3",
	}
}
