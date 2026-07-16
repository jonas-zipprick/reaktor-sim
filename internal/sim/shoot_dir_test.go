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

func TestDefaultConfigRandomizesEmitterShootDir(t *testing.T) {
	s := board.NewEmpty()
	s.ApplyDemands(board.ShiftDemands{Plant: 1})
	cfg := sim.DefaultConfig()
	if cfg.StartDir != -1 {
		t.Fatalf("StartDir = %d, want -1 (random)", cfg.StartDir)
	}
	if cfg.TurbineStartDir != -1 {
		t.Fatalf("TurbineStartDir = %d, want -1 (random)", cfg.TurbineStartDir)
	}
	seen := map[int]bool{}
	for seed := int64(1); seed <= 100; seed++ {
		chips := sim.EmitterChips(s, cfg, rand.New(rand.NewSource(seed)))
		if len(chips) != 1 {
			t.Fatalf("seed %d: chips = %d", seed, len(chips))
		}
		seen[chips[0].Dir] = true
	}
	want := map[int]bool{
		hex.RotNE.TravelDir(): true,
		hex.RotE.TravelDir():  true,
		hex.RotSE.TravelDir(): true,
	}
	for dir := range want {
		if !seen[dir] {
			t.Fatalf("missing shoot dir %d in 100 seeds (seen %v)", dir, seen)
		}
	}
	for dir := range seen {
		if !want[dir] {
			t.Fatalf("unexpected shoot dir %d", dir)
		}
	}
}

func TestRunTurbineVoltageUsesFixedShootDir(t *testing.T) {
	tCoord := hex.Coord{Q: hex.TurbineCol, R: hex.TurbineRow}
	cfg := testCfg()
	cfg.TurbineStartDir = hex.RotE.TravelDir()
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
	if dirs[0] != hex.RotE.TravelDir() {
		t.Fatalf("turbine dir = %d, want East", dirs[0])
	}
}
