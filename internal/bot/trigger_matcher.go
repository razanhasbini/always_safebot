package bot

import "database/sql"

func MatchTriggers(db *sql.DB, ruleID int64, normalizedMessage string) (bool, string) {
	// 1) exact
	var exact bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM rule_triggers
			WHERE rule_id = $1 AND trigger_type = 'exact' AND trigger_value = $2
		)
	`, ruleID, normalizedMessage).Scan(&exact)
	if err == nil && exact {
		return true, "exact match"
	}

	// 2) keyword (trigger_value stored as keyword)
	var keyword bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM rule_triggers
			WHERE rule_id = $1 AND trigger_type = 'keyword'
			  AND $2 ILIKE '%' || trigger_value || '%'
		)
	`, ruleID, normalizedMessage).Scan(&keyword)
	if err == nil && keyword {
		return true, "keyword match"
	}

	// 3) fuzzy (pg_trgm)
	var trig string
	var score float64
	err = db.QueryRow(`
		SELECT trigger_value, similarity(trigger_value, $1) AS score
		FROM rule_triggers
		WHERE rule_id = $2
		  AND trigger_type = 'fuzzy'
		  AND similarity(trigger_value, $1) > COALESCE(fuzzy_threshold, 0.4)
		ORDER BY score DESC
		LIMIT 1
	`, normalizedMessage, ruleID).Scan(&trig, &score)

	if err == nil {
		return true, "fuzzy match"
	}

	return false, ""
}
