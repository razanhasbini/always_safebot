package bot

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
)

func SendMessage(chatID int64, text string) {
	url := "https://api.telegram.org/bot" + os.Getenv("BOT_TOKEN") + "/sendMessage"
	payload := map[string]any{"chat_id": chatID, "text": text}
	body, _ := json.Marshal(payload)
	http.Post(url, "application/json", bytes.NewBuffer(body))
}

func SendLocation(chatID int64, lat, lon float64) {
	url := "https://api.telegram.org/bot" + os.Getenv("BOT_TOKEN") + "/sendLocation"
	payload := map[string]any{
		"chat_id":   chatID,
		"latitude":  lat,
		"longitude": lon,
	}
	body, _ := json.Marshal(payload)
	http.Post(url, "application/json", bytes.NewBuffer(body))
}

func SendLiveLocation(chatID int64, lat, lon float64, liveSeconds int) {
	url := "https://api.telegram.org/bot" + os.Getenv("BOT_TOKEN") + "/sendLocation"
	payload := map[string]any{
		"chat_id":     chatID,
		"latitude":    lat,
		"longitude":   lon,
		"live_period": liveSeconds,
	}
	body, _ := json.Marshal(payload)
	http.Post(url, "application/json", bytes.NewBuffer(body))
}
