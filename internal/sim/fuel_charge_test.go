package sim_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestGasBoilerEachReactionReducesCharge(t *testing.T) {
	pos := hex.Coord{Q: 1, R: 2}
	s := board.NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.NewTile(field.GasBoiler, 0, 0)

	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipHeat,
		Pos:  hex.Coord{Q: 0, R: 2},
		Dir:  hex.RotE.TravelDir(),
	}}

	prev := s.Tiles[pos.Q][pos.R].Charge
	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(7)), cfg)
	for _, snap := range snaps {
		if snap.Event != "Feldreaktion" {
			continue
		}
		tile := snap.Board.Tiles[pos.Q][pos.R]
		if tile.Type != field.GasBoiler {
			continue
		}
		if tile.Charge >= prev {
			t.Fatalf("charge stuck at %d across gas reactions", tile.Charge)
		}
		prev = tile.Charge
	}
}

func TestGasBoilerLastChargeBurnsOut(t *testing.T) {
	pos := hex.Coord{Q: 1, R: 2}
	s := board.NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.NewTile(field.GasBoiler, 0, 0)
	s.Tiles[pos.Q][pos.R].Charge = 1

	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipHeat,
		Pos:  hex.Coord{Q: 0, R: 2},
		Dir:  hex.RotE.TravelDir(),
	}}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(1)), cfg)
	for _, snap := range snaps {
		if snap.Event != "Feldreaktion" {
			continue
		}
		tile := snap.Board.Tiles[pos.Q][pos.R]
		if tile.Type != field.GasBoiler {
			continue
		}
		if tile.Charge != 0 {
			t.Fatalf("expected charge depleted, got %d", tile.Charge)
		}
		if !tile.BurnedOut {
			t.Fatal("expected gas boiler burned out after last charge spent")
		}
		return
	}
	t.Fatal("expected gas boiler reaction")
}
