package sim_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

// Regression for run2: voltage at (5,2) heading SE hits the Bahn wall (b=0),
// so it must add rail damage instead of consuming Wohnviertel demand.
func TestVoltageSEFromBottomRowDamagesWhenRailEmpty(t *testing.T) {
	cfg := testCfg()
	cfg.ShiftDemands = board.ShiftDemands{Industry: 1, Residential: 1, Plant: 1}
	cfg.InitialHeat = 0
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  hex.Coord{Q: 5, R: 2},
		Dir:  hex.RotSE.TravelDir(),
	}}

	_, snaps := sim.RunTrace(board.NewEmpty(), rand.New(rand.NewSource(1)), cfg)
	for _, snap := range snaps {
		if snap.Event == board.BorderDemandEvent(board.ZoneResidential) {
			t.Fatalf("unexpected residential delivery: %s", snap.Narrative)
		}
		if snap.Event == board.BorderDamageEvent(board.ZoneRail) {
			return
		}
	}
	t.Fatal("expected rail damage when SE hits Bahn wall with b=0")
}

func TestVoltageSEFromBottomRowConsumesRail(t *testing.T) {
	cfg := testCfg()
	cfg.ShiftDemands = board.ShiftDemands{Rail: 1, Residential: 1}
	cfg.InitialHeat = 0
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  hex.Coord{Q: 5, R: 2},
		Dir:  hex.RotSE.TravelDir(),
	}}

	res := sim.Run(board.NewEmpty(), rand.New(rand.NewSource(1)), cfg)
	if res.ZoneDeliveries[board.ZoneRail] != 1 {
		t.Fatalf("rail deliveries = %d, want 1", res.ZoneDeliveries[board.ZoneRail])
	}
	if res.ZoneDeliveries[board.ZoneResidential] != 0 {
		t.Fatalf("residential deliveries = %d, want 0", res.ZoneDeliveries[board.ZoneResidential])
	}
}
