package sim_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestReactorRepairBudgetReducesEmitterDamage(t *testing.T) {
	state := board.NewEmpty()
	state.EmitterDamage = 3

	cfg := sim.DefaultConfig()
	cfg.ReactorRepairBudget = 2
	cfg.ShiftDemands = board.DefaultShiftDemands()

	res, snaps := sim.RunTrace(state, rand.New(rand.NewSource(1)), cfg)
	if res.ReactorRepairSpent != 2 {
		t.Fatalf("reactor repair spent = %d, want 2", res.ReactorRepairSpent)
	}
	if len(snaps) == 0 {
		t.Fatal("no snapshots")
	}
	if got := snaps[0].Board.EmitterDamage; got != 1 {
		t.Fatalf("emitter damage after repair = %d, want 1", got)
	}
}

func TestEmitterDamageCountsTowardCriticalMass(t *testing.T) {
	s := board.NewEmpty()
	s.EmitterDamage = 8
	cfg := sim.DefaultConfig()
	cfg.ShiftDemands = board.ShiftDemands{}
	cfg.InitialChips = []sim.Chip{}

	res := sim.Run(s, rand.New(rand.NewSource(1)), cfg)
	if !res.CriticalFailure {
		t.Fatal("8 emitter damage should exceed critical limit of 7 on player 1")
	}
	if !res.CriticalP1 {
		t.Fatal("expected critical failure on player 1 side")
	}
}
