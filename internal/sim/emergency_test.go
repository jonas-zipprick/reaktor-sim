package sim_test

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestEmergencyGeneratorRemovedOnVoltageHit(t *testing.T) {
	pos := hex.Coord{Q: 6, R: 1}
	s := board.NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.NewTile(field.EmergencyGenerator, 0, 0)

	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{
		{Type: sim.ChipVoltage, Pos: hex.Coord{Q: 5, R: 1}, Dir: hex.RotE.TravelDir()},
		{Type: sim.ChipVoltage, Pos: pos, Dir: hex.RotNE.TravelDir()},
	}
	cfg.Trace = true

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(1)), cfg)

	destroyed := false
	for _, snap := range snaps {
		if snap.Event == "Notgenerator zerstoert" {
			destroyed = true
			if !strings.Contains(snap.Narrative, "den Notgenerator") {
				t.Fatalf("narrative = %q, want Notgenerator", snap.Narrative)
			}
			if strings.Contains(snap.Narrative, "ein Feld") {
				t.Fatalf("narrative = %q, should not say ein Feld", snap.Narrative)
			}
		}
		if !destroyed {
			continue
		}
		if snap.Board.Tiles[pos.Q][pos.R].Type != field.Empty {
			t.Fatalf("step %d: tile not empty after destruction", snap.Step)
		}
		for _, chip := range snap.Queue {
			if chip.Pos == pos {
				t.Fatalf("chip still on removed generator at step %d: %+v", snap.Step, chip)
			}
		}
	}
	if !destroyed {
		t.Fatal("emergency generator should be removed from board")
	}
}

func TestEmergencyGeneratorBoundChargeRemoved(t *testing.T) {
	pos := hex.Coord{Q: 6, R: 1}
	s := board.NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.NewTile(field.EmergencyGenerator, 0, 0)
	if s.Tiles[pos.Q][pos.R].Charge != 2 {
		t.Fatalf("setup charge = %d, want 2", s.Tiles[pos.Q][pos.R].Charge)
	}

	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  hex.Coord{Q: 5, R: 1},
		Dir:  hex.RotE.TravelDir(),
	}}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(2)), cfg)
	for i := len(snaps) - 1; i >= 0; i-- {
		if snaps[i].Event == "Notgenerator zerstoert" {
			tile := snaps[i].Board.Tiles[pos.Q][pos.R]
			if tile.Type != field.Empty {
				t.Fatalf("tile type = %v, want empty", tile.Type)
			}
			if tile.Charge != 0 {
				t.Fatalf("bound charge = %d after removal", tile.Charge)
			}
			return
		}
	}
	t.Fatal("expected Notgenerator zerstoert event")
}
