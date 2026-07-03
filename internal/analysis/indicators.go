package analysis

import (
	"math"

	"github.com/NursultanKoshoev11/TrainingAgent/internal/domain"
)

func EMA(values []float64, period int) float64 {
	if len(values) == 0 { return 0 }
	if period <= 1 { return values[len(values)-1] }
	k := 2.0 / float64(period+1)
	ema := values[0]
	for _, v := range values[1:] { ema = v*k + ema*(1-k) }
	return ema
}

func RSI(values []float64, period int) float64 {
	if len(values) <= period || period <= 0 { return 50 }
	var gain, loss float64
	start := len(values) - period
	for i := start; i < len(values); i++ {
		change := values[i] - values[i-1]
		if change >= 0 { gain += change } else { loss += -change }
	}
	if loss == 0 { return 100 }
	rs := (gain / float64(period)) / (loss / float64(period))
	return 100 - (100 / (1 + rs))
}

func VolatilityPercent(candles []domain.Candle) float64 {
	if len(candles) == 0 { return 0 }
	var total float64
	for _, c := range candles {
		if c.Close > 0 { total += ((c.High - c.Low) / c.Close) * 100 }
	}
	return total / float64(len(candles))
}

func VolumeSpikeScore(candles []domain.Candle) float64 {
	if len(candles) < 10 { return 0 }
	last := candles[len(candles)-1].Volume
	var avg float64
	for _, c := range candles[:len(candles)-1] { avg += c.Volume }
	avg = avg / float64(len(candles)-1)
	if avg <= 0 { return 0 }
	return Clamp((last/avg-1)/2, 0, 1)
}

func CloseValues(candles []domain.Candle) []float64 {
	out := make([]float64, 0, len(candles))
	for _, c := range candles { out = append(out, c.Close) }
	return out
}

func MomentumScore(candles []domain.Candle) float64 {
	if len(candles) < 2 { return 0 }
	first := candles[0].Close
	last := candles[len(candles)-1].Close
	if first <= 0 { return 0 }
	return Clamp(((last-first)/first)*10, -1, 1)
}

func TrendScoreFromEMA(ema20, ema50, price float64) float64 {
	if price <= 0 || ema20 <= 0 || ema50 <= 0 { return 0 }
	score := 0.0
	if price > ema20 { score += 0.35 } else { score -= 0.35 }
	if ema20 > ema50 { score += 0.45 } else { score -= 0.45 }
	gap := math.Abs(ema20-ema50) / price
	return Clamp(score + Clamp(gap*10, 0, 0.2), -1, 1)
}
