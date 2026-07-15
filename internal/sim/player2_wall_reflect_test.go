package sim_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestVoltageReflectsOffPlayer2WestWall(t *testing.T) {
	cfg := testCfg()
	cfg.ShiftDemands = board.ShiftDemands{Industry: 1}
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  hex.Coord{Q: 6, R: 0},
		Dir:  hex.RotW.TravelDir(),
	}}

	_, snaps := sim.RunTrace(board.NewEmpty(), rand.New(rand.NewSource(3)), cfg)
	for _, snap := range snaps {
		if snap.Event == board.BorderDemandEvent(board.ZoneIndustry) {
			t.Fatal("west wall should reflect, not consume industry demand")
		}
		if snap.Event == "Spannungs-Reflektion" {
			return
		}
	}
	t.Fatal("expected voltage reflection off player-2 west wall")
}

func TestVoltageReflectsOffMarkedDiagonalWalls(t *testing.T) {
	cases := []struct {
		pos hex.Coord
		dir hex.Rotation
	}{
		{hex.Coord{Q: 4, R: 1}, hex.RotNE},
		{hex.Coord{Q: 5, R: 1}, hex.RotNW},
		{hex.Coord{Q: 4, R: 3}, hex.RotSE},
		{hex.Coord{Q: 5, R: 3}, hex.RotSW},
	}
	for _, tc := range cases {
		t.Run(tc.dir.String(), func(t *testing.T) {
			cfg := testCfg()
			cfg.InitialChips = []sim.Chip{{
				Type: sim.ChipVoltage,
				Pos:  tc.pos,
				Dir:  tc.dir.TravelDir(),
			}}
			_, snaps := sim.RunTrace(board.NewEmpty(), rand.New(rand.NewSource(5)), cfg)
			for _, snap := range snaps {
				if snap.Event == "Spannungs-Reflektion" {
					return
				}
			}
			t.Fatalf("expected reflection from (%d,%d) dir %s", tc.pos.Q, tc.pos.R, tc.dir)
		})
	}
}
