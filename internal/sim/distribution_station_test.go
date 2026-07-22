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

func TestDistributionStationStoresFirstVoltage(t *testing.T) {
	pos := hex.Coord{Q: 6, R: 1}
	s := board.NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.NewTile(field.DistributionStation, hex.RotE, 0)

	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  hex.Coord{Q: 5, R: 1},
		Dir:  hex.RotE.TravelDir(),
	}}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(1)), cfg)
	for _, snap := range snaps {
		if snap.Event != "Feldreaktion" {
			continue
		}
		tile := snap.Board.Tiles[pos.Q][pos.R]
		if tile.Charge != 1 {
			t.Fatalf("charge = %d, want 1 after first voltage", tile.Charge)
		}
		if !strings.Contains(snap.Narrative, "gespeichert") {
			t.Fatalf("expected storage narrative, got %q", snap.Narrative)
		}
		return
	}
	t.Fatal("expected distribution station storage reaction")
}

func TestDistributionStationSplitsToEdges0And3(t *testing.T) {
	pos := hex.Coord{Q: 6, R: 1}
	s := board.NewEmpty()
	tile := field.NewTile(field.DistributionStation, hex.RotE, 0)
	tile.Charge = 1
	s.Tiles[pos.Q][pos.R] = tile

	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  hex.Coord{Q: 5, R: 1},
		Dir:  hex.RotE.TravelDir(),
	}}

	want0 := hex.RotE.TravelDir()
	want3 := hex.RotW.TravelDir()
	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(2)), cfg)
	for _, snap := range snaps {
		if snap.Event != "Feldreaktion" {
			continue
		}
		boardTile := snap.Board.Tiles[pos.Q][pos.R]
		if boardTile.Charge != 0 {
			t.Fatalf("charge = %d after release, want 0", boardTile.Charge)
		}
		seen := map[int]int{}
		for _, c := range snap.Queue {
			if c.Pos != pos || c.Type != sim.ChipVoltage {
				continue
			}
			seen[c.Dir]++
		}
		if seen[want0] != 1 || seen[want3] != 1 || len(seen) != 2 {
			t.Fatalf("dirs = %v, want one %s and one %s", seen, hex.DisplayDirName(want0), hex.DisplayDirName(want3))
		}
		if boardTile.Orientation != hex.RotSE {
			t.Fatalf("orientation after fire = %s, want SE (rotated clockwise from E)", boardTile.Orientation)
		}
		return
	}
	t.Fatal("expected distribution station release")
}

func TestDistributionStationHeatPasses(t *testing.T) {
	pos := hex.Coord{Q: 6, R: 1}
	s := board.NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.NewTile(field.DistributionStation, hex.RotE, 0)

	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipHeat,
		Pos:  hex.Coord{Q: 5, R: 1},
		Dir:  hex.RotE.TravelDir(),
	}}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(3)), cfg)
	for _, snap := range snaps {
		if snap.Event != "Feldreaktion" {
			continue
		}
		tile := snap.Board.Tiles[pos.Q][pos.R]
		if tile.Charge != 0 {
			t.Fatalf("heat should not charge station, charge=%d", tile.Charge)
		}
		return
	}
	t.Fatal("expected heat pass-through reaction")
}

func TestDistributionStationHasRotation(t *testing.T) {
	if !field.HasRotation(field.DistributionStation) {
		t.Fatal("distribution station orientation must affect simulation")
	}
	info := field.Catalog[field.DistributionStation]
	if info.Cost != 2 || info.MaxCharge != 2 || info.Sector != "grid" {
		t.Fatalf("catalog = %+v, want cost 2 max 2 sector grid", info)
	}
}
