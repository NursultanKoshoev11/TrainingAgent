package domain

import "time"

type Signal struct {
	Symbol              string        `json:"symbol"`
	Action              string        `json:"action"`
	Probability         float64       `json:"probability"`
	ExpectedMovePercent float64       `json:"expected_move_percent"`
	RiskScore           float64       `json:"risk_score"`
	Confidence          float64       `json:"confidence"`
	GeneratedAt         time.Time     `json:"generated_at"`
	Reasons             []string      `json:"reasons"`
	Market              Ticker        `json:"market"`
	News                []NewsArticle `json:"news"`
}

type SignalResponse struct {
	QuoteAsset string   `json:"quote_asset"`
	Count      int      `json:"count"`
	Signals    []Signal `json:"signals"`
}
