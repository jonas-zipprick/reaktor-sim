package main

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/finance"
)

func TestRepairBudgetsForRunUsesLeftoverOnly(t *testing.T) {
	state := board.NewEmpty()
	state.Damage = [4]int{5, 0, 0, 0}
	state.EmitterDamage = 4

	fin, _ := finance.ByID("schwerindustrie")
	reactor, grid := repairBudgetsForRun(state, fin, -1, 2, 3)
	if reactor != 2 {
		t.Fatalf("reactor budget = %d, want 2", reactor)
	}
	if grid != 3 {
		t.Fatalf("grid budget = %d, want 3", grid)
	}
	reactor, grid = repairBudgetsForRun(state, fin, -1, 0, 0)
	if reactor != 0 || grid != 0 {
		t.Fatalf("no leftover = reactor %d grid %d, want 0/0", reactor, grid)
	}
	reactor, grid = repairBudgetsForRun(state, fin, -1, 10, 10)
	if reactor != 4 {
		t.Fatalf("reactor capped by damage = %d, want 4", reactor)
	}
	if grid != 5 {
		t.Fatalf("grid capped by damage = %d, want 5", grid)
	}
}

func TestRepairBudgetDeniedByFinanceCard(t *testing.T) {
	state := board.NewEmpty()
	state.Damage = [4]int{2, 0, 0, 0}
	state.EmitterDamage = 2
	fin, _ := finance.ByID("wettruesten")
	reactor, grid := repairBudgetsForRun(state, fin, -1, 5, 5)
	if reactor != 0 || grid != 0 {
		t.Fatalf("wettruesten budgets = %d/%d, want 0/0", reactor, grid)
	}
}
