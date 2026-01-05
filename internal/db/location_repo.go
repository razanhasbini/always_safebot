package db

import (
	"database/sql"
	"time"
)

func UpsertUserLocation(
	db *sql.DB,
	userID int64,
	lat, lon float64,
	liveUntil *time.Time,
) error {
	_, err := db.Exec(`
		INSERT INTO user_locations (user_id, latitude, longitude, live_until)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id)
		DO UPDATE SET
		  latitude = EXCLUDED.latitude,
		  longitude = EXCLUDED.longitude,
		  live_until = EXCLUDED.live_until,
		  updated_at = now()
	`, userID, lat, lon, liveUntil)

	return err
}

func GetUserLocation(
	db *sql.DB,
	userID int64,
) (lat float64, lon float64, liveUntil *time.Time, ok bool, err error) {

	var lu sql.NullTime

	err = db.QueryRow(`
		SELECT latitude, longitude, live_until
		FROM user_locations
		WHERE user_id = $1
	`, userID).Scan(&lat, &lon, &lu)

	if err == sql.ErrNoRows {
		return 0, 0, nil, false, nil
	}
	if err != nil {
		return 0, 0, nil, false, err
	}

	if lu.Valid {
		t := lu.Time
		return lat, lon, &t, true, nil
	}

	return lat, lon, nil, true, nil
}
