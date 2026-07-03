package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/NursultanKoshoev11/TrainingAgent/internal/domain"
)

type Store struct { db *sql.DB }

type SignalHistoryItem struct {
	ID          int64           `json:"id"`
	Symbol      string          `json:"symbol"`
	Action      string          `json:"action"`
	Probability float64         `json:"probability"`
	RiskScore   float64         `json:"risk_score"`
	Confidence  float64         `json:"confidence"`
	Signal      json.RawMessage `json:"signal"`
	CreatedAt   time.Time       `json:"created_at"`
}

type PendingEvaluation struct {
	SignalID    int64           `json:"signal_id"`
	Symbol      string          `json:"symbol"`
	Action      string          `json:"action"`
	EntryPrice  float64         `json:"entry_price"`
	Signal      json.RawMessage `json:"signal"`
	CreatedAt   time.Time       `json:"created_at"`
}

type SignalEvaluation struct {
	ID                int64     `json:"id"`
	SignalID          int64     `json:"signal_id"`
	Symbol            string    `json:"symbol"`
	Action            string    `json:"action"`
	HorizonMinutes    int       `json:"horizon_minutes"`
	EntryPrice        float64   `json:"entry_price"`
	CheckedPrice      float64   `json:"checked_price"`
	MovePercent       float64   `json:"move_percent"`
	ExpectedDirection string    `json:"expected_direction"`
	ActualDirection   string    `json:"actual_direction"`
	Passed            bool      `json:"passed"`
	Reason            string    `json:"reason"`
	SignalCreatedAt   time.Time `json:"signal_created_at"`
	EvaluatedAt        time.Time `json:"evaluated_at"`
}

func Open(ctx context.Context, databaseURL string) (*Store, error) {
	if databaseURL == "" { return &Store{}, nil }
	db, err := sql.Open("pgx", databaseURL)
	if err != nil { return nil, err }
	db.SetMaxOpenConns(10); db.SetMaxIdleConns(5); db.SetConnMaxLifetime(30*time.Minute)
	store := &Store{db: db}
	if err := store.Migrate(ctx); err != nil { _ = db.Close(); return nil, err }
	return store, nil
}
func (s *Store) Enabled() bool { return s != nil && s.db != nil }
func (s *Store) Close() error { if !s.Enabled() { return nil }; return s.db.Close() }
func (s *Store) Ping(ctx context.Context) bool { if !s.Enabled() { return false }; return s.db.PingContext(ctx) == nil }

