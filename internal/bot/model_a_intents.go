package bot

import "strings"

type ModelAIntent int

const (
	IntentNone ModelAIntent = iota
	IntentDrivingStatusRequest
	IntentConfirmShare
	IntentCancelRequest
)

// DetectModelAIntent detects high-level conversational intent
func DetectModelAIntent(normalized string) ModelAIntent {

	// Cancel intent
	if isCancelIntent(normalized) {
		return IntentCancelRequest
	}

	// Confirmation intent
	if isConfirmIntent(normalized) {
		return IntentConfirmShare
	}

	// Driving status intent (name-agnostic)
	if strings.Contains(normalized, "driving") {
		return IntentDrivingStatusRequest
	}

	return IntentNone
}

func isConfirmIntent(s string) bool {
	switch s {
	case "yes",
		"ok",
		"okay",
		"send it",
		"share location",
		"send location":
		return true
	}
	return false
}

func isCancelIntent(s string) bool {
	switch s {
	case "no",
		"cancel",
		"stop",
		"never mind",
		"nevermind":
		return true
	}
	return false
}
func ExtractName(normalized string) string {
	parts := strings.Fields(normalized)
	for _, p := range parts {
		if p != "is" && p != "driving" && p != "are" && p != "you" {
			return p
		}
	}
	return ""
}
