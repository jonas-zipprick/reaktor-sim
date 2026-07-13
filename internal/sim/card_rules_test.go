package sim_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/energy"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestSchturmowFiresTwoHeatChips(t *testing.T) {
	card, _ := energy.ByID("schturmowschtschina")
	cfg := sim.DefaultConfig()
	cfg.EnergyCard = card
	cfg.InitialChips = nil

	s := board.NewEmpty()
	s.ApplyDemands(board.ShiftDemands{Plant: 1})
	chips := sim.EmitterChips(s, cfg, rand.New(rand.NewSource(1)))
	if len(chips) != 2 {
		t.Fatalf("chips = %d, want 2", len(chips))
	}
	for i, c := range chips {
		if c.Type != sim.ChipHeat {
			t.Fatalf("chip %d type = %v, want heat", i, c.Type)
		}
		if i > 0 && c.Dir != chips[0].Dir {
			t.Fatalf("chip %d dir = %d, want %d", i, c.Dir, chips[0].Dir)
		}
	}
}
