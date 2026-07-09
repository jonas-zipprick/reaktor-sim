package sim_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestCapacitorExplodesWhenFull(t *testing.T) {
	pos := hex.Coord{Q: 6, R: 1}
	s := board.NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.NewTile(field.CapacitorBank, 0, 0)
	s.Tiles[pos.Q][pos.R].StoredVoltage = 5

	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  hex.Coord{Q: 5, R: 1},
		Dir:  hex.RotE.TravelDir(),
	}}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(1)), cfg)
	exploded := false
	for _, snap := range snaps {
		if snap.Event == "Kondensator explodiert" {
			exploded = true
			if snap.Board.Tiles[pos.Q][pos.R].Type != field.Empty {
				t.Fatalf("capacitor tile should be removed after explosion")
			}
		}
	}
	if !exploded {
		t.Fatal("expected capacitor explosion when receiving 6th voltage")
	}
}

func TestCapacitorClearsAtShiftEnd(t *testing.T) {
	pos := hex.Coord{Q: 6, R: 1}
	s := board.NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.NewTile(field.CapacitorBank, 0, 0)
	s.Tiles[pos.Q][pos.R].StoredVoltage = 3

	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(1)), cfg)
	if len(snaps) == 0 {
		t.Fatal("no snapshots")
	}
	last := snaps[len(snaps)-1]
	if last.Board.Tiles[pos.Q][pos.R].StoredVoltage != 0 {
		t.Fatalf("stored voltage after shift end = %d, want 0", last.Board.Tiles[pos.Q][pos.R].StoredVoltage)
	}
}

func TestLeadAccumulatorReleasesAtShiftStart(t *testing.T) {
	pos := hex.Coord{Q: 6, R: 1}
	s := board.NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.NewTile(field.LeadAccumulator, 0, 0)
	s.Tiles[pos.Q][pos.R].StoredVoltage = 2

	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(3)), cfg)
	if len(snaps) < 2 {
		t.Fatal("expected snapshots")
	}
	if snaps[0].Board.Tiles[pos.Q][pos.R].StoredVoltage != 1 {
		t.Fatalf("stored after shift start = %d, want 1", snaps[0].Board.Tiles[pos.Q][pos.R].StoredVoltage)
	}
	found := false
	for _, snap := range snaps {
		for _, chip := range snap.Queue {
			if chip.Pos == pos && chip.Type == sim.ChipVoltage {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("expected released voltage from lead accumulator at shift start")
	}
}
