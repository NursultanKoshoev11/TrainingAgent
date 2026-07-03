package analysis

import (
	"fmt"
	"sort"
	"time"

	"github.com/NursultanKoshoev11/TrainingAgent/internal/domain"
)

type ScoreConfig struct {
	MinimumProbability float64
}

func BuildSignals(tickers []domain.Ticker, articles []domain.NewsArticle, quote string, limit int, cfg ScoreConfig) []domain.Signal {
	if limit <= 0 { limit = 20 }
	items := make([]domain.Signal, 0, len(tickers))
	for _, ticker := range tickers {
		news := relatedNews(ticker.Symbol, articles, 3)
		avg := averageSentiment(news)
		risk := riskScore(ticker)
		prob := Clamp(0.50+Clamp(ticker.PriceChangePercent/25, -0.25, 0.25)+Clamp(ticker.VolumeRankScore*0.12, 0, 0.12)+Clamp(avg*0.15, -0.15, 0.15)-Clamp(risk*0.18, 0, 0.18), 0.05, 0.95)
		expected := Clamp(ticker.PriceChangePercent*0.35+avg*2.5, -20, 20)
		action := "HOLD_WATCH"
		if risk >= 0.75 { action = "AVOID_WATCH" } else if prob >= cfg.MinimumProbability && expected > 0 { action = "BUY_WATCH" } else if prob < 0.45 || expected < -1.5 { action = "SELL_WATCH" }
		items = append(items, domain.Signal{Symbol: ticker.Symbol, Action: action, Probability: round(prob), ExpectedMovePercent: round(expected), RiskScore: round(risk), Confidence: round(Clamp(0.4+ticker.VolumeRankScore*0.3+float64(len(news))*0.08, 0, 0.95)), GeneratedAt: time.Now().UTC(), Reasons: []string{fmt.Sprintf("Изменение цены за 24 часа: %.2f%%", ticker.PriceChangePercent), fmt.Sprintf("Сила объёма относительно других монет: %.2f", ticker.VolumeRankScore), fmt.Sprintf("Средний фон связанных новостей: %.2f", avg), fmt.Sprintf("Оценка риска по волатильности: %.2f", risk)}, Market: ticker, News: news})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Probability > items[j].Probability })
	if len(items) > limit { return items[:limit] }
	return items
}

func averageSentiment(news []domain.NewsArticle) float64 {
	if len(news) == 0 { return 0 }
	var total float64
	for _, n := range news { total += n.Sentiment }
	return total / float64(len(news))
}

func riskScore(t domain.Ticker) float64 {
	if t.LastPrice <= 0 { return 0.9 }
	rangePct := ((t.HighPrice - t.LowPrice) / t.LastPrice) * 100
	return Clamp(rangePct/18, 0, 1)
}

func relatedNews(symbol string, articles []domain.NewsArticle, limit int) []domain.NewsArticle {
	base := symbol
	if len(base) > 4 { base = base[:len(base)-4] }
	out := make([]domain.NewsArticle, 0, limit)
	for _, a := range articles {
		text := a.Title + " " + a.Summary
		if QueryMatches(text, base) || QueryMatches(text, "crypto bitcoin ethereum binance") {
			out = append(out, a)
			if len(out) >= limit { break }
		}
	}
	return out
}

func round(v float64) float64 { return float64(int(v*10000+0.5)) / 10000 }
