package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"telegram_chabot/internal/bot"
	"telegram_chabot/internal/db"
)

func main() {
	database := db.Connect(os.Getenv("DATABASE_URL"))
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

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is required")
	}

	log.Println("Server running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
