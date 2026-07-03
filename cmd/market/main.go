package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/NursultanKoshoev11/TrainingAgent/internal/analysis"
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
	mux.HandleFunc("/v1/tickers", platform.Method(http.MethodGet, handleTickers))
	mux.HandleFunc("/v1/candles", platform.Method(http.MethodGet, handleCandles))
	mux.HandleFunc("/v1/tech", platform.Method(http.MethodGet, handleTech))
	_ = platform.StartServer("market", port, mux)
}

func handleTickers(w http.ResponseWriter, r *http.Request) {
	quote := strings.ToUpper(r.URL.Query().Get("quote"))
	if quote == "" { quote = "USDT" }
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 { limit = 20 }
	items, err := load(quote, limit)
	if err != nil { items = fallback(quote) }
	platform.JSON(w, http.StatusOK, domain.TickerResponse{QuoteAsset: quote, Count: len(items), Tickers: items})
}

func handleCandles(w http.ResponseWriter, r *http.Request) {
	symbol := strings.ToUpper(r.URL.Query().Get("symbol"))
	if symbol == "" { symbol = "BTCUSDT" }
	interval := r.URL.Query().Get("interval")
	if interval == "" { interval = "5m" }
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 1000 { limit = 100 }
	candles, err := loadCandles(symbol, interval, limit)
	if err != nil { platform.Fail(w, http.StatusBadGateway, err.Error()); return }
	platform.JSON(w, http.StatusOK, domain.CandleResponse{Symbol: symbol, Interval: interval, Count: len(candles), Candles: candles})
}

func handleTech(w http.ResponseWriter, r *http.Request) {
	symbol := strings.ToUpper(r.URL.Query().Get("symbol"))
	if symbol == "" { symbol = "BTCUSDT" }
	tech, err := buildTech(symbol)
	if err != nil { platform.Fail(w, http.StatusBadGateway, err.Error()); return }
	platform.JSON(w, http.StatusOK, domain.TechResponse{Symbol: symbol, MarketType: "spot", Tech: tech})
}

func buildTech(symbol string) (domain.TechSnapshot, error) {
	intervals := []string{"5m", "15m", "1h", "4h"}
	out := domain.TechSnapshot{Symbol: symbol, MarketType: "spot", Intervals: intervals}
	var trendTotal, momentumTotal, volumeTotal, volTotal float64
	var confirmations, warnings []string
	for _, interval := range intervals {
		candles, err := loadCandles(symbol, interval, 80)
		if err != nil { warnings = append(warnings, "нет свечей для "+interval); continue }
		closes := analysis.CloseValues(candles)
		ema20 := analysis.EMA(closes, 20)
		ema50 := analysis.EMA(closes, 50)
		rsi := analysis.RSI(closes, 14)
		trend := analysis.TrendScoreFromEMA(ema20, ema50, closes[len(closes)-1])
		mom := analysis.MomentumScore(candles)
		volSpike := analysis.VolumeSpikeScore(candles)
		vol := analysis.VolatilityPercent(candles)
		trendTotal += trend; momentumTotal += mom; volumeTotal += volSpike; volTotal += vol
		if interval == "1h" { out.RSI = rsi; out.EMA20 = ema20; out.EMA50 = ema50 }
		if trend > 0.25 { confirmations = append(confirmations, interval+": тренд выше EMA") }
		if trend < -0.25 { warnings = append(warnings, interval+": тренд слабый") }
		if rsi > 75 { warnings = append(warnings, interval+": RSI перегрет") }
		if rsi < 30 { warnings = append(warnings, interval+": RSI перепродан") }
		if vol > 3.5 { warnings = append(warnings, interval+": высокая волатильность") }
	}
	checked := float64(len(intervals))
	out.TrendScore = analysis.Clamp(trendTotal/checked, -1, 1)
	out.MomentumScore = analysis.Clamp(momentumTotal/checked, -1, 1)
	out.VolumeScore = analysis.Clamp(volumeTotal/checked, 0, 1)
	out.Volatility = volTotal/checked
	out.QualityScore = analysis.Clamp(0.45+out.TrendScore*0.25+out.MomentumScore*0.15+out.VolumeScore*0.10-analysis.Clamp(out.Volatility/10,0,0.25), 0, 1)
	out.Confirmations = confirmations
	out.Warnings = warnings
	return out, nil
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

func loadCandles(symbol, interval string, limit int) ([]domain.Candle, error) {
	base := strings.TrimRight(platform.GetEnv("BINANCE_API_BASE", "https://api.binance.com"), "/")
	u, _ := url.Parse(base + "/api/v3/klines")
	q := u.Query(); q.Set("symbol", symbol); q.Set("interval", interval); q.Set("limit", strconv.Itoa(limit)); u.RawQuery = q.Encode()
	client := &http.Client{Timeout: platform.HTTPTimeoutFromEnv("BINANCE_HTTP_TIMEOUT_SECONDS", 10)}
	resp, err := client.Get(u.String())
	if err != nil { return nil, err }
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 { return nil, fmt.Errorf("candles request failed: %s", resp.Status) }
	var raw [][]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil { return nil, err }
	out := make([]domain.Candle, 0, len(raw))
	for _, row := range raw { if len(row) >= 7 { out = append(out, domain.Candle{Symbol:symbol,Interval:interval,OpenTime:msTime(row[0]),Open:numAny(row[1]),High:numAny(row[2]),Low:numAny(row[3]),Close:numAny(row[4]),Volume:numAny(row[5]),CloseTime:msTime(row[6])}) } }
	return out, nil
}

func fallback(quote string) []domain.Ticker { return []domain.Ticker{{Symbol:"BTC"+quote,LastPrice:100,OpenPrice:95,HighPrice:105,LowPrice:94,PriceChangePercent:5.5,QuoteVolume:1000000,VolumeRankScore:1},{Symbol:"ETH"+quote,LastPrice:50,OpenPrice:51,HighPrice:54,LowPrice:48,PriceChangePercent:-1.2,QuoteVolume:700000,VolumeRankScore:.7}} }
func num(s string) float64 { v, _ := strconv.ParseFloat(s, 64); return v }
func max(a,b int) int { if a>b { return a }; return b }
func numAny(v any) float64 { switch x := v.(type) { case string: return num(x); case float64: return x; default: return 0 } }
func msTime(v any) time.Time { return time.UnixMilli(int64(numAny(v))).UTC() }
