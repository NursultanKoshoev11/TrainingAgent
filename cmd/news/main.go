package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/NursultanKoshoev11/TrainingAgent/internal/analysis"
	"github.com/NursultanKoshoev11/TrainingAgent/internal/domain"
	"github.com/NursultanKoshoev11/TrainingAgent/internal/platform"
)

type rssFeed struct {
	Channel struct {
		Title string `xml:"title"`
		Items []struct {
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			Description string `xml:"description"`
			PubDate     string `xml:"pubDate"`
		} `xml:"item"`
	} `xml:"channel"`
}

func main() {
	port := platform.GetEnv("PORT", "8081")
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", platform.Method(http.MethodGet, platform.HealthHandler("news")))
	mux.HandleFunc("/v1/news", platform.Method(http.MethodGet, handle))
	_ = platform.StartServer("news", port, mux)
}

func handle(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" { query = "crypto bitcoin ethereum binance" }
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 { limit = 20 }
	items := loadRSSNews(query, limit)
	if len(items) == 0 { items = fallbackNews() }
	for i := range items {
		items[i].Sentiment = analysis.SentimentScore(items[i].Title+" "+items[i].Summary)
		items[i].Matched = analysis.QueryMatches(items[i].Title+" "+items[i].Summary, query)
	}
	if len(items) > limit { items = items[:limit] }
	platform.JSON(w, http.StatusOK, domain.NewsResponse{Query: query, Count: len(items), Articles: items})
}

func loadRSSNews(query string, limit int) []domain.NewsArticle {
	feeds := platform.GetEnvCSV("NEWS_FEEDS", nil)
	if len(feeds) == 0 { return nil }
	client := &http.Client{Timeout: platform.HTTPTimeoutFromEnv("NEWS_HTTP_TIMEOUT_SECONDS", 10)}
	out := make([]domain.NewsArticle, 0, limit)
	for _, feedURL := range feeds {
		resp, err := client.Get(feedURL)
		if err != nil { continue }
		body, err := io.ReadAll(io.LimitReader(resp.Body, 2_000_000))
		_ = resp.Body.Close()
		if err != nil || resp.StatusCode < 200 || resp.StatusCode >= 300 { continue }
		var feed rssFeed
		if err := xml.Unmarshal(body, &feed); err != nil { continue }
		for _, item := range feed.Channel.Items {
			article := domain.NewsArticle{ID: stableID(item.Title+item.Link), Title: strings.TrimSpace(item.Title), Link: strings.TrimSpace(item.Link), Source: firstNonEmpty(feed.Channel.Title, feedURL), PublishedAt: parseTime(item.PubDate), Summary: strings.TrimSpace(stripTags(item.Description))}
			text := article.Title + " " + article.Summary
			if query == "" || analysis.QueryMatches(text, query) {
				out = append(out, article)
				if len(out) >= limit { return out }
			}
		}
	}
	return out
}

func fallbackNews() []domain.NewsArticle {
	now := time.Now().UTC()
	return []domain.NewsArticle{{ID:"fallback-1",Title:"Bitcoin bullish rally with inflows",Source:"fallback",PublishedAt:now,Summary:"market research fallback item"},{ID:"fallback-2",Title:"Exchange hack risk warning",Source:"fallback",PublishedAt:now,Summary:"risk research fallback item"}}
}

func stableID(value string) string { h := sha1.Sum([]byte(value)); return hex.EncodeToString(h[:]) }
func firstNonEmpty(a,b string) string { if strings.TrimSpace(a) != "" { return strings.TrimSpace(a) }; return strings.TrimSpace(b) }
func parseTime(value string) time.Time { for _, layout := range []string{time.RFC1123Z, time.RFC1123, time.RFC822Z, time.RFC822} { if t, err := time.Parse(layout, strings.TrimSpace(value)); err == nil { return t.UTC() } }; return time.Now().UTC() }
func stripTags(value string) string { var b strings.Builder; inside := false; for _, r := range value { if r == '<' { inside = true; continue }; if r == '>' { inside = false; continue }; if !inside { b.WriteRune(r) } }; return b.String() }
