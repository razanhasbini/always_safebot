package bot

import (
	"database/sql"
	"telegram_chabot/internal/db"
)

type MatchResult struct {
	RuleID int64
	Reason string
}

func MatchRules(
	database *sql.DB,
	userID int64,
	message string,
) (*MatchResult, error) {

	normalized := NormalizeText(message)

	rules, err := db.GetActiveRules(database, userID)
	if err != nil {
		return nil, err
	}

	for _, rule := range rules {
		ok, reason := MatchTriggers(database, rule.ID, normalized)
		if ok {
			return &MatchResult{
				RuleID: rule.ID,
				Reason: reason,
			}, nil
		}
	}

	return nil, nil
}
