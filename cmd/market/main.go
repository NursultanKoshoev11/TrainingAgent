package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/NursultanKoshoev11/TrainingAgent/internal/domain"
	"github.com/NursultanKoshoev11/TrainingAgent/internal/platform"
)

type remoteTicker struct {
	Symbol string `json:"symbol"`
	LastPrice string `json:"lastPrice"`
	OpenPrice string `json:"openPrice"`
	HighPrice string `json:"highPrice"`
	LowPrice string `json:"lowPrice"`
	PriceChange string `json:"priceChange"`
	PriceChangePercent string `json:"priceChangePercent"`
	Volume string `json:"volume"`
	QuoteVolume string `json:"quoteVolume"`
}

func main() {
	port := platform.GetEnv("PORT", "8082")
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", platform.Method(http.MethodGet, platform.HealthHandler("market")))
	mux.HandleFunc("/v1/tickers", platform.Method(http.MethodGet, handle))
	_ = platform.StartServer("market", port, mux)
}

func handle(w http.ResponseWriter, r *http.Request) {
	quote := strings.ToUpper(r.URL.Query().Get("quote"))
	if quote == "" { quote = "USDT" }
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 { limit = 20 }
	items, err := load(quote, limit)
	if err != nil { items = fallback(quote) }
	platform.JSON(w, http.StatusOK, domain.TickerResponse{QuoteAsset: quote, Count: len(items), Tickers: items})
}

func load(quote string, limit int) ([]domain.Ticker, error) {
	base := strings.TrimRight(platform.GetEnv("BINANCE_API_BASE", "https://api.binance.com"), "/")
	client := &http.Client{Timeout: platform.HTTPTimeoutFromEnv("BINANCE_HTTP_TIMEOUT_SECONDS", 10)}
	resp, err := client.Get(base + "/api/v3/ticker/24hr")
	if err != nil { return nil, err }
	defer resp.Body.Close()
	var raw []remoteTicker
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil { return nil, err }
	out := make([]domain.Ticker, 0, limit)
	for _, x := range raw {
		if !strings.HasSuffix(x.Symbol, quote) { continue }
		out = append(out, domain.Ticker{Symbol:x.Symbol,LastPrice:num(x.LastPrice),OpenPrice:num(x.OpenPrice),HighPrice:num(x.HighPrice),LowPrice:num(x.LowPrice),PriceChange:num(x.PriceChange),PriceChangePercent:num(x.PriceChangePercent),Volume:num(x.Volume),QuoteVolume:num(x.QuoteVolume)})
	}
	sort.Slice(out, func(i,j int) bool { return out[i].QuoteVolume > out[j].QuoteVolume })
	for i := range out { out[i].VolumeRankScore = 1 - float64(i)/float64(max(1,len(out))) }
	if len(out) > limit { out = out[:limit] }
	return out, nil
}

func fallback(quote string) []domain.Ticker { return []domain.Ticker{{Symbol:"BTC"+quote,LastPrice:100,OpenPrice:95,HighPrice:105,LowPrice:94,PriceChangePercent:5.5,QuoteVolume:1000000,VolumeRankScore:1},{Symbol:"ETH"+quote,LastPrice:50,OpenPrice:51,HighPrice:54,LowPrice:48,PriceChangePercent:-1.2,QuoteVolume:700000,VolumeRankScore:.7}} }
func num(s string) float64 { v, _ := strconv.ParseFloat(s, 64); return v }
func max(a,b int) int { if a>b { return a }; return b }
