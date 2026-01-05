package db

import "database/sql"

type User struct {
	ID         int64
	TelegramID int64
	Username   string
	FirstName  string
	LastName   string
}

func GetOrCreateUser(db *sql.DB, tgID int64, username, firstName, lastName string) (*User, error) {
	var user User

	err := db.QueryRow(`
		SELECT id, telegram_id, username, first_name, last_name
		FROM users
		WHERE telegram_id = $1
	`, tgID).Scan(
		&user.ID,
		&user.TelegramID,
		&user.Username,
		&user.FirstName,
		&user.LastName,
	)

	if err == sql.ErrNoRows {
		err = db.QueryRow(`
			INSERT INTO users (telegram_id, username, first_name, last_name)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, tgID, username, firstName, lastName).Scan(&user.ID)

		if err != nil {
			return nil, err
		}

		user.TelegramID = tgID
		user.Username = username
		user.FirstName = firstName
		user.LastName = lastName
	}

	return &user, nil
}

func GetUserDisplayName(db *sql.DB, userID int64) (string, error) {
	var firstName, username sql.NullString

	err := db.QueryRow(`
		SELECT first_name, username
		FROM users
		WHERE id = $1
	`, userID).Scan(&firstName, &username)

	if err != nil {
		return "", err
	}

	if firstName.Valid && firstName.String != "" {
		return firstName.String, nil
	}
	if username.Valid && username.String != "" {
		return username.String, nil
	}

	return "this person", nil
}

func FindUserByFirstName(
	db *sql.DB,
	firstName string,
) (int64, bool, error) {

	var id int64
	err := db.QueryRow(`
		SELECT id
		FROM users
		WHERE LOWER(first_name) = LOWER($1)
		LIMIT 1
	`, firstName).Scan(&id)

	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}

	return id, true, nil
}
