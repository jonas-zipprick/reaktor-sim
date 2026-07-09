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

func testCfg() sim.Config {
	cfg := sim.DefaultConfig()
	cfg.EnergyCard = energy.Card{}
	return cfg
}

func TestHeatReflectsOffPlayer1OuterWall(t *testing.T) {
	cfg := testCfg()
	cfg.StartDir = 3 // west from emitter into outer wall

	rng := rand.New(rand.NewSource(1))
	_, snaps := sim.RunTrace(board.NewEmpty(), rng, cfg)
	found := false
	for _, snap := range snaps {
		if snap.Event == "Waerme-Reflektion" {
			found = true
			if !strings.Contains(snap.Narrative, "Richtung E") {
				t.Fatalf("west into wall should reflect east, got %q", snap.Narrative)
			}
			break
		}
	}
	if !found {
		t.Fatal("expected heat reflection event at player 1 outer wall")
	}
}

func TestHeatReflectionNWToSE(t *testing.T) {
	cfg := testCfg()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipHeat,
		Pos:  hex.Coord{Q: 1, R: 0},
		Dir:  hex.RotNW.TravelDir(),
	}}

	_, snaps := sim.RunTrace(board.NewEmpty(), rand.New(rand.NewSource(1)), cfg)
	for _, snap := range snaps {
		if snap.Event != "Waerme-Reflektion" {
			continue
		}
		if !strings.Contains(snap.Narrative, "Richtung SE") {
			t.Fatalf("NW into wall should reflect to SE, got %q", snap.Narrative)
		}
		return
	}
	t.Fatal("expected heat reflection from NW")
}

func TestVoltageDeliveryConsumesDemand(t *testing.T) {
	cfg := testCfg()
	cfg.ShiftDemands = board.ShiftDemands{Residential: 1}
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  hex.Coord{Q: 8, R: 0},
		Dir:  hex.RotE.TravelDir(),
	}}

	rng := rand.New(rand.NewSource(5))
	res := sim.Run(board.NewEmpty(), rng, cfg)
	if res.ZoneDeliveries[board.ZoneResidential] != 1 {
		t.Fatalf("expected one residential delivery, got %d", res.ZoneDeliveries[board.ZoneResidential])
	}
}

func TestHeatReflectsOffInternalWall(t *testing.T) {
	cfg := testCfg()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipHeat,
		Pos:  hex.Coord{Q: 4, R: 0},
		Dir:  0, // east into internal wall
	}}

	rng := rand.New(rand.NewSource(4))
	_, snaps := sim.RunTrace(board.NewEmpty(), rng, cfg)
	for _, snap := range snaps {
		if snap.Event != "Waerme-Reflektion" {
			continue
		}
		if !strings.Contains(snap.Narrative, "Richtung W") {
			t.Fatalf("east into internal wall should reflect west, got %q", snap.Narrative)
		}
		return
	}
	t.Fatal("expected heat reflection at internal wall")
}

func TestHeatReflectsOffInternalWallFromPlayer2(t *testing.T) {
	cfg := testCfg()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipHeat,
		Pos:  hex.Coord{Q: 5, R: 2},
		Dir:  hex.RotW.TravelDir(),
	}}

	_, snaps := sim.RunTrace(board.NewEmpty(), rand.New(rand.NewSource(1)), cfg)
	for _, snap := range snaps {
		if snap.Event != "Waerme-Reflektion" {
			continue
		}
		if !strings.Contains(snap.Narrative, "Richtung E") {
			t.Fatalf("west into internal wall should reflect east, got %q", snap.Narrative)
		}
		return
	}
	t.Fatal("expected heat reflection from player 2 side of internal wall")
}

func TestVoltageSEDoesNotConsumePlantDemand(t *testing.T) {
	cfg := testCfg()
	cfg.ShiftDemands = board.ShiftDemands{Plant: 1}
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  hex.Coord{Q: 8, R: 1},
		Dir:  hex.RotSE.TravelDir(),
	}}

	_, snaps := sim.RunTrace(board.NewEmpty(), rand.New(rand.NewSource(1)), cfg)
	firstBoundary := ""
	for _, snap := range snaps {
		if strings.HasPrefix(snap.Event, "Rand-Bedarf ") || strings.HasPrefix(snap.Event, "Rand-Schaden ") {
			firstBoundary = snap.Event
			break
		}
	}
	if firstBoundary != board.BorderDamageEvent(board.ZoneResidential) {
		t.Fatalf("SE into wall with only plant demand must damage residential first, got %q", firstBoundary)
	}
}

func TestVoltageNWConsumesIndustryNotPlant(t *testing.T) {
	cfg := testCfg()
	cfg.ShiftDemands = board.ShiftDemands{Plant: 1, Industry: 1}
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  hex.Coord{Q: 6, R: 0},
		Dir:  hex.RotNW.TravelDir(),
	}}

	res := sim.Run(board.NewEmpty(), rand.New(rand.NewSource(1)), cfg)
	if res.ZoneDeliveries[board.ZoneIndustry] != 1 {
		t.Fatalf("industry deliveries = %d, want 1", res.ZoneDeliveries[board.ZoneIndustry])
	}
	if res.ZoneDeliveries[board.ZonePlant] != 0 {
		t.Fatalf("plant deliveries = %d, want 0", res.ZoneDeliveries[board.ZonePlant])
	}
}

func TestDemandDecrementsOnWallDelivery(t *testing.T) {
	s := board.NewEmpty()
	s.ApplyDemands(board.ShiftDemands{Residential: 2})
	if s.TotalDemand(board.ZoneResidential) != 2 {
		t.Fatalf("expected 2 residential demands, got %d", s.TotalDemand(board.ZoneResidential))
	}
	rng := rand.New(rand.NewSource(6))
	_, ok := s.TryConsumeWallDemand(hex.Coord{Q: 8, R: 1}, hex.RotE.TravelDir(), rng)
	if !ok {
		t.Fatal("expected demand consumption")
	}
	if s.TotalDemand(board.ZoneResidential) != 1 {
		t.Fatalf("expected 1 remaining demand, got %d", s.TotalDemand(board.ZoneResidential))
	}
}
