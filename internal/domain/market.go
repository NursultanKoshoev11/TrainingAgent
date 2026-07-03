package domain

type Ticker struct {
	Symbol             string  `json:"symbol"`
	LastPrice          float64 `json:"last_price"`
	OpenPrice          float64 `json:"open_price"`
	HighPrice          float64 `json:"high_price"`
	LowPrice           float64 `json:"low_price"`
	PriceChange        float64 `json:"price_change"`
	PriceChangePercent float64 `json:"price_change_percent"`
	Volume             float64 `json:"volume"`
	QuoteVolume        float64 `json:"quote_volume"`
	VolumeRankScore    float64 `json:"volume_rank_score"`
}

type TickerResponse struct {
	QuoteAsset string   `json:"quote_asset"`
	Count      int      `json:"count"`
	Tickers    []Ticker `json:"tickers"`
}
