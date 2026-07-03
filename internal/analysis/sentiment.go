package analysis

import (
	"strings"
	"unicode"
)

func SentimentScore(text string) float64 {
	words := tokenize(strings.ToLower(text))
	var score float64
	for _, w := range words {
		switch w {
		case "bullish", "rally", "surge", "growth", "inflow", "inflows", "adoption", "approved", "breakout":
			score += 0.25
		case "bearish", "hack", "exploit", "selloff", "crackdown", "ban", "lawsuit", "risk", "warning":
			score -= 0.25
		}
	}
	return Clamp(score, -1, 1)
}

func QueryMatches(text, query string) bool {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return true
	}
	text = strings.ToLower(text)
	for _, token := range tokenize(query) {
		if len(token) >= 3 && strings.Contains(text, token) {
			return true
		}
	}
	return false
}

func tokenize(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}
