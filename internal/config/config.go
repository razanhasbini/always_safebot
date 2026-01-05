package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	BotToken    string
	Port        string
}

func Load() Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system env vars")
	}

	return Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		BotToken:    os.Getenv("BOT_TOKEN"),
		Port:        os.Getenv("PORT"),
	}
}
