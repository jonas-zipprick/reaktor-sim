package sim_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestShiftEndRemovesBurnedFields(t *testing.T) {
	pos := hex.Coord{Q: 1, R: 1}
	s := board.NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.Tile{Type: field.CoalChamber, BurnedOut: true}

	cfg := sim.DefaultConfig()
	cfg.StartDir = hex.RotW.TravelDir()

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(3)), cfg)
	for _, snap := range snaps {
		if snap.Event != "Ende" {
			continue
		}
		tile := snap.Board.Tiles[pos.Q][pos.R]
		if tile.Type != field.Empty {
			t.Fatalf("expected burned coal removed at shift end, got %+v", tile)
		}
		return
	}
	t.Fatal("expected shift end event")
}

func TestApplyShiftCarryClearsBurnedSlot(t *testing.T) {
	pos := hex.Coord{Q: 1, R: 1}
	s := board.NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.Tile{Type: field.CoalChamber, BurnedOut: true}

	cfg := sim.DefaultConfig()
	cfg.StartDir = hex.RotW.TravelDir()

	sim.ApplyShiftCarry(s, 11, 1, cfg)
	tile := s.Tiles[pos.Q][pos.R]
	if tile.Type != field.Empty {
		t.Fatalf("carry board tile = %+v, want empty slot after burnout", tile)
	}
}
