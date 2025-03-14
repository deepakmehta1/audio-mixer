package config

// Config holds the application settings.
type Config struct {
	Port          string
	YoutubeAPIKey string // optional API key for YouTube conversion
}

// LoadConfig loads and returns the configuration.
func LoadConfig() Config {
	return Config{
		Port:          "8080",
		YoutubeAPIKey: "", // set your API key if needed
	}
}
