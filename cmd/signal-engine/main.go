package main

import (
	"net/http"
	"time"

	"github.com/NursultanKoshoev11/TrainingAgent/internal/analysis"
	"github.com/NursultanKoshoev11/TrainingAgent/internal/domain"
	"github.com/NursultanKoshoev11/TrainingAgent/internal/platform"
)

func main() {
	port := platform.GetEnv("PORT", "8083")
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", platform.Method(http.MethodGet, platform.HealthHandler("signal-engine")))
	mux.HandleFunc("/v1/signals", platform.Method(http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
		quote := r.URL.Query().Get("quote"); if quote == "" { quote = "USDT" }
		tickers := []domain.Ticker{{Symbol:"BTC"+quote,LastPrice:100,OpenPrice:95,HighPrice:105,LowPrice:94,PriceChangePercent:5.5,QuoteVolume:1000000,VolumeRankScore:1}}
		news := []domain.NewsArticle{{ID:"1",Title:"Bitcoin bullish rally",PublishedAt:time.Now().UTC(),Sentiment:0.4,Matched:true}}
		items := analysis.BuildSignals(tickers, news, quote, 20, analysis.ScoreConfig{MinimumProbability:0.58})
		platform.JSON(w, http.StatusOK, domain.SignalResponse{QuoteAsset:quote,Count:len(items),Signals:items})
	}))
	_ = platform.StartServer("signal-engine", port, mux)
}
