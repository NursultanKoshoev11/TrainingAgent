package main

import "testing"

func TestNumParsesFloatAndFallbacksToZero(t *testing.T) {
	if got := num("12.5"); got != 12.5 {
		t.Fatalf("expected 12.5, got %v", got)
	}
	if got := num("bad"); got != 0 {
		t.Fatalf("expected zero for bad input, got %v", got)
	}
}

func TestMax(t *testing.T) {
	if got := max(5, 3); got != 5 {
		t.Fatalf("expected 5, got %d", got)
	}
	if got := max(2, 9); got != 9 {
		t.Fatalf("expected 9, got %d", got)
	}
}

func TestFallbackTickersUseQuote(t *testing.T) {
	items := fallback("USDT")
	if len(items) == 0 {
		t.Fatalf("expected fallback tickers")
	}
	for _, item := range items {
		if item.Symbol == "" || item.LastPrice <= 0 {
			t.Fatalf("expected populated ticker: %+v", item)
		}
	}
}
