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

func TestPressureValveStoresFirstHeat(t *testing.T) {
	pos := hex.Coord{Q: 1, R: 2}
	s := board.NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.NewTile(field.PressureValve, hex.RotE, 0)

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
		if tile.Charge != 1 {
			t.Fatalf("charge = %d, want 1 after first heat", tile.Charge)
		}
		if !strings.Contains(snap.Narrative, "gespeichert") {
			t.Fatalf("expected storage narrative, got %q", snap.Narrative)
		}
		return
	}
	t.Fatal("expected pressure valve storage reaction")
}

func TestPressureValveFiresBothTowardOrientation(t *testing.T) {
	pos := hex.Coord{Q: 1, R: 2}
	s := board.NewEmpty()
	tile := field.NewTile(field.PressureValve, hex.RotE, 0)
	tile.Charge = 1
	s.Tiles[pos.Q][pos.R] = tile

	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipHeat,
		Pos:  hex.Coord{Q: 0, R: 2},
		Dir:  hex.RotE.TravelDir(),
	}}

	wantDir := hex.RotE.TravelDir()
	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(2)), cfg)
	for _, snap := range snaps {
		if snap.Event != "Feldreaktion" {
			continue
		}
		boardTile := snap.Board.Tiles[pos.Q][pos.R]
		if boardTile.Charge != 0 {
			t.Fatalf("charge = %d after release, want 0", boardTile.Charge)
		}
		heat := 0
		for _, c := range snap.Queue {
			if c.Pos != pos || c.Type != sim.ChipHeat {
				continue
			}
			heat++
			if c.Dir != wantDir {
				t.Fatalf("heat dir = %s, want E", hex.DisplayDirName(c.Dir))
			}
		}
		if heat != 2 {
			t.Fatalf("emitted heat = %d, want 2", heat)
		}
		if boardTile.Orientation != hex.RotSE {
			t.Fatalf("orientation after fire = %s, want SE (rotated clockwise from E)", boardTile.Orientation)
		}
		return
	}
	t.Fatal("expected pressure valve release")
}

func TestPressureValveNeutronsPass(t *testing.T) {
	pos := hex.Coord{Q: 1, R: 2}
	s := board.NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.NewTile(field.PressureValve, hex.RotE, 0)

	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipNeutron,
		Pos:  hex.Coord{Q: 0, R: 2},
		Dir:  hex.RotE.TravelDir(),
	}}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(3)), cfg)
	for _, snap := range snaps {
		if snap.Event != "Feldreaktion" {
			continue
		}
		tile := snap.Board.Tiles[pos.Q][pos.R]
		if tile.Charge != 0 {
			t.Fatalf("neutron should not charge valve, charge=%d", tile.Charge)
		}
		return
	}
	t.Fatal("expected neutron pass-through reaction")
}

func TestPressureValveHasRotation(t *testing.T) {
	if !field.HasRotation(field.PressureValve) {
		t.Fatal("pressure valve orientation must affect simulation")
	}
	info := field.Catalog[field.PressureValve]
	if info.Cost != 2 || info.MaxCharge != 2 || info.InitialCharge != 0 {
		t.Fatalf("catalog = %+v, want cost 2 max 2 initial 0", info)
	}
}
