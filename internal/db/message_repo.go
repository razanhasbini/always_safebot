package db

import "database/sql"

func SaveMessageEvent(
	db *sql.DB,
	userID int64,
	senderTelegramID int64,
	chatID int64,
	text string,
	rawPayload []byte,
) error {

	_, err := db.Exec(`
		INSERT INTO message_events (
			user_id,
			sender_telegram_id,
			chat_id,
			message_text,
			raw_payload
		)
		VALUES ($1, $2, $3, $4, $5)
	`,
		userID,
		senderTelegramID,
		chatID,
		text,
		rawPayload,
	)

	return err
}
