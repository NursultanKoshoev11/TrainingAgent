package main

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/NursultanKoshoev11/TrainingAgent/internal/analysis"
	"github.com/NursultanKoshoev11/TrainingAgent/internal/domain"
	"github.com/NursultanKoshoev11/TrainingAgent/internal/platform"
)

func main() {
	port := platform.GetEnv("PORT", "8083")
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", platform.Method(http.MethodGet, platform.HealthHandler("engine")))
	mux.HandleFunc("/v1/signals", platform.Method(http.MethodGet, handle))
	_ = platform.StartServer("engine", port, mux)
}

func handle(w http.ResponseWriter, r *http.Request) {
	quote := strings.ToUpper(r.URL.Query().Get("quote"))
	if quote == "" { quote = "USDT" }
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 { limit = 20 }
	query := r.URL.Query().Get("query")
	if query == "" { query = "crypto bitcoin ethereum binance" }
	items := analysis.BuildSignals(getTickers(quote, limit), getNews(query, limit), quote, limit, analysis.ScoreConfig{MinimumProbability: platform.GetEnvFloat("SIGNAL_MIN_PROBABILITY", 0.58)})
	platform.JSON(w, http.StatusOK, domain.SignalResponse{QuoteAsset: quote, Count: len(items), Signals: items})
}

func getTickers(quote string, limit int) []domain.Ticker {
	base := strings.TrimRight(platform.GetEnv("MARKET_SERVICE_URL", "http://localhost:8082"), "/")
	u, _ := url.Parse(base + "/v1/tickers")
	q := u.Query(); q.Set("quote", quote); q.Set("limit", strconv.Itoa(limit)); u.RawQuery = q.Encode()
	var response domain.TickerResponse
	client := &http.Client{Timeout: 15 * time.Second}
	if err := platform.GetJSON(client, u.String(), &response); err != nil {
		return []domain.Ticker{{Symbol:"BTC"+quote,LastPrice:100,OpenPrice:95,HighPrice:105,LowPrice:94,PriceChangePercent:5.5,QuoteVolume:1000000,VolumeRankScore:1}}
	}
	return response.Tickers
}

func getNews(query string, limit int) []domain.NewsArticle {
	base := strings.TrimRight(platform.GetEnv("NEWS_SERVICE_URL", "http://localhost:8081"), "/")
	u, _ := url.Parse(base + "/v1/news")
	q := u.Query(); q.Set("query", query); q.Set("limit", strconv.Itoa(limit)); u.RawQuery = q.Encode()
	var response domain.NewsResponse
	client := &http.Client{Timeout: 15 * time.Second}
	if err := platform.GetJSON(client, u.String(), &response); err != nil { return nil }
	return response.Articles
}
