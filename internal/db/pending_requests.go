package db

import (
	"database/sql"
	"time"
)

type PendingRequest struct {
	ID                  int64
	OwnerUserID         int64
	RequesterTelegramID int64
	ChatID              int64
	RequestType         string
	Status              string
	ExpiresAt           time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// CreatePendingRequest creates a new pending request and returns its ID
func CreatePendingRequest(
	db *sql.DB,
	ownerUserID int64,
	requesterTelegramID int64,
	chatID int64,
	expiresAt time.Time,
) (int64, error) {

	var id int64

	err := db.QueryRow(`
		INSERT INTO pending_requests (
			owner_user_id,
			requester_telegram_id,
			chat_id,
			request_type,
			status,
			expires_at
		)
		VALUES ($1, $2, $3, 'location_share', 'pending', $4)
		RETURNING id
	`,
		ownerUserID,
		requesterTelegramID,
		chatID,
		expiresAt,
	).Scan(&id)

	return id, err
}

// GetLatestPendingRequest returns the most recent non-expired pending request
func GetLatestPendingRequest(
	db *sql.DB,
	requesterTelegramID int64,
) (*PendingRequest, error) {

	row := db.QueryRow(`
		SELECT
			id,
			owner_user_id,
			requester_telegram_id,
			chat_id,
			request_type,
			status,
			expires_at,
			created_at,
			updated_at
		FROM pending_requests
		WHERE
			requester_telegram_id = $1
			AND status = 'pending'
			AND expires_at > now()
		ORDER BY created_at DESC
		LIMIT 1
	`, requesterTelegramID)

	var pr PendingRequest
	err := row.Scan(
		&pr.ID,
		&pr.OwnerUserID,
		&pr.RequesterTelegramID,
		&pr.ChatID,
		&pr.RequestType,
		&pr.Status,
		&pr.ExpiresAt,
		&pr.CreatedAt,
		&pr.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &pr, nil
}

// MarkPendingRequestCompleted marks a request as completed
func MarkPendingRequestCompleted(
	db *sql.DB,
	id int64,
) error {

	_, err := db.Exec(`
		UPDATE pending_requests
		SET status = 'completed',
		    updated_at = now()
		WHERE id = $1
	`, id)

	return err
}

// CancelPendingRequestsForRequester cancels all active pending requests for a requester
func CancelPendingRequestsForRequester(
	db *sql.DB,
	requesterTelegramID int64,
) error {

	_, err := db.Exec(`
		UPDATE pending_requests
		SET status = 'cancelled',
		    updated_at = now()
		WHERE
			requester_telegram_id = $1
			AND status = 'pending'
	`, requesterTelegramID)

	return err
}
