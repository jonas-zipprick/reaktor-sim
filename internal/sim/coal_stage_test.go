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

func TestCoalChamberStagesFirstHeat(t *testing.T) {
	pos := hex.Coord{Q: 1, R: 2}
	s := board.NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.NewTile(field.CoalChamber, 0, 0)
	startCharge := s.Tiles[pos.Q][pos.R].Charge

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
		if tile.UnboundHeat != 1 {
			t.Fatalf("unbound heat = %d, want 1 after first hit", tile.UnboundHeat)
		}
		if tile.Charge != startCharge-1 {
			t.Fatalf("charge = %d, want %d", tile.Charge, startCharge-1)
		}
		if !strings.Contains(snap.Narrative, "ungebundene Ladung") {
			t.Fatalf("first hit should stage, got %q", snap.Narrative)
		}
		return
	}
	t.Fatal("expected coal staging reaction")
}

func TestCoalChamberFiresFourOnSecondHeat(t *testing.T) {
	pos := hex.Coord{Q: 1, R: 2}
	s := board.NewEmpty()
	tile := field.NewTile(field.CoalChamber, 0, 0)
	tile.Charge = 4
	tile.UnboundHeat = 1
	s.Tiles[pos.Q][pos.R] = tile

	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipHeat,
		Pos:  hex.Coord{Q: 0, R: 2},
		Dir:  hex.RotE.TravelDir(),
	}}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(3)), cfg)
	for _, snap := range snaps {
		if snap.Event != "Feldreaktion" {
			continue
		}
		boardTile := snap.Board.Tiles[pos.Q][pos.R]
		if boardTile.UnboundHeat != 0 {
			t.Fatalf("unbound heat = %d after firing, want 0", boardTile.UnboundHeat)
		}
		if boardTile.Charge != 3 {
			t.Fatalf("charge = %d after second hit, want 3", boardTile.Charge)
		}
		heat := 0
		for _, c := range snap.Queue {
			if c.Pos == pos && c.Type == sim.ChipHeat {
				heat++
			}
		}
		if heat != 4 {
			t.Fatalf("emitted heat at coal = %d, want 4", heat)
		}
		return
	}
	t.Fatal("expected coal firing reaction")
}

func TestCoalChamberClearsUnboundAtShiftEnd(t *testing.T) {
	pos := hex.Coord{Q: 1, R: 2}
	s := board.NewEmpty()
	tile := field.NewTile(field.CoalChamber, 0, 0)
	tile.UnboundHeat = 1
	s.Tiles[pos.Q][pos.R] = tile

	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{}
	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(1)), cfg)
	if len(snaps) == 0 {
		t.Fatal("expected snapshots")
	}
	last := snaps[len(snaps)-1]
	if last.Board.Tiles[pos.Q][pos.R].UnboundHeat != 0 {
		t.Fatalf("unbound heat after shift end = %d, want 0", last.Board.Tiles[pos.Q][pos.R].UnboundHeat)
	}
}
