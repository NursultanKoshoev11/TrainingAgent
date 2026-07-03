package domain

import "time"

type Candle struct {
	Symbol    string    `json:"symbol"`
	Interval  string    `json:"interval"`
	OpenTime  time.Time `json:"open_time"`
	CloseTime time.Time `json:"close_time"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    float64   `json:"volume"`
}

type CandleResponse struct {
	Symbol   string   `json:"symbol"`
	Interval string   `json:"interval"`
	Count    int      `json:"count"`
	Candles  []Candle `json:"candles"`
}
