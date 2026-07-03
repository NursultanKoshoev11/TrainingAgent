package domain

type TechSnapshot struct {
	Symbol        string   `json:"symbol"`
	MarketType    string   `json:"market_type"`
	Intervals     []string `json:"intervals"`
	RSI           float64  `json:"rsi"`
	EMA20         float64  `json:"ema20"`
	EMA50         float64  `json:"ema50"`
	TrendScore    float64  `json:"trend_score"`
	MomentumScore float64  `json:"momentum_score"`
	VolumeScore   float64  `json:"volume_score"`
	Volatility    float64  `json:"volatility"`
	QualityScore  float64  `json:"quality_score"`
	Warnings      []string `json:"warnings"`
	Confirmations []string `json:"confirmations"`
}

type TechResponse struct {
	Symbol     string       `json:"symbol"`
	MarketType string       `json:"market_type"`
	Tech       TechSnapshot `json:"tech"`
}
