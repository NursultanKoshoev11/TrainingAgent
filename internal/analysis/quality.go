package analysis

import "github.com/NursultanKoshoev11/TrainingAgent/internal/domain"

func ExtraScoreFromTech(t domain.TechSnapshot, ok bool) float64 {
	if !ok { return -0.25 }
	return Clamp(t.TrendScore*0.45+t.MomentumScore*0.35+t.VolumeScore*0.20, -1, 1)
}

func ExtraRiskFromTech(t domain.TechSnapshot, ok bool) float64 {
	if !ok { return 0.25 }
	extra := 0.0
	if t.Volatility > 3.0 { extra += Clamp((t.Volatility-3)/8, 0, 0.25) }
	if t.RSI > 78 { extra += 0.12 }
	if t.RSI < 25 { extra += 0.08 }
	if t.QualityScore < 0.35 { extra += 0.10 }
	return extra
}

func SaferLabel(prob, expected, risk, confidence float64, tech domain.TechSnapshot, ok bool, minProb float64) string {
	if minProb <= 0 { minProb = 0.58 }
	if !ok || confidence < 0.42 { return "NO_SIGNAL" }
	if risk >= 0.72 || tech.Volatility > 4.0 { return "AVOID_WATCH" }
	if risk >= 0.62 { return "HOLD_WATCH" }
	if tech.RSI > 76 && expected > 0 { return "HOLD_WATCH" }
	if prob >= minProb && expected > 0.25 && confidence >= 0.58 && tech.TrendScore > 0.15 { return "BUY_WATCH" }
	if prob < 0.44 || expected < -1.2 || tech.TrendScore < -0.35 { return "SELL_WATCH" }
	return "HOLD_WATCH"
}
