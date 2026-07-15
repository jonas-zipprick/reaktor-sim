package sim_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestEmitterChipsShareShootDir(t *testing.T) {
	s := board.NewEmpty()
	s.ApplyDemands(board.ShiftDemands{Plant: 1})
	cfg := sim.DefaultConfig()
	rng := rand.New(rand.NewSource(7))
	chips := sim.EmitterChips(s, cfg, rng)
	if len(chips) != 1 {
		t.Fatalf("chips = %d, want 1", len(chips))
	}
	cfg.StartDir = hex.RotE.TravelDir()
	chips2 := sim.EmitterChips(s, cfg, rng)
	if chips2[0].Dir != hex.RotE.TravelDir() {
		t.Fatalf("dir = %d, want %d", chips2[0].Dir, hex.RotE.TravelDir())
	}
}

func TestRunTurbineVoltageUsesFixedShootDir(t *testing.T) {
	tCoord := hex.Coord{Q: hex.TurbineCol, R: hex.TurbineRow}
	cfg := testCfg()
	cfg.InitialChips = []sim.Chip{
		{Type: sim.ChipHeat, Pos: hex.Coord{Q: 3, R: 1}, Dir: 0},
		{Type: sim.ChipHeat, Pos: hex.Coord{Q: 3, R: 3}, Dir: 0},
	}
	_, snaps := sim.RunTrace(board.NewEmpty(), rand.New(rand.NewSource(11)), cfg)
	var dirs []int
	for _, snap := range snaps {
		if snap.Event != "Turbine" {
			continue
		}
		for _, chip := range snap.Queue {
			if chip.Type == sim.ChipVoltage && chip.Pos == tCoord {
				dirs = append(dirs, chip.Dir)
			}
		}
	}
	if len(dirs) < 2 {
		t.Fatalf("expected at least 2 turbine voltage chips in trace, got %d", len(dirs))
	}
	for i := 1; i < len(dirs); i++ {
		if dirs[i] != dirs[0] {
			t.Fatalf("turbine dirs differ: %v", dirs)
		}
	}
}
