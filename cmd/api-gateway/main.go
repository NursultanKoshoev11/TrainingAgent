package main

import (
	_ "embed"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/NursultanKoshoev11/TrainingAgent/internal/platform"
)

//go:embed web/dashboard_ru.html
var dashboardHTML string

var signalCache = struct {
	sync.Mutex
	query string
	body []byte
	status int
	updatedAt time.Time
}{status: http.StatusOK}

func main() {
	port := platform.GetEnv("PORT", "8080")
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", platform.Method(http.MethodGet, platform.HealthHandler("api-gateway")))
	mux.HandleFunc("/v1/status", platform.Method(http.MethodGet, requireAuth(statusHandler)))
	mux.HandleFunc("/v1/signals", platform.Method(http.MethodGet, requireAuth(proxySignals)))
	mux.HandleFunc("/v1/signals/history", platform.Method(http.MethodGet, requireAuth(proxyHistory)))
	mux.HandleFunc("/v1/candles", platform.Method(http.MethodGet, requireAuth(proxyCandles)))
	mux.HandleFunc("/", platform.Method(http.MethodGet, requireAuth(dashboard)))
	_ = platform.StartServer("api-gateway", port, mux)
}

func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	user := platform.GetEnv("DASHBOARD_USER", "")
	pass := platform.GetEnv("DASHBOARD_PASSWORD", "")
	if user == "" || pass == "" { return next }
	return func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != user || p != pass {
			w.Header().Set("WWW-Authenticate", `Basic realm="TrainingAgent"`)
			platform.Fail(w, http.StatusUnauthorized, "authentication required")
			return
		}
		next(w, r)
	}
}

func proxySignals(w http.ResponseWriter, r *http.Request) {
	cacheSeconds := platform.GetEnvInt("SIGNAL_CACHE_SECONDS", 60)
	cacheKey := r.URL.RawQuery
	if body, status, ok := readSignalCache(cacheKey, cacheSeconds); ok {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-TrainingAgent-Cache", "hit")
		w.WriteHeader(status)
		_, _ = w.Write(body)
		return
	}
	proxyWithCache(w, r, platform.GetEnv("ENGINE_SERVICE_URL", "http://localhost:8083"), "/v1/signals", true, cacheKey)
}

func proxyHistory(w http.ResponseWriter, r *http.Request) {
	proxyWithCache(w, r, platform.GetEnv("ENGINE_SERVICE_URL", "http://localhost:8083"), "/v1/signals/history", false, "")
}

func proxyCandles(w http.ResponseWriter, r *http.Request) {
	proxyWithCache(w, r, platform.GetEnv("MARKET_SERVICE_URL", "http://localhost:8082"), "/v1/candles", false, "")
}

func proxyWithCache(w http.ResponseWriter, r *http.Request, baseURL, path string, cache bool, cacheKey string) {
	base := strings.TrimRight(baseURL, "/")
	upstream := base + path
	if r.URL.RawQuery != "" { upstream += "?" + r.URL.RawQuery }
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Get(upstream)
	if err != nil { platform.Fail(w, http.StatusBadGateway, err.Error()); return }
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4_000_000))
	if err != nil { platform.Fail(w, http.StatusBadGateway, err.Error()); return }
	if cache { writeSignalCache(cacheKey, body, resp.StatusCode) }
	w.Header().Set("Content-Type", "application/json")
	if cache { w.Header().Set("X-TrainingAgent-Cache", "miss") }
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(body)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{Timeout: 5 * time.Second}
	engineOK := ping(client, platform.GetEnv("ENGINE_SERVICE_URL", "http://localhost:8083"))
	marketOK := ping(client, platform.GetEnv("MARKET_SERVICE_URL", "http://localhost:8082"))
	platform.JSON(w, http.StatusOK, map[string]any{
		"service": "api-gateway",
		"status": "ok",
		"engine_ok": engineOK,
		"market_ok": marketOK,
		"auth_enabled": platform.GetEnv("DASHBOARD_USER", "") != "" && platform.GetEnv("DASHBOARD_PASSWORD", "") != "",
		"cache_seconds": platform.GetEnvInt("SIGNAL_CACHE_SECONDS", 60),
		"time": time.Now().UTC().Format(time.RFC3339),
	})
}

func ping(client *http.Client, baseURL string) bool {
	resp, err := client.Get(strings.TrimRight(baseURL, "/") + "/healthz")
	ok := err == nil && resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300
	if resp != nil && resp.Body != nil { _ = resp.Body.Close() }
	return ok
}

func readSignalCache(query string, ttlSeconds int) ([]byte, int, bool) {
	if ttlSeconds <= 0 { return nil, 0, false }
	signalCache.Lock()
	defer signalCache.Unlock()
	if signalCache.query != query || len(signalCache.body) == 0 { return nil, 0, false }
	if time.Since(signalCache.updatedAt) > time.Duration(ttlSeconds)*time.Second { return nil, 0, false }
	return append([]byte(nil), signalCache.body...), signalCache.status, true
}

func writeSignalCache(query string, body []byte, status int) {
	if status < 200 || status >= 300 { return }
	signalCache.Lock()
	defer signalCache.Unlock()
	signalCache.query = query
	signalCache.body = append([]byte(nil), body...)
	signalCache.status = status
	signalCache.updatedAt = time.Now()
}

func dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.WriteString(w, dashboardHTML)
}
