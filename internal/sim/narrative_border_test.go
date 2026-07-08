package sim_test

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/energy"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestBorderDemandNarrativeNamesZone(t *testing.T) {
	cfg := sim.DefaultConfig()
	cfg.EnergyCard = energy.Card{}
	cfg.ShiftDemands = board.ShiftDemands{Residential: 1}
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  hex.Coord{Q: 8, R: 0},
		Dir:  hex.RotE.TravelDir(),
	}}

	_, snaps := sim.RunTrace(board.NewEmpty(), rand.New(rand.NewSource(5)), cfg)
	for _, snap := range snaps {
		if !strings.HasPrefix(snap.Event, "Rand-Bedarf ") {
			continue
		}
		if snap.Event != "Rand-Bedarf Wohnviertel erfuellt" {
			t.Fatalf("event = %q", snap.Event)
		}
		if !strings.Contains(snap.Narrative, "Rand-Bedarf Wohnviertel") {
			t.Fatalf("narrative = %q", snap.Narrative)
		}
		return
	}
	t.Fatal("expected border demand event")
}
