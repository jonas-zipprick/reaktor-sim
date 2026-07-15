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

// Voltage fired east from the reactor-wall cell must reflect and never enter
// player 2 through a mirror on the far side of the wall.
func TestVoltageIntoReactorWallReflects(t *testing.T) {
	cfg := testCfg()
	cfg.ShiftDemands = board.ShiftDemands{Plant: 1}
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  hex.Coord{Q: 3, R: 1},
		Dir:  hex.RotE.TravelDir(),
	}}

	s := board.NewEmpty()
	s.Tiles[4][1] = field.NewTile(field.Mirror, hex.RotNW, 0)

	rng := rand.New(rand.NewSource(1))
	res, snaps := sim.RunTrace(s, rng, cfg)

	if res.ZoneDeliveries[board.ZonePlant] != 0 {
		t.Fatalf("plant deliveries = %d, want 0 (wall reflects, does not consume)", res.ZoneDeliveries[board.ZonePlant])
	}
	reflected := false
	for _, snap := range snaps {
		if snap.Event == "Spannungs-Reflektion" {
			reflected = true
		}
		if strings.Contains(snap.Narrative, "Spiegel") {
			t.Fatalf("voltage must not reach the player-2 mirror, got %q", snap.Narrative)
		}
		if snap.Active != nil && snap.Active.Type == sim.ChipVoltage && snap.Active.Pos.Q >= hex.Player2MinCol {
			t.Fatalf("voltage entered player 2 at %+v", snap.Active.Pos)
		}
	}
	if !reflected {
		t.Fatal("expected voltage reflection at reactor wall")
	}
}

func TestVoltageIntoReactorWallReflectsWhenNoDemand(t *testing.T) {
	cfg := testCfg()
	cfg.ShiftDemands = board.ShiftDemands{}
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  hex.Coord{Q: 3, R: 3},
		Dir:  hex.RotE.TravelDir(),
	}}

	_, snaps := sim.RunTrace(board.NewEmpty(), rand.New(rand.NewSource(2)), cfg)

	for _, snap := range snaps {
		if snap.Event == board.BorderDamageEvent(board.ZonePlant) {
			t.Fatal("reactor wall should reflect voltage, not damage plant zone")
		}
		if snap.Event == "Spannungs-Reflektion" {
			return
		}
	}
	t.Fatal("expected voltage reflection at reactor wall")
}
