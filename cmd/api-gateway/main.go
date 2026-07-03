package main

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/NursultanKoshoev11/TrainingAgent/internal/platform"
)

func main() {
	port := platform.GetEnv("PORT", "8080")
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", platform.Method(http.MethodGet, platform.HealthHandler("api-gateway")))
	mux.HandleFunc("/v1/signals", platform.Method(http.MethodGet, proxySignals))
	mux.HandleFunc("/", platform.Method(http.MethodGet, dashboard))
	_ = platform.StartServer("api-gateway", port, mux)
}

func proxySignals(w http.ResponseWriter, r *http.Request) {
	base := strings.TrimRight(platform.GetEnv("ENGINE_SERVICE_URL", "http://localhost:8083"), "/")
	upstream := base + "/v1/signals"
	if r.URL.RawQuery != "" { upstream += "?" + r.URL.RawQuery }
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Get(upstream)
	if err != nil { platform.Fail(w, http.StatusBadGateway, err.Error()); return }
	defer resp.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.WriteString(w, page)
}

const page = `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>TrainingAgent Dashboard</title><style>body{font-family:Arial;background:#0b1020;color:#edf3ff;margin:0}.wrap{max-width:1200px;margin:auto;padding:24px}.card{background:#121a2f;border:1px solid #24304b;border-radius:16px;padding:16px;margin:12px 0}button,select,input{padding:10px;border-radius:10px;border:1px solid #33415f;background:#17213a;color:#fff}table{width:100%;border-collapse:collapse}td,th{padding:10px;border-bottom:1px solid #24304b;text-align:left}.BUY_WATCH{color:#4ade80}.SELL_WATCH{color:#fb7185}.HOLD_WATCH{color:#facc15}.AVOID_WATCH{color:#f97316}</style></head><body><div class="wrap"><h1>TrainingAgent Crypto Research Dashboard</h1><p>Research/watch signals only. Not financial advice and not auto-trading.</p><div class="card"><select id="quote"><option>USDT</option><option>BTC</option><option>ETH</option></select> <input id="limit" value="20" type="number" min="1" max="100"> <select id="interval"><option value="300000">5 min</option><option value="600000">10 min</option><option value="60000">1 min test</option></select> <button onclick="loadNow()">Refresh</button> <span id="status"></span></div><div class="card"><table><thead><tr><th>Symbol</th><th>Label</th><th>Probability</th><th>Expected move</th><th>Risk</th><th>Confidence</th><th>24h</th></tr></thead><tbody id="rows"></tbody></table></div><div class="card"><h2>Details</h2><pre id="details">Select a row</pre></div></div><script>let timer;async function loadNow(){status.textContent='loading...';let q=quote.value,l=limit.value;let r=await fetch('/v1/signals?quote='+q+'&limit='+l,{cache:'no-store'});let d=await r.json();rows.innerHTML='';(d.signals||[]).forEach(s=>{let tr=document.createElement('tr');tr.innerHTML='<td>'+s.symbol+'</td><td class="'+s.action+'">'+s.action+'</td><td>'+Math.round(s.probability*100)+'%</td><td>'+s.expected_move_percent+'%</td><td>'+Math.round(s.risk_score*100)+'%</td><td>'+Math.round(s.confidence*100)+'%</td><td>'+s.market.price_change_percent+'%</td>';tr.onclick=()=>details.textContent=JSON.stringify(s,null,2);rows.appendChild(tr)});status.textContent='updated '+new Date().toLocaleTimeString()}function reset(){clearInterval(timer);timer=setInterval(loadNow,Number(interval.value));loadNow()}interval.onchange=reset;reset()</script></body></html>`
