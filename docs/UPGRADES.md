# TrainingAgent upgrades

This document describes the larger MVP upgrades added after the initial dashboard.

## Added capabilities

- PostgreSQL-backed signal history.
- Automatic signal evaluation after a configurable horizon, default 20 minutes.
- Optional storage: if `DATABASE_URL` is empty, the system still runs, but history and evaluations are disabled.
- Market candles endpoint for OHLCV data.
- Gateway routes for signals, signal history, signal evaluations, candles, and status.
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
- `evaluator` on port `8084`
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

Signal evaluations:

```bash
curl "http://localhost:8080/v1/evaluations?limit=50"
```

Manual evaluation run:

```bash
curl -X POST "http://localhost:8080/v1/evaluations/run"
```

Candles:

```bash
curl "http://localhost:8080/v1/candles?symbol=BTCUSDT&interval=5m&limit=100"
```

Status:

```bash
curl "http://localhost:8080/v1/status"
```

## Evaluation logic

The evaluator checks old signals after `EVALUATION_HORIZON_MINUTES`, default `20`.

- `BUY_WATCH` passes if price moved up.
- `SELL_WATCH` passes if price moved down.
- `HOLD_WATCH` passes if price stayed almost flat.
- `AVOID_WATCH` passes if the move was volatile, confirming that risk was high.

Results are saved in `signal_evaluations` and can be used later to improve scoring.

## Optional auth

Set both variables to enable Basic Auth:

```bash
DASHBOARD_USER=admin
DASHBOARD_PASSWORD=change-me
```

Leave them empty for local open access.

## Database

The engine and evaluator read `DATABASE_URL`. Example:

```text
postgres://trainingagent:trainingagent@localhost:5432/trainingagent?sslmode=disable
```

The app creates the `signal_history` and `signal_evaluations` tables automatically.

## Next major items

- Backtesting service over longer historical windows.
- Technical indicators from candles.
- Better news classifier and symbol extraction.
- Full status page with last successful fetch timestamps.
- Structured logs and request IDs.
- Learning loop that adjusts scoring weights based on evaluation results.
