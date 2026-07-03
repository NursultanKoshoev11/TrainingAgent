package main

import (
	"net/http"
	"testing"
	"time"
)

func TestSignalCacheHitAndMiss(t *testing.T) {
	signalCache.Lock()
	signalCache.query = ""
	signalCache.body = nil
	signalCache.status = http.StatusOK
	signalCache.updatedAt = time.Time{}
	signalCache.Unlock()

	if _, _, ok := readSignalCache("quote=USDT", 60); ok {
		t.Fatalf("expected empty cache miss")
	}

	writeSignalCache("quote=USDT", []byte(`{"ok":true}`), http.StatusOK)
	body, status, ok := readSignalCache("quote=USDT", 60)
	if !ok {
		t.Fatalf("expected cache hit")
	}
	if status != http.StatusOK {
		t.Fatalf("expected status 200, got %d", status)
	}
	if string(body) != `{"ok":true}` {
		t.Fatalf("unexpected cached body: %s", string(body))
	}
}

func TestSignalCacheDifferentQueryMiss(t *testing.T) {
	writeSignalCache("quote=USDT", []byte(`{"ok":true}`), http.StatusOK)
	if _, _, ok := readSignalCache("quote=BTC", 60); ok {
		t.Fatalf("expected miss for different query")
	}
}

func TestSignalCacheExpiredMiss(t *testing.T) {
	writeSignalCache("quote=USDT", []byte(`{"ok":true}`), http.StatusOK)
	signalCache.Lock()
	signalCache.updatedAt = time.Now().Add(-2 * time.Minute)
	signalCache.Unlock()
	if _, _, ok := readSignalCache("quote=USDT", 1); ok {
		t.Fatalf("expected expired cache miss")
	}
}

func TestSignalCacheSkipsErrorStatus(t *testing.T) {
	signalCache.Lock()
	signalCache.query = ""
	signalCache.body = nil
	signalCache.status = http.StatusOK
	signalCache.updatedAt = time.Time{}
	signalCache.Unlock()

	writeSignalCache("quote=USDT", []byte(`{"error":true}`), http.StatusBadGateway)
	if _, _, ok := readSignalCache("quote=USDT", 60); ok {
		t.Fatalf("expected miss because error responses are not cached")
	}
}
