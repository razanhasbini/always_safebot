package db

import (
	"database/sql"
	"encoding/json"
)

type RuleAction struct {
	ID      int64
	Type    string
	Payload map[string]any
}

func GetRuleActions(db *sql.DB, ruleID int64) ([]RuleAction, error) {
	rows, err := db.Query(`
		SELECT id, action_type, action_payload
		FROM rule_actions
		WHERE rule_id = $1
		ORDER BY id ASC
	`, ruleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions []RuleAction
	for rows.Next() {
		var a RuleAction
		var raw []byte
		if err := rows.Scan(&a.ID, &a.Type, &raw); err != nil {
			return nil, err
		}
		a.Payload = map[string]any{}
		_ = json.Unmarshal(raw, &a.Payload)
		actions = append(actions, a)
	}
	return actions, nil
}
