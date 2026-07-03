package analysis

import (
	"testing"

	"github.com/NursultanKoshoev11/TrainingAgent/internal/domain"
)

func TestClampBounds(t *testing.T) {
	if got := Clamp(12, 0, 10); got != 10 {
		t.Fatalf("Clamp upper bound = %v", got)
	}
	if got := Clamp(-2, 0, 10); got != 0 {
		t.Fatalf("Clamp lower bound = %v", got)
	}
	if got := Clamp(5, 0, 10); got != 5 {
		t.Fatalf("Clamp middle value = %v", got)
	}
}

func TestSentimentScorePolarity(t *testing.T) {
	positive := SentimentScore("bullish rally growth inflows")
	negative := SentimentScore("hack exploit warning risk")
	if positive <= 0 {
		t.Fatalf("expected positive sentiment, got %v", positive)
	}
	if negative >= 0 {
		t.Fatalf("expected negative sentiment, got %v", negative)
	}
}

func TestBuildSignalsReturnsRankedResearchItems(t *testing.T) {
	tickers := []domain.Ticker{
		{Symbol: "BTCUSDT", LastPrice: 100, OpenPrice: 95, HighPrice: 105, LowPrice: 94, PriceChangePercent: 5.5, QuoteVolume: 1000000, VolumeRankScore: 1},
		{Symbol: "ETHUSDT", LastPrice: 50, OpenPrice: 51, HighPrice: 54, LowPrice: 48, PriceChangePercent: -1.2, QuoteVolume: 700000, VolumeRankScore: 0.7},
	}
	articles := []domain.NewsArticle{
		{ID: "n1", Title: "Bitcoin bullish rally with inflows", Sentiment: 0.5, Matched: true},
	}

	items := BuildSignals(tickers, articles, "USDT", 10, ScoreConfig{MinimumProbability: 0.58})
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Symbol == "" || items[0].Action == "" {
		t.Fatalf("expected populated signal, got %+v", items[0])
	}
	if items[0].Probability < items[1].Probability {
		t.Fatalf("expected signals sorted by probability: %+v", items)
	}
	if items[0].RiskScore < 0 || items[0].RiskScore > 1 {
		t.Fatalf("risk score out of range: %v", items[0].RiskScore)
	}
	if len(items[0].Reasons) == 0 {
		t.Fatalf("expected explanatory reasons")
	}
}
