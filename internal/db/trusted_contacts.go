package db

import "database/sql"

func IsTrustedContact(db *sql.DB, userID int64, senderTelegramID int64) (bool, error) {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM trusted_contacts
			WHERE user_id = $1
			  AND contact_telegram_id = $2
		)
	`, userID, senderTelegramID).Scan(&exists)

	return exists, err
}

func GetOwnersForRequester(
	db *sql.DB,
	requesterTelegramID int64,
) ([]int64, error) {

	rows, err := db.Query(`
		SELECT user_id
		FROM trusted_contacts
		WHERE contact_telegram_id = $1
	`, requesterTelegramID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var owners []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		owners = append(owners, id)
	}
	return owners, nil
}
func AddTrustedContact(
	db *sql.DB,
	userID int64,
	contactTelegramID int64,
	contactName string,
) error {

	_, err := db.Exec(`
		INSERT INTO trusted_contacts (user_id, contact_telegram_id, contact_name)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
	`, userID, contactTelegramID, contactName)

	return err
}
func RemoveTrustedContact(
	db *sql.DB,
	userID int64,
	contactTelegramID int64,
) error {

	_, err := db.Exec(`
		DELETE FROM trusted_contacts
		WHERE user_id = $1
		  AND contact_telegram_id = $2
	`, userID, contactTelegramID)

	return err
}
