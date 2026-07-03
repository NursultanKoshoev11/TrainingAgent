package analysis

import (
	"testing"

	"github.com/NursultanKoshoev11/TrainingAgent/internal/domain"
)

func TestSaferLabelNoSignalWhenNoTech(t *testing.T) {
	got := SaferLabel(0.8, 2, 0.2, 0.8, domain.TechSnapshot{}, false, 0.58)
	if got != "NO_SIGNAL" {
		t.Fatalf("expected NO_SIGNAL without tech, got %s", got)
	}
}

func TestSaferLabelAvoidsHighRisk(t *testing.T) {
	tech := domain.TechSnapshot{RSI: 55, TrendScore: 0.5, QualityScore: 0.8}
	got := SaferLabel(0.8, 2, 0.8, 0.8, tech, true, 0.58)
	if got != "AVOID_WATCH" {
		t.Fatalf("expected AVOID_WATCH for high risk, got %s", got)
	}
}

func TestSaferLabelBuyWatchNeedsConfirmation(t *testing.T) {
	tech := domain.TechSnapshot{RSI: 55, TrendScore: 0.6, MomentumScore: 0.4, VolumeScore: 0.3, QualityScore: 0.8}
	got := SaferLabel(0.75, 2, 0.3, 0.75, tech, true, 0.58)
	if got != "BUY_WATCH" {
		t.Fatalf("expected BUY_WATCH with strong confirmation, got %s", got)
	}
}
