package bot

type Update struct {
	UpdateID int      `json:"update_id"`
	Message  *Message `json:"message"`
}

type Message struct {
	MessageID   int       `json:"message_id"`
	Text        string    `json:"text"`
	From        *User     `json:"from"`
	Chat        *Chat     `json:"chat"`
	Location    *Location `json:"location"`
	LivePeriod  *int      `json:"live_period"`
	ForwardFrom *User     `json:"forward_from"` // ✅ REQUIRED
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type Chat struct {
	ID int64 `json:"id"`
}
