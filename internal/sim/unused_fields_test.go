package sim_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestUnusedFieldsCountsNeverHitPlaceables(t *testing.T) {
	s := board.NewEmpty()
	// On the east shot path from the emitter.
	s.Tiles[1][2] = field.NewTile(field.GasBoiler, 0, 0)
	// Off-path placeable that should stay unused.
	s.Tiles[1][1] = field.NewTile(field.Mirror, hex.RotE, 0)

	cfg := sim.DefaultConfig()
	cfg.StartDir = hex.RotE.TravelDir()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipHeat,
		Pos:  hex.Coord{Q: hex.EmitterCol, R: hex.EmitterRow},
		Dir:  hex.RotE.TravelDir(),
	}}

	res := sim.Run(s, rand.New(rand.NewSource(1)), cfg)
	if res.UnusedFields < 1 {
		t.Fatalf("UnusedFields = %d, want at least the off-path mirror", res.UnusedFields)
	}
}

func TestUnusedFieldsZeroWhenAllHit(t *testing.T) {
	s := board.NewEmpty()
	s.Tiles[1][2] = field.NewTile(field.GasBoiler, 0, 0)

	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipHeat,
		Pos:  hex.Coord{Q: 0, R: 2},
		Dir:  hex.RotE.TravelDir(),
	}}

	res := sim.Run(s, rand.New(rand.NewSource(2)), cfg)
	if res.UnusedFields != 0 {
		t.Fatalf("UnusedFields = %d, want 0 when only field is hit", res.UnusedFields)
	}
}
