package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/NursultanKoshoev11/TrainingAgent/internal/domain"
)

type Store struct {
	db *sql.DB
}

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

func Open(ctx context.Context, databaseURL string) (*Store, error) {
	if databaseURL == "" {
		return &Store{}, nil
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	store := &Store{db: db}
	if err := store.Migrate(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Enabled() bool {
	return s != nil && s.db != nil
}

func (s *Store) Close() error {
	if !s.Enabled() {
		return nil
	}
	return s.db.Close()
}

func (s *Store) Ping(ctx context.Context) bool {
	if !s.Enabled() {
		return false
	}
	return s.db.PingContext(ctx) == nil
}

func (s *Store) Migrate(ctx context.Context) error {
	if !s.Enabled() {
		return nil
	}
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
`)
	return err
}

func (s *Store) SaveSignals(ctx context.Context, signals []domain.Signal) error {
	if !s.Enabled() || len(signals) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO signal_history(symbol, action, probability, risk_score, confidence, signal, created_at) VALUES($1,$2,$3,$4,$5,$6,$7)`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, signal := range signals {
		payload, err := json.Marshal(signal)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		createdAt := signal.GeneratedAt
		if createdAt.IsZero() {
			createdAt = time.Now().UTC()
		}
		if _, err := stmt.ExecContext(ctx, signal.Symbol, signal.Action, signal.Probability, signal.RiskScore, signal.Confidence, payload, createdAt); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) SignalHistory(ctx context.Context, symbol, action string, limit int) ([]SignalHistoryItem, error) {
	if !s.Enabled() {
		return []SignalHistoryItem{}, nil
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	query := `SELECT id, symbol, action, probability, risk_score, confidence, signal, created_at FROM signal_history WHERE ($1 = '' OR symbol = $1) AND ($2 = '' OR action = $2) ORDER BY created_at DESC LIMIT $3`
	rows, err := s.db.QueryContext(ctx, query, symbol, action, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]SignalHistoryItem, 0, limit)
	for rows.Next() {
		var item SignalHistoryItem
		if err := rows.Scan(&item.ID, &item.Symbol, &item.Action, &item.Probability, &item.RiskScore, &item.Confidence, &item.Signal, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
