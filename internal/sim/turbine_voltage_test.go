package sim_test

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestVoltageAtTurbineConsumesPlantDemand(t *testing.T) {
	cfg := testCfg()
	cfg.ShiftDemands = board.ShiftDemands{Plant: 1}
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  hex.Coord{Q: 5, R: 2},
		Dir:  hex.RotW.TravelDir(),
	}}

	res := sim.Run(board.NewEmpty(), rand.New(rand.NewSource(1)), cfg)
	if res.ZoneDeliveries[board.ZonePlant] != 1 {
		t.Fatalf("plant deliveries = %d, want 1", res.ZoneDeliveries[board.ZonePlant])
	}

	_, snaps := sim.RunTrace(board.NewEmpty(), rand.New(rand.NewSource(1)), cfg)
	found := false
	for _, snap := range snaps {
		if snap.Event == board.BorderDemandEvent(board.ZonePlant) {
			found = true
			if !strings.Contains(snap.Narrative, "Reaktoreigenbedarf") {
				t.Fatalf("narrative = %q", snap.Narrative)
			}
		}
		if snap.Event == "Spannungs-Spike" {
			t.Fatal("expected plant consumption, not spike")
		}
	}
	if !found {
		t.Fatal("expected plant demand event at turbine")
	}
}
