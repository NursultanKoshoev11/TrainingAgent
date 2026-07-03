package analysis

import (
	"testing"

	"github.com/NursultanKoshoev11/TrainingAgent/internal/domain"
)

func TestBuildGuardedSignalsUsesNoSignalWhenTechMissing(t *testing.T) {
	tickers := []domain.Ticker{{Symbol:"BTCUSDT", LastPrice:100, OpenPrice:99, HighPrice:101, LowPrice:98, PriceChangePercent:1, VolumeRankScore:1}}
	items := BuildGuardedSignals(tickers, nil, map[string]domain.TechSnapshot{}, "USDT", 10, ScoreConfig{MinimumProbability:0.58})
	if len(items) != 1 { t.Fatalf("expected one signal, got %d", len(items)) }
	if items[0].Action != "NO_SIGNAL" { t.Fatalf("expected NO_SIGNAL without tech, got %s", items[0].Action) }
}

func TestBuildGuardedSignalsAllowsStrongConfirmedSetup(t *testing.T) {
	tickers := []domain.Ticker{{Symbol:"ETHUSDT", LastPrice:100, OpenPrice:95, HighPrice:102, LowPrice:94, PriceChangePercent:4, VolumeRankScore:1}}
	tech := map[string]domain.TechSnapshot{"ETHUSDT":{RSI:55, TrendScore:0.7, MomentumScore:0.5, VolumeScore:0.4, QualityScore:0.9, Volatility:1.2}}
	items := BuildGuardedSignals(tickers, nil, tech, "USDT", 10, ScoreConfig{MinimumProbability:0.58})
	if len(items) != 1 { t.Fatalf("expected one signal, got %d", len(items)) }
	if items[0].Action == "NO_SIGNAL" || items[0].Action == "AVOID_WATCH" { t.Fatalf("expected usable watch label, got %s", items[0].Action) }
}
