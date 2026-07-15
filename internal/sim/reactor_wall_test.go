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

// Voltage fired west from a player-2 field must never cross the reactor wall
// into player 1 (e.g. through a mirror). It vanishes and consumes plant demand.
func TestVoltageIntoReactorWallConsumesPlantDemand(t *testing.T) {
	cfg := testCfg()
	cfg.ShiftDemands = board.ShiftDemands{Plant: 1}
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  hex.Coord{Q: 5, R: 1},
		Dir:  hex.RotW.TravelDir(),
	}}

	s := board.NewEmpty()
	// A mirror on the player-1 cell just past the wall reproduces the bug where
	// voltage used to "passieren" the mirror into the reactor half.
	s.Tiles[3][1] = field.NewTile(field.Mirror, hex.RotNW, 0)

	rng := rand.New(rand.NewSource(1))
	res, snaps := sim.RunTrace(s, rng, cfg)

	if res.ZoneDeliveries[board.ZonePlant] != 1 {
		t.Fatalf("plant deliveries = %d, want 1", res.ZoneDeliveries[board.ZonePlant])
	}
	for _, snap := range snaps {
		if strings.Contains(snap.Narrative, "Spiegel") {
			t.Fatalf("voltage must never reach the reactor-side mirror, got %q", snap.Narrative)
		}
		if snap.Active != nil && snap.Active.Type == sim.ChipVoltage && snap.Active.Pos.IsPlayer1() {
			t.Fatalf("voltage entered player 1 at %+v", snap.Active.Pos)
		}
	}
}

// With no plant demand left, voltage hitting the reactor wall damages the plant zone.
func TestVoltageIntoReactorWallDamagesWhenNoDemand(t *testing.T) {
	cfg := testCfg()
	cfg.ShiftDemands = board.ShiftDemands{}
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  hex.Coord{Q: 5, R: 3},
		Dir:  hex.RotW.TravelDir(),
	}}

	rng := rand.New(rand.NewSource(2))
	res, snaps := sim.RunTrace(board.NewEmpty(), rng, cfg)

	if res.EndDamage[board.ZonePlant] != 1 {
		t.Fatalf("plant damage = %d, want 1", res.EndDamage[board.ZonePlant])
	}
	found := false
	for _, snap := range snaps {
		if snap.Event == board.BorderDamageEvent(board.ZonePlant) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected reactor-wall damage event")
	}
}
