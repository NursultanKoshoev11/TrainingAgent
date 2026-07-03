package main

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/NursultanKoshoev11/TrainingAgent/internal/platform"
)

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
	mux.HandleFunc("/v1/status", platform.Method(http.MethodGet, statusHandler))
	mux.HandleFunc("/v1/signals", platform.Method(http.MethodGet, proxySignals))
	mux.HandleFunc("/", platform.Method(http.MethodGet, dashboard))
	_ = platform.StartServer("api-gateway", port, mux)
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
	base := strings.TrimRight(platform.GetEnv("ENGINE_SERVICE_URL", "http://localhost:8083"), "/")
	upstream := base + "/v1/signals"
	if r.URL.RawQuery != "" { upstream += "?" + r.URL.RawQuery }
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Get(upstream)
	if err != nil { platform.Fail(w, http.StatusBadGateway, err.Error()); return }
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4_000_000))
	if err != nil { platform.Fail(w, http.StatusBadGateway, err.Error()); return }
	writeSignalCache(cacheKey, body, resp.StatusCode)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-TrainingAgent-Cache", "miss")
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(body)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	engineURL := strings.TrimRight(platform.GetEnv("ENGINE_SERVICE_URL", "http://localhost:8083"), "/") + "/healthz"
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(engineURL)
	engineOK := err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300
	if resp != nil && resp.Body != nil { _ = resp.Body.Close() }
	platform.JSON(w, http.StatusOK, map[string]any{
		"service": "api-gateway",
		"status": "ok",
		"engine_ok": engineOK,
		"cache_seconds": platform.GetEnvInt("SIGNAL_CACHE_SECONDS", 60),
		"time": time.Now().UTC().Format(time.RFC3339),
	})
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
	_, _ = io.WriteString(w, page)
}

const page = `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>TrainingAgent Dashboard</title><style>body{font-family:Arial;background:#0b1020;color:#edf3ff;margin:0}.wrap{max-width:1200px;margin:auto;padding:24px}.card{background:#121a2f;border:1px solid #24304b;border-radius:16px;padding:16px;margin:12px 0}.grid{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:12px}.metric{background:#0f172a;border:1px solid #24304b;border-radius:14px;padding:12px}button,select,input{padding:10px;border-radius:10px;border:1px solid #33415f;background:#17213a;color:#fff}table{width:100%;border-collapse:collapse}td,th{padding:10px;border-bottom:1px solid #24304b;text-align:left}.BUY_WATCH{color:#4ade80}.SELL_WATCH{color:#fb7185}.HOLD_WATCH{color:#facc15}.AVOID_WATCH{color:#f97316}.ok{color:#4ade80}.bad{color:#fb7185}pre{white-space:pre-wrap}</style></head><body><div class="wrap"><h1>TrainingAgent Crypto Research Dashboard</h1><p>Research/watch signals only. Not financial advice and not auto-trading.</p><div class="card"><select id="quote"><option>USDT</option><option>BTC</option><option>ETH</option></select> <input id="limit" value="20" type="number" min="1" max="100"> <select id="filter"><option value="">All labels</option><option>BUY_WATCH</option><option>SELL_WATCH</option><option>HOLD_WATCH</option><option>AVOID_WATCH</option></select> <select id="interval"><option value="300000">5 min</option><option value="600000">10 min</option><option value="60000">1 min test</option></select> <button onclick="loadNow()">Refresh</button> <span id="status"></span></div><div class="grid"><div class="metric">Backend: <b id="backend">checking</b></div><div class="metric">Signals: <b id="mTotal">0</b></div><div class="metric">High risk: <b id="mRisk">0</b></div><div class="metric">Updated: <b id="mUpdated">never</b></div></div><div class="card"><table><thead><tr><th>Symbol</th><th>Label</th><th>Score</th><th>Expected move</th><th>Risk</th><th>Confidence</th><th>24h</th></tr></thead><tbody id="rows"></tbody></table></div><div class="card"><h2>Details</h2><pre id="details">Select a row</pre></div></div><script>let timer,last=[];async function loadStatus(){try{let r=await fetch('/v1/status',{cache:'no-store'});let d=await r.json();backend.textContent=d.engine_ok?'OK':'ENGINE DOWN';backend.className=d.engine_ok?'ok':'bad'}catch(e){backend.textContent='DOWN';backend.className='bad'}}function render(){let f=filter.value;let data=f?last.filter(x=>x.action===f):last;rows.innerHTML='';let high=0;data.forEach(s=>{if(s.risk_score>=0.7)high++;let tr=document.createElement('tr');tr.innerHTML='<td>'+s.symbol+'</td><td class="'+s.action+'">'+s.action+'</td><td>'+Math.round(s.probability*100)+'%</td><td>'+s.expected_move_percent+'%</td><td>'+Math.round(s.risk_score*100)+'%</td><td>'+Math.round(s.confidence*100)+'%</td><td>'+s.market.price_change_percent+'%</td>';tr.onclick=()=>details.textContent=JSON.stringify(s,null,2);rows.appendChild(tr)});mTotal.textContent=data.length;mRisk.textContent=high;mUpdated.textContent=new Date().toLocaleTimeString()}async function loadNow(){status.textContent='loading...';await loadStatus();let q=quote.value,l=limit.value;try{let r=await fetch('/v1/signals?quote='+q+'&limit='+l,{cache:'no-store'});let d=await r.json();last=d.signals||[];render();status.textContent='updated'}catch(e){status.textContent='load failed'}}function reset(){clearInterval(timer);timer=setInterval(loadNow,Number(interval.value));loadNow()}filter.onchange=render;interval.onchange=reset;reset()</script></body></html>`
