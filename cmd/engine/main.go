package main

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/NursultanKoshoev11/TrainingAgent/internal/analysis"
	"github.com/NursultanKoshoev11/TrainingAgent/internal/domain"
	"github.com/NursultanKoshoev11/TrainingAgent/internal/platform"
	"github.com/NursultanKoshoev11/TrainingAgent/internal/storage"
)

var store *storage.Store

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	opened, err := storage.Open(ctx, platform.GetEnv("DATABASE_URL", ""))
	if err == nil { store = opened; defer store.Close() }
	port := platform.GetEnv("PORT", "8083")
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", platform.Method(http.MethodGet, platform.HealthHandler("engine")))
	mux.HandleFunc("/v1/signals", platform.Method(http.MethodGet, handleSignals))
	mux.HandleFunc("/v1/signals/history", platform.Method(http.MethodGet, handleHistory))
	_ = platform.StartServer("engine", port, mux)
}

func handleSignals(w http.ResponseWriter, r *http.Request) {
	quote := strings.ToUpper(r.URL.Query().Get("quote")); if quote == "" { quote = "USDT" }
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit")); if limit <= 0 { limit = 20 }
	query := r.URL.Query().Get("query"); if query == "" { query = "crypto market" }
	tickers := getTickers(quote, limit)
	news := getNews(query, limit)
	quality := loadQuality(tickers)
	items := analysis.BuildGuardedSignals(tickers, news, quality, quote, limit, analysis.ScoreConfig{MinimumProbability: platform.GetEnvFloat("SIGNAL_MIN_PROBABILITY", 0.58)})
	if store != nil { ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second); _ = store.SaveSignals(ctx, items); cancel() }
	platform.JSON(w, http.StatusOK, domain.SignalResponse{QuoteAsset: quote, Count: len(items), Signals: items})
}

func handleHistory(w http.ResponseWriter, r *http.Request) {
	if store == nil || !store.Enabled() { platform.JSON(w, http.StatusOK, map[string]any{"count":0,"items":[]storage.SignalHistoryItem{},"storage_enabled":false}); return }
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit")); if limit <= 0 { limit = 100 }
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second); defer cancel()
	items, err := store.SignalHistory(ctx, strings.ToUpper(r.URL.Query().Get("symbol")), r.URL.Query().Get("action"), limit)
	if err != nil { platform.Fail(w, http.StatusBadGateway, err.Error()); return }
	platform.JSON(w, http.StatusOK, map[string]any{"count":len(items),"items":items,"storage_enabled":true})
}

func getTickers(quote string, limit int) []domain.Ticker {
	base := strings.TrimRight(platform.GetEnv("MARKET_SERVICE_URL", "http://localhost:8082"), "/")
	u, _ := url.Parse(base + "/v1/tickers")
	q := u.Query(); q.Set("quote", quote); q.Set("limit", strconv.Itoa(limit)); u.RawQuery = q.Encode()
	var response domain.TickerResponse
	client := &http.Client{Timeout: 15*time.Second}
	if err := platform.GetJSON(client, u.String(), &response); err != nil { return nil }
	return response.Tickers
}

func getNews(query string, limit int) []domain.NewsArticle {
	base := strings.TrimRight(platform.GetEnv("NEWS_SERVICE_URL", "http://localhost:8081"), "/")
	u, _ := url.Parse(base + "/v1/news")
	q := u.Query(); q.Set("query", query); q.Set("limit", strconv.Itoa(limit)); u.RawQuery = q.Encode()
	var response domain.NewsResponse
	client := &http.Client{Timeout: 15*time.Second}
	if err := platform.GetJSON(client, u.String(), &response); err != nil { return nil }
	return response.Articles
}
