# TrainingAgent upgrades

This document describes the larger MVP upgrades added after the initial dashboard.

## Added capabilities

- PostgreSQL-backed signal history.
- Optional storage: if `DATABASE_URL` is empty, the system still runs, but history is disabled.
- Market candles endpoint for OHLCV data.
- Gateway routes for signals, signal history, candles, and status.
- Optional Basic Auth for the dashboard and API gateway.
- Shared retry helper for future external calls.

## Docker stack

Use the main compose stack:

```bash
docker compose -f compose.yaml up --build
```

It includes:

- `postgres` on port `5432`
- `news` on port `8081`
- `market` on port `8082`
- `engine` on port `8083`
- `api-gateway` on port `8080`

Open dashboard:

```text
http://localhost:8080/
```

## API endpoints

Signals:

```bash
curl "http://localhost:8080/v1/signals?quote=USDT&limit=20"
```

Signal history:

```bash
curl "http://localhost:8080/v1/signals/history?limit=30"
```

Candles:

```bash
curl "http://localhost:8080/v1/candles?symbol=BTCUSDT&interval=5m&limit=100"
```

Status:

```bash
curl "http://localhost:8080/v1/status"
```

## Optional auth

Set both variables to enable Basic Auth:

```bash
DASHBOARD_USER=admin
DASHBOARD_PASSWORD=change-me
```

Leave them empty for local open access.

## Database

The engine reads `DATABASE_URL`. Example:

```text
postgres://trainingagent:trainingagent@localhost:5432/trainingagent?sslmode=disable
```

The app creates the `signal_history` table automatically.

## Next major items

- Backtesting service.
- Technical indicators from candles.
- Better news classifier and symbol extraction.
- Full status page with last successful fetch timestamps.
- Structured logs and request IDs.
- Paper-trading simulator.
