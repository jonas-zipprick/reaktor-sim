package sim_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestRunEndsWhenQueueEmptyAndNoStorage(t *testing.T) {
	s := board.NewEmpty()
	s.Tiles[2][1] = field.NewTile(field.CoolingTower, 0, 0)

	cfg := sim.DefaultConfig()
	cfg.MixedEmitterTrigger = false
	cfg.InitialHeat = 0
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipHeat,
		Pos:  hex.Coord{Q: 1, R: 1},
		Dir:  hex.RotE.TravelDir(),
	}}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(1)), cfg)
	if len(snaps) == 0 {
		t.Fatal("no snapshots")
	}
	last := snaps[len(snaps)-1]
	if last.Event != "Ende" {
		t.Fatalf("last event = %q, want Ende", last.Event)
	}
	if last.QueueSize != 0 {
		t.Fatalf("final queue size = %d, want 0", last.QueueSize)
	}
}

func TestVoluntaryFireContinuesAfterQueueEmpty(t *testing.T) {
	s := board.NewEmpty()
	s.Tiles[2][1] = field.NewTile(field.CoolingTower, 0, 0)
	s.Tiles[6][1] = field.NewTile(field.CapacitorBank, 0, 0)
	s.Tiles[6][1].StoredVoltage = 5

	cfg := sim.DefaultConfig()
	cfg.MixedEmitterTrigger = false
	cfg.InitialHeat = 0
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipHeat,
		Pos:  hex.Coord{Q: 1, R: 1},
		Dir:  hex.RotE.TravelDir(),
	}}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(1)), cfg)
	foundVoluntary := false
	for _, snap := range snaps {
		if snap.Event == "Freiwilliger Schuss" {
			foundVoluntary = true
			break
		}
	}
	if !foundVoluntary {
		t.Fatal("expected voluntary fire from bound storage after queue emptied")
	}
}
