package seedsearch_test

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/energy"
	"github.com/jonas/reaktor-sim/internal/finance"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestSimExportConfigNeedsEnergyCardForWins(t *testing.T) {
	fp := "b2_AAICAAAAAgQAAAAABAoABQAABgkCAAAADQkCAAAAEAwAAQAAEQMABAAAEgIAAAAAFgkAAAAA"
	state, err := board.FromFingerprint(fp)
	if err != nil {
		t.Fatal(err)
	}
	demands := board.ShiftDemands{Industry: 2, Residential: 1, Plant: 1}
	card, _ := energy.ByID("eroeffnungsfeier")
	fin, _ := finance.ByID("schwerindustrie")

	broken := sim.DefaultConfig()
	broken.EnergyCard = energy.Card{}
	broken.Shift = 2
	broken.ShiftDemands = demands

	full := sim.DefaultConfig()
	full.EnergyCard = card
	full.FinanceCard = fin
	full.CriticalLimit = fin.CriticalLimit()
	full.Shift = 2
	full.ShiftDemands = demands

	const runs = 100
	const seed int64 = 616
	brokenResults := sim.RunMonteCarlo(state, runs, seed, broken)
	fullResults := sim.RunMonteCarlo(state, runs, seed, full)

	if got := len(sim.WinTraceRunIndices(brokenResults, 1)); got != 0 {
		t.Fatalf("broken config win traces = %d, want 0", got)
	}
	if got := len(sim.WinTraceRunIndices(fullResults, 1)); got != 1 {
		t.Fatalf("full config win traces = %d, want 1", got)
	}
}
