package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL   string
	BotToken      string
	WebhookSecret string
	Port          string
}

// Load reads .env (if present) and the process environment, then validates
// that every required setting is present.
func Load() (Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	cfg := Config{
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		BotToken:      os.Getenv("BOT_TOKEN"),
		WebhookSecret: os.Getenv("TELEGRAM_WEBHOOK_SECRET"),
		Port:          os.Getenv("PORT"),
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	var missing []string
	for name, value := range map[string]string{
		"DATABASE_URL":            cfg.DatabaseURL,
		"BOT_TOKEN":               cfg.BotToken,
		"TELEGRAM_WEBHOOK_SECRET": cfg.WebhookSecret,
	} {
		if value == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return cfg, fmt.Errorf("missing required environment variable(s): %s (see .env.example)", strings.Join(missing, ", "))
	}
	return cfg, nil
}
