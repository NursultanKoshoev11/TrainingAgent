# TrainingAgent

Go microservice system for crypto-market research, news monitoring, public market-data ingestion, explainable watch signals, and a browser dashboard.

## Safety model

This project does not place trades and does not provide guaranteed financial advice. Dashboard labels are research/watch labels:

- `BUY_WATCH` means positive setup to research, not an instruction to buy.
- `SELL_WATCH` means negative/weak setup to review, not an instruction to sell.
- `HOLD_WATCH` means mixed evidence.
- `AVOID_WATCH` means risk is high or evidence is weak.

## Current services

| Service | Port | Purpose |
|---|---:|---|
| `news` | 8081 | Reads RSS feeds from `NEWS_FEEDS`, filters by query, and assigns keyword sentiment. |
| `market` | 8082 | Reads public 24h ticker market data and ranks markets by quote volume. |
| `engine` | 8083 | Combines market data, news sentiment, and risk into explainable watch signals. |
| `api-gateway` | 8080 | Public API, cached signal proxy, backend status endpoint, and embedded dashboard website. |

## Run with Docker

Both `compose.yaml` and `docker-compose.yml` use the same current service layout.

```bash
docker compose up --build
```

Or explicitly:

```bash
docker compose -f compose.yaml up --build
```

Open:

```text
http://localhost:8080/
```

API:

```bash
curl "http://localhost:8080/v1/signals?quote=USDT&limit=20"
```

Status:

```bash
curl "http://localhost:8080/v1/status"
```

## Run locally without Docker

Use 4 terminal windows:

```bash
go run ./cmd/news
```

```bash
go run ./cmd/market
```

```bash
MARKET_SERVICE_URL=http://localhost:8082 NEWS_SERVICE_URL=http://localhost:8081 go run ./cmd/engine
```

```bash
ENGINE_SERVICE_URL=http://localhost:8083 SIGNAL_CACHE_SECONDS=60 go run ./cmd/api-gateway
```

## News feeds

`news` reads comma-separated RSS feed URLs from `NEWS_FEEDS`.

Example:

```bash
NEWS_FEEDS="https://cointelegraph.com/rss,https://decrypt.co/feed" go run ./cmd/news
```

If feeds are unavailable, the service returns a small fallback dataset so the dashboard continues working.

## Dashboard

The dashboard auto-refreshes every 5 minutes by default. You can switch it to 10 minutes or 1 minute for local testing. It shows signal label, score, expected move estimate, risk, confidence, market context, related news, explanations, backend status, and simple summary metrics.

## Gateway cache

`api-gateway` caches successful `/v1/signals` responses for `SIGNAL_CACHE_SECONDS` seconds. Default: `60`. This keeps the dashboard responsive and reduces unnecessary requests to the engine.

## Roadmap

1. Persistent storage: PostgreSQL or ClickHouse.
2. Queue/event bus: NATS or Kafka.
3. Backtesting service.
4. Paper-trading simulator.
5. WebSocket market ingestion.
6. Probability calibration and model registry.
7. Auth and private deployment.