func (s *Store) Migrate(ctx context.Context) error {
	if !s.Enabled() { return nil }
	_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS signal_history (
  id BIGSERIAL PRIMARY KEY,
  symbol TEXT NOT NULL,
  action TEXT NOT NULL,
  probability DOUBLE PRECISION NOT NULL,
  risk_score DOUBLE PRECISION NOT NULL,
  confidence DOUBLE PRECISION NOT NULL,
  signal JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_signal_history_symbol_created_at ON signal_history(symbol, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_signal_history_action_created_at ON signal_history(action, created_at DESC);

CREATE TABLE IF NOT EXISTS signal_evaluations (
  id BIGSERIAL PRIMARY KEY,
  signal_id BIGINT NOT NULL REFERENCES signal_history(id) ON DELETE CASCADE,
  symbol TEXT NOT NULL,
  action TEXT NOT NULL,
  horizon_minutes INT NOT NULL,
  entry_price DOUBLE PRECISION NOT NULL,
  checked_price DOUBLE PRECISION NOT NULL,
  move_percent DOUBLE PRECISION NOT NULL,
  expected_direction TEXT NOT NULL,
  actual_direction TEXT NOT NULL,
  passed BOOLEAN NOT NULL,
  reason TEXT NOT NULL,
  signal_created_at TIMESTAMPTZ NOT NULL,
  evaluated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(signal_id, horizon_minutes)
);
CREATE INDEX IF NOT EXISTS idx_signal_evaluations_symbol_evaluated_at ON signal_evaluations(symbol, evaluated_at DESC);
CREATE INDEX IF NOT EXISTS idx_signal_evaluations_passed_evaluated_at ON signal_evaluations(passed, evaluated_at DESC);
`)
	return err
}

func (s *Store) SaveSignals(ctx context.Context, signals []domain.Signal) error {
	if !s.Enabled() || len(signals)==0 { return nil }
	tx, err := s.db.BeginTx(ctx, nil); if err != nil { return err }
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO signal_history(symbol, action, probability, risk_score, confidence, signal, created_at) VALUES($1,$2,$3,$4,$5,$6,$7)`)
	if err != nil { _ = tx.Rollback(); return err }
	defer stmt.Close()
	for _, signal := range signals {
		payload, err := json.Marshal(signal); if err != nil { _ = tx.Rollback(); return err }
		createdAt := signal.GeneratedAt; if createdAt.IsZero() { createdAt = time.Now().UTC() }
		if _, err := stmt.ExecContext(ctx, signal.Symbol, signal.Action, signal.Probability, signal.RiskScore, signal.Confidence, payload, createdAt); err != nil { _ = tx.Rollback(); return err }
	}
	return tx.Commit()
}

func (s *Store) SignalHistory(ctx context.Context, symbol, action string, limit int) ([]SignalHistoryItem, error) {
	if !s.Enabled() { return []SignalHistoryItem{}, nil }
	if limit <= 0 || limit > 500 { limit = 100 }
	rows, err := s.db.QueryContext(ctx, `SELECT id, symbol, action, probability, risk_score, confidence, signal, created_at FROM signal_history WHERE ($1 = '' OR symbol = $1) AND ($2 = '' OR action = $2) ORDER BY created_at DESC LIMIT $3`, symbol, action, limit)
	if err != nil { return nil, err }
	defer rows.Close()
	items := make([]SignalHistoryItem,0,limit)
	for rows.Next(){ var item SignalHistoryItem; if err := rows.Scan(&item.ID,&item.Symbol,&item.Action,&item.Probability,&item.RiskScore,&item.Confidence,&item.Signal,&item.CreatedAt); err != nil { return nil, err }; items=append(items,item) }
	return items, rows.Err()
}

func (s *Store) PendingEvaluations(ctx context.Context, horizonMinutes, limit int) ([]PendingEvaluation, error) {
	if !s.Enabled() { return []PendingEvaluation{}, nil }
	if horizonMinutes <= 0 { horizonMinutes = 20 }
	if limit <= 0 || limit > 1000 { limit = 200 }
	rows, err := s.db.QueryContext(ctx, `SELECT h.id, h.symbol, h.action, COALESCE((h.signal #>> '{market,last_price}')::double precision, 0), h.signal, h.created_at FROM signal_history h LEFT JOIN signal_evaluations e ON e.signal_id = h.id AND e.horizon_minutes = $1 WHERE e.id IS NULL AND h.created_at <= now() - make_interval(mins => $1) AND h.action IN ('BUY_WATCH','SELL_WATCH','HOLD_WATCH','AVOID_WATCH') ORDER BY h.created_at ASC LIMIT $2`, horizonMinutes, limit)
	if err != nil { return nil, err }
	defer rows.Close()
	items := make([]PendingEvaluation,0,limit)
	for rows.Next(){ var item PendingEvaluation; if err := rows.Scan(&item.SignalID,&item.Symbol,&item.Action,&item.EntryPrice,&item.Signal,&item.CreatedAt); err != nil { return nil, err }; items=append(items,item) }
	return items, rows.Err()
}

func (s *Store) SaveEvaluation(ctx context.Context, e SignalEvaluation) error {
	if !s.Enabled() { return nil }
	_, err := s.db.ExecContext(ctx, `INSERT INTO signal_evaluations(signal_id, symbol, action, horizon_minutes, entry_price, checked_price, move_percent, expected_direction, actual_direction, passed, reason, signal_created_at, evaluated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) ON CONFLICT(signal_id, horizon_minutes) DO NOTHING`, e.SignalID,e.Symbol,e.Action,e.HorizonMinutes,e.EntryPrice,e.CheckedPrice,e.MovePercent,e.ExpectedDirection,e.ActualDirection,e.Passed,e.Reason,e.SignalCreatedAt,time.Now().UTC())
	return err
}

func (s *Store) Evaluations(ctx context.Context, symbol string, limit int) ([]SignalEvaluation, error) {
	if !s.Enabled() { return []SignalEvaluation{}, nil }
	if limit <= 0 || limit > 500 { limit = 100 }
	rows, err := s.db.QueryContext(ctx, `SELECT id, signal_id, symbol, action, horizon_minutes, entry_price, checked_price, move_percent, expected_direction, actual_direction, passed, reason, signal_created_at, evaluated_at FROM signal_evaluations WHERE ($1 = '' OR symbol = $1) ORDER BY evaluated_at DESC LIMIT $2`, symbol, limit)
	if err != nil { return nil, err }
	defer rows.Close()
	items := make([]SignalEvaluation,0,limit)
	for rows.Next(){ var e SignalEvaluation; if err := rows.Scan(&e.ID,&e.SignalID,&e.Symbol,&e.Action,&e.HorizonMinutes,&e.EntryPrice,&e.CheckedPrice,&e.MovePercent,&e.ExpectedDirection,&e.ActualDirection,&e.Passed,&e.Reason,&e.SignalCreatedAt,&e.EvaluatedAt); err != nil { return nil, err }; items=append(items,e) }
	return items, rows.Err()
}
