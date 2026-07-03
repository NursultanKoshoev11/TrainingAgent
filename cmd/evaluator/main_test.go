package main

import (
	"testing"
	"time"

	"github.com/NursultanKoshoev11/TrainingAgent/internal/storage"
)

func TestEvaluateBuyWatchPassesWhenPriceGoesUp(t *testing.T) {
	item := storage.PendingEvaluation{SignalID: 1, Symbol: "BTCUSDT", Action: "BUY_WATCH", EntryPrice: 100, CreatedAt: time.Now().Add(-20 * time.Minute)}
	got := evaluate(item, 105, 20)
	if !got.Passed { t.Fatalf("expected BUY_WATCH to pass when price goes up: %+v", got) }
	if got.ActualDirection != "up" { t.Fatalf("expected actual direction up, got %s", got.ActualDirection) }
}

func TestEvaluateSellWatchPassesWhenPriceGoesDown(t *testing.T) {
	item := storage.PendingEvaluation{SignalID: 2, Symbol: "ETHUSDT", Action: "SELL_WATCH", EntryPrice: 100, CreatedAt: time.Now().Add(-20 * time.Minute)}
	got := evaluate(item, 95, 20)
	if !got.Passed { t.Fatalf("expected SELL_WATCH to pass when price goes down: %+v", got) }
	if got.ActualDirection != "down" { t.Fatalf("expected actual direction down, got %s", got.ActualDirection) }
}

func TestEvaluateHoldWatchPassesWhenPriceIsFlat(t *testing.T) {
	item := storage.PendingEvaluation{SignalID: 3, Symbol: "BNBUSDT", Action: "HOLD_WATCH", EntryPrice: 100, CreatedAt: time.Now().Add(-20 * time.Minute)}
	got := evaluate(item, 100.2, 20)
	if !got.Passed { t.Fatalf("expected HOLD_WATCH to pass when price is flat: %+v", got) }
}

func TestEvaluateAvoidWatchPassesWhenMoveIsVolatile(t *testing.T) {
	item := storage.PendingEvaluation{SignalID: 4, Symbol: "SOLUSDT", Action: "AVOID_WATCH", EntryPrice: 100, CreatedAt: time.Now().Add(-20 * time.Minute)}
	got := evaluate(item, 102, 20)
	if !got.Passed { t.Fatalf("expected AVOID_WATCH to pass when move is volatile: %+v", got) }
}
