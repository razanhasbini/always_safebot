package bot

import (
	"strings"
	"unicode"
)

func NormalizeText(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))

	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
