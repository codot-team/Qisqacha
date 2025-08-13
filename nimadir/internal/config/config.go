package config

import (
	"os"
)

type Config struct {
	TelegramToken string
	GeminiAPIKey  string
}

func Load() *Config {
	return &Config{
		TelegramToken: os.Getenv("TELEGRAM_TOKEN"),
		GeminiAPIKey:  os.Getenv("GEMINI_API_KEY"),
	}
}
