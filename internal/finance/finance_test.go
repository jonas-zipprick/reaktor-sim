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
	if c.ReactorBudget != 3 || c.GridBudget != 2 {
		t.Fatalf("budget = %d/%d, want 3/2", c.ReactorBudget, c.GridBudget)
	}
}

func TestSchwerindustrieBudget(t *testing.T) {
	c, ok := finance.ByID("schwerindustrie")
	if !ok {
		t.Fatal("card not found")
	}
	if c.ReactorBudget != 3 || c.GridBudget != 3 {
		t.Fatalf("budget = %d/%d, want 3/3", c.ReactorBudget, c.GridBudget)
	}
}

func TestWettruestenBudget(t *testing.T) {
	c, ok := finance.ByID("wettruesten")
	if !ok {
		t.Fatal("card not found")
	}
	if c.ReactorBudget != 4 || c.GridBudget != 2 {
		t.Fatalf("budget = %d/%d, want 4/2", c.ReactorBudget, c.GridBudget)
	}
	if c.RepairsAllowed() {
		t.Fatal("repairs should be disabled")
	}
}
