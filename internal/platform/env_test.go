package platform

import (
	"reflect"
	"testing"
	"time"
)

func TestGetEnvFallbackAndValue(t *testing.T) {
	t.Setenv("TA_TEST_ENV", "custom")
	if got := GetEnv("TA_TEST_ENV", "fallback"); got != "custom" {
		t.Fatalf("expected custom value, got %q", got)
	}
	if got := GetEnv("TA_MISSING_ENV", "fallback"); got != "fallback" {
		t.Fatalf("expected fallback, got %q", got)
	}
}

func TestGetEnvIntAndFloatFallbacks(t *testing.T) {
	t.Setenv("TA_TEST_INT", "42")
	t.Setenv("TA_BAD_INT", "bad")
	t.Setenv("TA_TEST_FLOAT", "0.75")
	if got := GetEnvInt("TA_TEST_INT", 1); got != 42 {
		t.Fatalf("expected int 42, got %d", got)
	}
	if got := GetEnvInt("TA_BAD_INT", 7); got != 7 {
		t.Fatalf("expected fallback int 7, got %d", got)
	}
	if got := GetEnvFloat("TA_TEST_FLOAT", 0.1); got != 0.75 {
		t.Fatalf("expected float 0.75, got %v", got)
	}
}

func TestGetEnvCSV(t *testing.T) {
	t.Setenv("TA_TEST_CSV", " one, two ,, three ")
	want := []string{"one", "two", "three"}
	if got := GetEnvCSV("TA_TEST_CSV", nil); !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestHTTPTimeoutFromEnv(t *testing.T) {
	t.Setenv("TA_TEST_TIMEOUT", "3")
	if got := HTTPTimeoutFromEnv("TA_TEST_TIMEOUT", 10); got != 3*time.Second {
		t.Fatalf("expected 3s, got %v", got)
	}
}
