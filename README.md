# TrainingAgent

Go microservice system for crypto-market research, news monitoring, Binance public market ingestion, explainable watch signals, and a browser dashboard.

## Safety model

This project does not place trades and does not provide guaranteed financial advice. Dashboard labels are research/watch labels:

- `BUY_WATCH` means positive setup to research, not an instruction to buy.
- `SELL_WATCH` means negative/weak setup to review, not an instruction to sell.
- `HOLD_WATCH` means mixed evidence.
- `AVOID_WATCH` means risk is high or evidence is weak.

## Services

| Service | Port | Purpose |
|---|---:|---|
| `news-ingestor` | 8081 | Reads crypto/news RSS feeds and assigns basic keyword sentiment. |
| `binance-ingestor` | 8082 | Reads Binance public 24h ticker data and ranks markets. |
| `signal-engine` | 8083 | Combines market data, news sentiment, and risk into explainable watch signals. |
| `api-gateway` | 8080 | Public API and embedded dashboard website. |

## Run

```bash
make test
make build
docker compose up --build
```

Open:

```text
http://localhost:8080/
```

API:

```bash
curl "http://localhost:8080/v1/signals?quote=USDT&limit=20"
```

## Dashboard

The dashboard auto-refreshes every 5 minutes by default. You can switch it to 10 minutes or 1 minute for local testing. It shows signal label, probability, expected move estimate, risk, confidence, market context, related news, and explanations.

## Roadmap

1. Persistent storage: PostgreSQL or ClickHouse.
2. Queue/event bus: NATS or Kafka.
3. Backtesting service.
4. Paper-trading simulator.
5. WebSocket market ingestion.
6. Probability calibration and model registry.
7. Auth and private deployment.
