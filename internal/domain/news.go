package domain

import "time"

type NewsArticle struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Link        string    `json:"link"`
	Source      string    `json:"source"`
	PublishedAt time.Time `json:"published_at"`
	Summary     string    `json:"summary"`
	Sentiment   float64   `json:"sentiment"`
	Matched     bool      `json:"matched"`
}

type NewsResponse struct {
	Query    string        `json:"query"`
	Count    int           `json:"count"`
	Articles []NewsArticle `json:"articles"`
}
