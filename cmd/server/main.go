package main

import (
	"log"
	"net/http"

	"telegram_chabot/internal/bot"
	"telegram_chabot/internal/config"
	"telegram_chabot/internal/db"
	"telegram_chabot/internal/webhook"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	database := db.Connect(cfg.DatabaseURL)
	defer database.Close()

	handler := &bot.Handler{DB: database}

	mux := http.NewServeMux()
	mux.Handle("/webhook", webhook.Handler(cfg.WebhookSecret, handler.HandleUpdate))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	log.Println("Server running on port", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
