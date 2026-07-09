package sim_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestRepairBudgetReducesStartDamage(t *testing.T) {
	state := board.NewEmpty()
	state.Damage = [4]int{3, 0, 0, 0}

	cfg := sim.DefaultConfig()
	cfg.RepairBudget = 1
	cfg.ShiftDemands = board.DefaultShiftDemands()

	res, snaps := sim.RunTrace(state, rand.New(rand.NewSource(1)), cfg)
	if res.RepairSpent != 1 {
		t.Fatalf("repair spent = %d, want 1", res.RepairSpent)
	}
	if len(snaps) == 0 {
		t.Fatal("no snapshots")
	}
	if got := snaps[0].Board.TotalDamage(board.ZoneIndustry); got != 2 {
		t.Fatalf("industry damage after repair = %d, want 2", got)
	}
}
