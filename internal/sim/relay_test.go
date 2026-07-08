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

func TestRelayOrientation5PassesVoltageEast(t *testing.T) {
	s := board.NewEmpty()
	pos := hex.Coord{Q: 6, R: 1}
	s.Tiles[pos.Q][pos.R] = field.NewTile(field.Relay, hex.RotW, 0)

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
		if strings.Contains(snap.Narrative, "fliegt durch das Relais") {
			return
		}
		if strings.Contains(snap.Narrative, "Relais") && strings.Contains(snap.Narrative, "Richtung W") {
			t.Fatalf("voltage E should pass through Re5, got %q", snap.Narrative)
		}
	}
	t.Fatal("expected relay pass-through narrative")
}

func TestRelayOrientation0DeflectsWestToSouthwest(t *testing.T) {
	s := board.NewEmpty()
	pos := hex.Coord{Q: 6, R: 1}
	s.Tiles[pos.Q][pos.R] = field.NewTile(field.Relay, hex.RotNW, 0)

	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  hex.Coord{Q: 5, R: 1},
		Dir:  hex.RotE.TravelDir(),
	}}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(1)), cfg)
	for _, snap := range snaps {
		if snap.Event == "Feldreaktion" && strings.Contains(snap.Narrative, "Richtung SW") {
			return
		}
	}
	t.Fatal("expected relay to deflect W to SW at orientation 0")
}
