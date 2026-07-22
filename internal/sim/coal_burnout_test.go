package sim_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestCoalChamberStartsWithEightCharge(t *testing.T) {
	tile := field.NewTile(field.CoalChamber, 0, 0)
	if tile.Charge != 8 {
		t.Fatalf("coal charge = %d, want 8", tile.Charge)
	}
}

func TestGasBoilerStartsWithFiveCharge(t *testing.T) {
	tile := field.NewTile(field.GasBoiler, 0, 0)
	if tile.Charge != 5 {
		t.Fatalf("gas charge = %d, want 5", tile.Charge)
	}
}

func TestBurnedGasRedirectsHeat(t *testing.T) {
	pos := hex.Coord{Q: 1, R: 2}
	s := board.NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.Tile{Type: field.GasBoiler, BurnedOut: true}

	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipHeat,
		Pos:  hex.Coord{Q: 0, R: 2},
		Dir:  hex.RotE.TravelDir(),
	}}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(5)), cfg)
	for _, snap := range snaps {
		if snap.Event != "Weiterleitung" {
			continue
		}
		for _, c := range snap.Queue {
			if c.Pos == pos && c.Type == sim.ChipHeat {
				return
			}
		}
	}
	t.Fatal("expected heat redirected from burned gas boiler")
}

func TestBurnedCoalDestroysHeat(t *testing.T) {
	pos := hex.Coord{Q: 1, R: 2}
	s := board.NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.Tile{Type: field.CoalChamber, BurnedOut: true}

	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipHeat,
		Pos:  hex.Coord{Q: 0, R: 2},
		Dir:  hex.RotE.TravelDir(),
	}}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(5)), cfg)
	for _, snap := range snaps {
		if snap.Event == "Ausgebrannt" {
			return
		}
	}
	t.Fatal("expected heat to vanish on burned coal")
}

func TestBurnedCoalStillDestroysNeutron(t *testing.T) {
	pos := hex.Coord{Q: 1, R: 2}
	s := board.NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.Tile{Type: field.CoalChamber, BurnedOut: true}

	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipNeutron,
		Pos:  hex.Coord{Q: 0, R: 2},
		Dir:  hex.RotE.TravelDir(),
	}}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(5)), cfg)
	for _, snap := range snaps {
		if snap.Event == "Ausgebrannt" {
			return
		}
	}
	t.Fatal("expected neutron to vanish on burned coal")
}
