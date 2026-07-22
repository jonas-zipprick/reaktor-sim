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
	if c.ReactorBudget != 3 || c.GridBudget != 3 {
		t.Fatalf("budget = %d/%d, want 3/3", c.ReactorBudget, c.GridBudget)
	}
	if c.AvailableFromMonth != 2 {
		t.Fatalf("AvailableFromMonth = %d, want 2", c.AvailableFromMonth)
	}
	if c.AvailableInMonth(1) {
		t.Fatal("sparmassnahmen must not be available in month 1")
	}
	if !c.AvailableInMonth(2) {
		t.Fatal("sparmassnahmen should be available in month 2")
	}
}

func TestSchwerindustrieBudget(t *testing.T) {
	c, ok := finance.ByID("schwerindustrie")
	if !ok {
		t.Fatal("card not found")
	}
	if c.ReactorBudget != 4 || c.GridBudget != 4 {
		t.Fatalf("budget = %d/%d, want 4/4", c.ReactorBudget, c.GridBudget)
	}
	if !c.AvailableInMonth(1) {
		t.Fatal("schwerindustrie should be available from month 1")
	}
}

func TestWettruestenBudget(t *testing.T) {
	c, ok := finance.ByID("wettruesten")
	if !ok {
		t.Fatal("card not found")
	}
	if c.ReactorBudget != 5 || c.GridBudget != 3 {
		t.Fatalf("budget = %d/%d, want 5/3", c.ReactorBudget, c.GridBudget)
	}
	if c.RepairsAllowed() {
		t.Fatal("repairs should be disabled")
	}
}
