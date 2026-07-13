package finance_test

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/finance"
)

func TestSparmaßnahmenBudget(t *testing.T) {
	c, ok := finance.ByID("sparmassnahmen")
	if !ok {
		t.Fatal("card not found")
	}
	if c.ReactorBudget != 2 || c.GridBudget != 2 {
		t.Fatalf("budget = %d/%d, want 2/2", c.ReactorBudget, c.GridBudget)
	}
}

func TestUmJedenPreisCriticalLimit(t *testing.T) {
	c, ok := finance.ByID("um-jeden-preis")
	if !ok {
		t.Fatal("card not found")
	}
	if c.CriticalLimit() != 10 {
		t.Fatalf("limit = %d, want 10", c.CriticalLimit())
	}
}
