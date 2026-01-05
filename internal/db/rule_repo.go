package db

import (
	"database/sql"
)

type Rule struct {
	ID              int64
	ApprovalMode    string
	RequiredContext sql.NullString
	Priority        int
}

func GetActiveRules(db *sql.DB, userID int64) ([]Rule, error) {
	rows, err := db.Query(`
		SELECT id, approval_mode, required_context, priority
		FROM rules
		WHERE user_id = $1 AND is_enabled = true
		ORDER BY priority ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []Rule
	for rows.Next() {
		var r Rule
		if err := rows.Scan(&r.ID, &r.ApprovalMode, &r.RequiredContext, &r.Priority); err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}
	return rules, nil
}
