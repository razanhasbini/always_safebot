package db

import "database/sql"

func SetContext(db *sql.DB, userID int64, name string, active bool) error {
	_, err := db.Exec(`
		INSERT INTO contexts (user_id, name, is_active)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, name)
		DO UPDATE SET is_active = EXCLUDED.is_active, updated_at = now()
	`, userID, name, active)
	return err
}

func IsContextActive(db *sql.DB, userID int64, name string) (bool, error) {
	var active bool
	err := db.QueryRow(`
		SELECT is_active
		FROM contexts
		WHERE user_id = $1 AND name = $2
	`, userID, name).Scan(&active)

	if err == sql.ErrNoRows {
		return false, nil
	}
	return active, err
}
