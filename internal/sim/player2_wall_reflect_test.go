package sim_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestVoltageEntersCol6ExtensionFromWest(t *testing.T) {
	cfg := testCfg()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  hex.Coord{Q: 6, R: 0},
		Dir:  hex.RotW.TravelDir(),
	}}

	_, snaps := sim.RunTrace(board.NewEmpty(), rand.New(rand.NewSource(3)), cfg)
	for _, snap := range snaps {
		if snap.Event == "Spannungs-Reflektion" {
			t.Fatal("west into (5,0) must not reflect")
		}
		for _, c := range snap.Queue {
			if c.Pos == (hex.Coord{Q: 5, R: 0}) {
				return
			}
		}
	}
	t.Fatal("expected voltage to enter extension slot (5,0)")
}

func TestVoltageReflectsOffCol6ExtensionOuterEdges(t *testing.T) {
	cases := []struct {
		pos hex.Coord
		dir hex.Rotation
	}{
		{hex.Coord{Q: 5, R: 0}, hex.RotW},
		{hex.Coord{Q: 5, R: 0}, hex.RotNE},
		{hex.Coord{Q: 5, R: 4}, hex.RotSE},
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
