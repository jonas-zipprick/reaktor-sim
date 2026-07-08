package sim_test

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/energy"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestVoltageOnHVCascadeTriggersReactionNotDemand(t *testing.T) {
	s := board.NewEmpty()
	pos := hex.Coord{Q: 8, R: 1}
	s.Tiles[pos.Q][pos.R] = field.NewTile(field.HVCascade, 0, 0)

	cfg := sim.DefaultConfig()
	cfg.EnergyCard = energy.Card{}
	cfg.ShiftDemands = board.ShiftDemands{Residential: 2}
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  hex.Coord{Q: 7, R: 1},
		Dir:  hex.RotE.TravelDir(),
	}}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(1)), cfg)
	var first string
	for _, snap := range snaps {
		if snap.Event == "Start" {
			continue
		}
		first = snap.Event
		break
	}
	if first != "Feldreaktion" {
		t.Fatalf("first step = %q, want Feldreaktion", first)
	}
	for _, snap := range snaps {
		if snap.Event != "Feldreaktion" {
			continue
		}
		tile := snap.Board.Tiles[pos.Q][pos.R]
		if tile.Charge != 5 {
			t.Fatalf("HV charge = %d, want 5 after one hit", tile.Charge)
		}
		return
	}
	t.Fatal("missing Feldreaktion snapshot")
}

func TestVoltageOnWiredBorderFieldDoesNotConsumeDemand(t *testing.T) {
	cfg := sim.DefaultConfig()
	cfg.EnergyCard = energy.Card{}
	cfg.ShiftDemands = board.ShiftDemands{Residential: 1}
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  hex.Coord{Q: 7, R: 1},
		Dir:  hex.RotE.TravelDir(),
	}}

	_, snaps := sim.RunTrace(board.NewEmpty(), rand.New(rand.NewSource(1)), cfg)
	for _, snap := range snaps {
		if snap.Event == "Start" {
			continue
		}
		if strings.HasPrefix(snap.Event, "Rand-Bedarf ") {
			t.Fatal("demand satisfied without hitting outer wall first")
		}
		if snap.Event == "Leeres Feld" {
			break
		}
	}
}
