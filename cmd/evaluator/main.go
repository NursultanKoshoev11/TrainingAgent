package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/NursultanKoshoev11/TrainingAgent/internal/domain"
	"github.com/NursultanKoshoev11/TrainingAgent/internal/platform"
	"github.com/NursultanKoshoev11/TrainingAgent/internal/storage"
)

type app struct { store *storage.Store }

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	store, err := storage.Open(ctx, platform.GetEnv("DATABASE_URL", ""))
	if err != nil { panic(err) }
	defer store.Close()
	a := &app{store: store}
	go a.loop()
	port := platform.GetEnv("PORT", "8084")
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", platform.Method(http.MethodGet, platform.HealthHandler("evaluator")))
	mux.HandleFunc("/v1/evaluations/run", platform.Method(http.MethodPost, a.runHandler))
	mux.HandleFunc("/v1/evaluations", platform.Method(http.MethodGet, a.listHandler))
	_ = platform.StartServer("evaluator", port, mux)
}

func (a *app) loop() {
	interval := time.Duration(platform.GetEnvInt("EVALUATOR_CHECK_INTERVAL_SECONDS", 60)) * time.Second
	if interval < 10*time.Second { interval = 10*time.Second }
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for { a.evaluateOnce(context.Background()); <-ticker.C }
}

func (a *app) runHandler(w http.ResponseWriter, r *http.Request) {
	count, err := a.evaluateOnce(r.Context())
	if err != nil { platform.Fail(w, http.StatusBadGateway, err.Error()); return }
	platform.JSON(w, http.StatusOK, map[string]any{"evaluated": count})
}

func (a *app) listHandler(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit")); if limit <= 0 { limit = 100 }
	since, _ := strconv.Atoi(r.URL.Query().Get("since_minutes"))
	symbol := strings.ToUpper(r.URL.Query().Get("symbol"))
	action := r.URL.Query().Get("action")
	passedFilter := r.URL.Query().Get("passed")
	items, err := a.store.Evaluations(r.Context(), symbol, action, passedFilter, since, limit)
	if err != nil { platform.Fail(w, http.StatusBadGateway, err.Error()); return }
	passed := 0
	for _, item := range items { if item.Passed { passed++ } }
	accuracy := 0.0
	if len(items) > 0 { accuracy = float64(passed) / float64(len(items)) }
	platform.JSON(w, http.StatusOK, map[string]any{"count": len(items), "passed": passed, "accuracy": accuracy, "filters": map[string]any{"symbol": symbol, "action": action, "passed": passedFilter, "since_minutes": since}, "items": items})
}

func (a *app) evaluateOnce(ctx context.Context) (int, error) {
	if a.store == nil || !a.store.Enabled() { return 0, nil }
	horizon := platform.GetEnvInt("EVALUATION_HORIZON_MINUTES", 20)
	pending, err := a.store.PendingEvaluations(ctx, horizon, platform.GetEnvInt("EVALUATOR_BATCH_SIZE", 200))
	if err != nil { return 0, err }
	count := 0
	for _, item := range pending {
		price, err := currentPrice(item.Symbol)
		if err != nil || price <= 0 || item.EntryPrice <= 0 { continue }
		result := evaluate(item, price, horizon)
		if err := a.store.SaveEvaluation(ctx, result); err != nil { return count, err }
		count++
	}
	return count, nil
}

func currentPrice(symbol string) (float64, error) {
	base := strings.TrimRight(platform.GetEnv("MARKET_SERVICE_URL", "http://localhost:8082"), "/")
	u, _ := url.Parse(base + "/v1/tickers")
	quote := "USDT"
	for _, q := range []string{"USDT","BTC","ETH"} { if strings.HasSuffix(symbol, q) { quote = q; break } }
	q := u.Query(); q.Set("quote", quote); q.Set("limit", "500"); u.RawQuery = q.Encode()
	var response domain.TickerResponse
	client := &http.Client{Timeout: 15*time.Second}
	if err := platform.GetJSON(client, u.String(), &response); err != nil { return 0, err }
	for _, ticker := range response.Tickers { if ticker.Symbol == symbol { return ticker.LastPrice, nil } }
	return 0, fmt.Errorf("symbol not found: %s", symbol)
}

func evaluate(item storage.PendingEvaluation, checkedPrice float64, horizon int) storage.SignalEvaluation {
	move := ((checkedPrice - item.EntryPrice) / item.EntryPrice) * 100
	expected := expectedDirection(item.Action)
	actual := actualDirection(move)
	passed := false
	reason := ""
	switch item.Action {
	case "BUY_WATCH": passed = move > 0; reason = fmt.Sprintf("Ожидали рост. Через %d минут движение: %.2f%%.", horizon, move)
	case "SELL_WATCH": passed = move < 0; reason = fmt.Sprintf("Ожидали снижение/слабость. Через %d минут движение: %.2f%%.", horizon, move)
	case "HOLD_WATCH": passed = move > -0.5 && move < 0.5; reason = fmt.Sprintf("Ожидали боковое движение. Через %d минут движение: %.2f%%.", horizon, move)
	case "AVOID_WATCH": passed = abs(move) > 1.0; reason = fmt.Sprintf("Ожидали повышенный риск. Через %d минут движение: %.2f%%.", horizon, move)
	default: reason = fmt.Sprintf("Нейтральная проверка. Через %d минут движение: %.2f%%.", horizon, move)
	}
	return storage.SignalEvaluation{SignalID:item.SignalID, Symbol:item.Symbol, Action:item.Action, HorizonMinutes:horizon, EntryPrice:item.EntryPrice, CheckedPrice:checkedPrice, MovePercent:move, ExpectedDirection:expected, ActualDirection:actual, Passed:passed, Reason:reason, SignalCreatedAt:item.CreatedAt}
}

func expectedDirection(action string) string { switch action { case "BUY_WATCH": return "up"; case "SELL_WATCH": return "down"; case "HOLD_WATCH": return "flat"; case "AVOID_WATCH": return "volatile"; default: return "unknown" } }
func actualDirection(move float64) string { if move > 0.2 { return "up" }; if move < -0.2 { return "down" }; return "flat" }
func abs(x float64) float64 { if x < 0 { return -x }; return x }
