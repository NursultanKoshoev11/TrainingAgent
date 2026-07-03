package analysis

import (
	"testing"
	"time"

	"github.com/NursultanKoshoev11/TrainingAgent/internal/domain"
)

func TestEMAUsesLatestTrend(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5}
	got := EMA(values, 3)
	if got <= 3 {
		t.Fatalf("expected EMA to follow rising values, got %v", got)
	}
}

func TestRSIRisingMarketIsHigh(t *testing.T) {
	values := []float64{1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16}
	got := RSI(values, 14)
	if got < 90 {
		t.Fatalf("expected high RSI for rising series, got %v", got)
	}
}

func TestVolatilityPercent(t *testing.T) {
	candles := []domain.Candle{{Close:100,High:105,Low:95},{Close:100,High:102,Low:98}}
	got := VolatilityPercent(candles)
	if got <= 0 {
		t.Fatalf("expected positive volatility, got %v", got)
	}
}

func TestVolumeSpikeScore(t *testing.T) {
	candles := make([]domain.Candle, 0, 11)
	for i := 0; i < 10; i++ { candles = append(candles, domain.Candle{Volume:100,Close:100,OpenTime:time.Now()}) }
	candles = append(candles, domain.Candle{Volume:400,Close:100})
	if got := VolumeSpikeScore(candles); got <= 0 {
		t.Fatalf("expected volume spike score, got %v", got)
	}
}
