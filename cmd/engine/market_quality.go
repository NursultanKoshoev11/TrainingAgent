package main

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/NursultanKoshoev11/TrainingAgent/internal/domain"
	"github.com/NursultanKoshoev11/TrainingAgent/internal/platform"
)

func loadQuality(tickers []domain.Ticker) map[string]domain.TechSnapshot {
	out := map[string]domain.TechSnapshot{}
	base := strings.TrimRight(platform.GetEnv("MARKET_SERVICE_URL", "http://localhost:8082"), "/")
	client := &http.Client{Timeout: 20 * time.Second}
	maxItems := platform.GetEnvInt("QUALITY_SYMBOL_LIMIT", 12)
	if maxItems <= 0 || maxItems > len(tickers) { maxItems = len(tickers) }
	for i := 0; i < maxItems; i++ {
		u, _ := url.Parse(base + "/v1/tech")
		q := u.Query()
		q.Set("symbol", tickers[i].Symbol)
		u.RawQuery = q.Encode()
		var response domain.TechResponse
		if err := platform.GetJSON(client, u.String(), &response); err == nil {
			out[tickers[i].Symbol] = response.Tech
		}
	}
	return out
}
