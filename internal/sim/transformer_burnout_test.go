package sim_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestBurnedTransformerRedirectsVoltage(t *testing.T) {
	pos := hex.Coord{Q: 6, R: 1}
	s := board.NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.Tile{Type: field.Transformer, BurnedOut: true}

	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  hex.Coord{Q: 5, R: 1},
		Dir:  hex.RotE.TravelDir(),
	}}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(3)), cfg)
	for _, snap := range snaps {
		if snap.Event != "Weiterleitung" {
			continue
		}
		for _, c := range snap.Queue {
			if c.Pos == pos && c.Type == sim.ChipVoltage {
				return
			}
		}
	}
	t.Fatal("expected voltage redirected from burned transformer")
}

func TestBurnedTransformerRedirectsHeat(t *testing.T) {
	pos := hex.Coord{Q: 6, R: 1}
	s := board.NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.Tile{Type: field.Transformer, BurnedOut: true}

	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipHeat,
		Pos:  hex.Coord{Q: 5, R: 1},
		Dir:  hex.RotE.TravelDir(),
	}}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(3)), cfg)
	for _, snap := range snaps {
		if snap.Event == "Weiterleitung" {
			return
		}
	}
	t.Fatal("expected heat redirected from burned transformer")
}