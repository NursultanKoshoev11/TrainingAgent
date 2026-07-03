package main

import (
	"strings"
	"testing"
	"time"
)

func TestStripTagsRemovesSimpleHTML(t *testing.T) {
	got := stripTags("<p>Hello <b>crypto</b></p>")
	if strings.TrimSpace(got) != "Hello crypto" {
		t.Fatalf("unexpected stripped text: %q", got)
	}
}

func TestStableIDIsStable(t *testing.T) {
	first := stableID("same-input")
	second := stableID("same-input")
	if first == "" || first != second {
		t.Fatalf("stableID should be deterministic: %q %q", first, second)
	}
}

func TestParseTimeKnownLayout(t *testing.T) {
	got := parseTime("Mon, 02 Jan 2006 15:04:05 -0700")
	if got.IsZero() {
		t.Fatalf("expected parsed time")
	}
	if got.Year() != 2006 || got.Month() != time.January || got.Day() != 2 {
		t.Fatalf("unexpected parsed date: %v", got)
	}
}

func TestFallbackNewsHasItems(t *testing.T) {
	items := fallbackNews()
	if len(items) == 0 {
		t.Fatalf("expected fallback news items")
	}
	if items[0].Title == "" || items[0].Source == "" {
		t.Fatalf("expected populated fallback item: %+v", items[0])
	}
}
