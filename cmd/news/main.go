package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/NursultanKoshoev11/TrainingAgent/internal/analysis"
	"github.com/NursultanKoshoev11/TrainingAgent/internal/domain"
	"github.com/NursultanKoshoev11/TrainingAgent/internal/platform"
)

func main() {
	port := platform.GetEnv("PORT", "8081")
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", platform.Method(http.MethodGet, platform.HealthHandler("news")))
	mux.HandleFunc("/v1/news", platform.Method(http.MethodGet, handle))
	_ = platform.StartServer("news", port, mux)
}

func handle(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" { query = "crypto bitcoin ethereum binance" }
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 { limit = 20 }
	items := []domain.NewsArticle{{ID:"1",Title:"Bitcoin bullish rally with inflows",Source:"demo",PublishedAt:time.Now().UTC(),Summary:"market research"},{ID:"2",Title:"Exchange hack risk warning",Source:"demo",PublishedAt:time.Now().UTC(),Summary:"risk research"}}
	for i := range items { items[i].Sentiment = analysis.SentimentScore(items[i].Title+" "+items[i].Summary); items[i].Matched = analysis.QueryMatches(items[i].Title+" "+items[i].Summary, query) }
	if len(items) > limit { items = items[:limit] }
	platform.JSON(w, http.StatusOK, domain.NewsResponse{Query: query, Count: len(items), Articles: items})
}
