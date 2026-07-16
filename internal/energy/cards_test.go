package energy_test

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/energy"
)

func TestCriticalLimitUmJedenPreis(t *testing.T) {
	c := energy.Card{ID: "um-jeden-preis"}
	if c.CriticalLimit() != 10 {
		t.Fatalf("limit = %d, want 10", c.CriticalLimit())
	}
}

func TestCriticalLimitEroeffnungsfeier(t *testing.T) {
	c, ok := energy.ByID("eroeffnungsfeier")
	if !ok {
		t.Fatal("card not found")
	}
	if c.CriticalLimit() != 8 {
		t.Fatalf("limit = %d, want 8", c.CriticalLimit())
	}
}

func TestCriticalLimitDefaultCard(t *testing.T) {
	if got := energy.DefaultCard().CriticalLimit(); got != 8 {
		t.Fatalf("limit = %d, want 8", got)
	}
}
