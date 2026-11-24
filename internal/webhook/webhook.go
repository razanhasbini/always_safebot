// Package webhook protects the Telegram webhook endpoint.
//
// Telegram sends the secret configured with setWebhook(secret_token=...) in the
// X-Telegram-Bot-Api-Secret-Token header of every update. Requests without the
// matching secret are rejected, so nobody else can post fake updates to the bot.
package webhook

import (
	"crypto/subtle"
	"io"
	"net/http"
)

// SecretHeader is the header Telegram uses to carry the webhook secret token.
const SecretHeader = "X-Telegram-Bot-Api-Secret-Token"

// MaxBodyBytes caps the size of an incoming update.
const MaxBodyBytes = 1 << 20

// Handler returns an http.Handler that accepts only POST requests carrying the
// expected secret and passes the request body to handle.
func Handler(secret string, handle func([]byte)) http.Handler {
	expected := []byte(secret)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		got := []byte(r.Header.Get(SecretHeader))
		if len(expected) == 0 || subtle.ConstantTimeCompare(got, expected) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, MaxBodyBytes))
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		handle(body)
		w.WriteHeader(http.StatusOK)
	})
}
