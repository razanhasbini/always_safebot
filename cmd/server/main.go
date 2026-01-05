package main

import (
	"io"
	"log"
	"net/http"

	"telegram_chabot/internal/bot"
	"telegram_chabot/internal/config"
	"telegram_chabot/internal/db"
)

func main() {
	cfg := config.Load()

	database := db.Connect(cfg.DatabaseURL)
	defer database.Close()

	handler := &bot.Handler{DB: database}
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		log.Println("WEBHOOK HIT")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		handler.HandleUpdate(body)
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	log.Println("Server running on port", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}
