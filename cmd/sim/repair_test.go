package main

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/finance"
)

func TestRepairBudgetForRunUsesLeftoverOnly(t *testing.T) {
	state := board.NewEmpty()
	state.Damage = [4]int{5, 0, 0, 0}

	fin, _ := finance.ByID("schwerindustrie")
	if got := repairBudgetForRun(state, fin, -1, 3); got != 3 {
		t.Fatalf("budget = %d, want 3", got)
	}
	if got := repairBudgetForRun(state, fin, -1, 0); got != 0 {
		t.Fatalf("no leftover = %d, want 0", got)
	}
	if got := repairBudgetForRun(state, fin, -1, 10); got != 5 {
		t.Fatalf("capped by damage = %d, want 5", got)
	}
}

func TestRepairBudgetDeniedByFinanceCard(t *testing.T) {
	state := board.NewEmpty()
	state.Damage = [4]int{2, 0, 0, 0}
	fin, _ := finance.ByID("wettruesten")
	if got := repairBudgetForRun(state, fin, -1, 5); got != 0 {
		t.Fatalf("wettruesten budget = %d, want 0", got)
	}
}
