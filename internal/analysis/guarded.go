package analysis

import (
	"fmt"
	"sort"
	"time"

	"github.com/NursultanKoshoev11/TrainingAgent/internal/domain"
)

func BuildGuardedSignals(tickers []domain.Ticker, articles []domain.NewsArticle, tech map[string]domain.TechSnapshot, quote string, limit int, cfg ScoreConfig) []domain.Signal {
	if limit <= 0 { limit = 20 }
	items := make([]domain.Signal, 0, len(tickers))
	for _, ticker := range tickers {
		news := relatedNews(ticker.Symbol, articles, 3)
		avg := averageSentiment(news)
		technical, ok := tech[ticker.Symbol]
		techScore := ExtraScoreFromTech(technical, ok)
		risk := Clamp(riskScore(ticker)+ExtraRiskFromTech(technical, ok), 0, 1)
		prob := Clamp(0.50+Clamp(ticker.PriceChangePercent/25,-0.20,0.20)+Clamp(ticker.VolumeRankScore*0.10,0,0.10)+Clamp(avg*0.12,-0.12,0.12)+Clamp(techScore*0.18,-0.18,0.18)-Clamp(risk*0.22,0,0.22),0.05,0.95)
		expected := Clamp(ticker.PriceChangePercent*0.25+avg*2.0+techScore*2.0,-20,20)
		confidence := Clamp(0.30+ticker.VolumeRankScore*0.22+float64(len(news))*0.06+technical.QualityScore*0.28-Clamp(risk*0.18,0,0.18),0,0.95)
		action := SaferLabel(prob, expected, risk, confidence, technical, ok, cfg.MinimumProbability)
		reasons := []string{fmt.Sprintf("spot mode"), fmt.Sprintf("24h %.2f", ticker.PriceChangePercent), fmt.Sprintf("volume %.2f", ticker.VolumeRankScore), fmt.Sprintf("news %.2f", avg), fmt.Sprintf("risk %.2f", risk), fmt.Sprintf("confidence %.2f", confidence)}
		if ok { reasons = append(reasons, fmt.Sprintf("rsi %.2f", technical.RSI), fmt.Sprintf("trend %.2f momentum %.2f volume %.2f", technical.TrendScore, technical.MomentumScore, technical.VolumeScore), fmt.Sprintf("volatility %.2f", technical.Volatility)) }
		items = append(items, domain.Signal{Symbol:ticker.Symbol, Action:action, Probability:round(prob), ExpectedMovePercent:round(expected), RiskScore:round(risk), Confidence:round(confidence), GeneratedAt:time.Now().UTC(), Reasons:reasons, Market:ticker, News:news})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Probability > items[j].Probability })
	if len(items) > limit { return items[:limit] }
	return items
}
