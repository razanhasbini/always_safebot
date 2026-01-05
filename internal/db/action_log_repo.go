package db

import "database/sql"

func InsertActionLog(db *sql.DB, userID int64, ruleID *int64, eventID *int64, triggerText, actionType, status, reason string) error {
	_, err := db.Exec(`
		INSERT INTO action_logs (user_id, rule_id, event_id, trigger_text, action_type, status, reason)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, userID, ruleID, eventID, triggerText, actionType, status, reason)
	return err
}
