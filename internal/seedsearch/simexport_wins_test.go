package seedsearch_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/energy"
	"github.com/jonas/reaktor-sim/internal/finance"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestSimExportConfigNeedsEnergyCardForWins(t *testing.T) {
	card, _ := energy.ByID("schturmowschtschina")
	fin, _ := finance.ByID("schwerindustrie")
	demands := card.ShiftDemands(2)

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

	const runs = 200
	const seed int64 = 616

	var state *board.State
	var brokenResults, fullResults []sim.Result
	for boardSeed := int64(1); boardSeed <= 200; boardSeed++ {
		state = board.Random(rand.New(rand.NewSource(boardSeed)), 0)
		brokenResults = sim.RunMonteCarlo(state, runs, seed, broken)
		fullResults = sim.RunMonteCarlo(state, runs, seed, full)
		if countWins(fullResults) > countWins(brokenResults) && len(sim.WinTraceRunIndices(fullResults, 1)) >= 1 {
			return
		}
	}
	t.Fatalf("no board in seed range 1-200 where full config outperforms broken (last: full=%d broken=%d)",
		countWins(fullResults), countWins(brokenResults))
}

func countWins(results []sim.Result) int {
	n := 0
	for _, r := range results {
		if r.AllDemandsMet {
			n++
		}
	}
	return n
}
