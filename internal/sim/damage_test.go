package sim_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestVoltageWallDamageWhenNoDemand(t *testing.T) {
	cfg := testCfg()
	cfg.ShiftDemands = board.ShiftDemands{}
	cfg.InitialHeat = 0
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  hex.Coord{Q: 8, R: 1},
		Dir:  hex.RotE.TravelDir(),
	}}

	_, snaps := sim.RunTrace(board.NewEmpty(), rand.New(rand.NewSource(3)), cfg)
	for _, snap := range snaps {
		if snap.Event == "Spannungs-Spike" {
			t.Fatal("wall overload should add damage, not spike")
		}
		if snap.Event == board.BorderDamageEvent(board.ZoneResidential) {
			if snap.Board.TotalDamage(board.ZoneResidential) < 1 {
				t.Fatal("expected residential damage")
			}
			return
		}
	}
	t.Fatal("expected border damage when no demand")
}

func TestVoltageAtTurbineDamagesWhenPlantEmpty(t *testing.T) {
	cfg := testCfg()
	cfg.ShiftDemands = board.ShiftDemands{}
	cfg.InitialHeat = 0
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  hex.Coord{Q: 5, R: 2},
		Dir:  hex.RotNW.TravelDir(),
	}}

	_, snaps := sim.RunTrace(board.NewEmpty(), rand.New(rand.NewSource(3)), cfg)
	for _, snap := range snaps {
		if snap.Event == "Spannungs-Spike" {
			t.Fatal("turbine overload should add plant damage, not spike")
		}
		if snap.Event == board.BorderDamageEvent(board.ZonePlant) {
			if snap.Board.TotalDamage(board.ZonePlant) < 1 {
				t.Fatal("expected plant damage")
			}
			return
		}
	}
	t.Fatal("expected plant damage on turbine with empty plant demand")
}

func TestZoneDamageCountsTowardCriticalMass(t *testing.T) {
	s := board.NewEmpty()
	for i := 0; i < 9; i++ {
		s.AddZoneDamage(board.ZoneResidential)
	}
	cfg := testCfg()
	cfg.ShiftDemands = board.ShiftDemands{}
	cfg.InitialHeat = 0
	cfg.InitialChips = nil

	res := sim.Run(s, rand.New(rand.NewSource(1)), cfg)
	if !res.CriticalFailure {
		t.Fatal("9 damage chips should exceed critical limit of 8 on player 2")
	}
	if !res.CriticalP2 {
		t.Fatal("expected critical failure on player 2 side")
	}
	if res.CriticalP1 {
		t.Fatal("did not expect critical failure on player 1 side")
	}
}
