package sim_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestTurbineVoltageShootDirVariesAcrossRuns(t *testing.T) {
	seen := map[int]int{}
	for seed := int64(1); seed <= 200; seed++ {
		cfg := sim.DefaultConfig()
		cfg.InitialChips = []sim.Chip{{
			Type: sim.ChipHeat,
			Pos:  hex.Coord{Q: 3, R: 2},
			Dir:  hex.RotE.TravelDir(),
		}}
		_, snaps := sim.RunTrace(board.NewEmpty(), rand.New(rand.NewSource(seed)), cfg)
		tCoord := hex.Coord{Q: hex.TurbineCol, R: hex.TurbineRow}
		for _, snap := range snaps {
			if snap.Event != "Turbine" {
				continue
			}
			for _, c := range snap.Queue {
				if c.Type == sim.ChipVoltage && c.Pos == tCoord {
					seen[c.Dir]++
				}
			}
		}
	}
	want := []int{hex.RotNE.TravelDir(), hex.RotE.TravelDir(), hex.RotSE.TravelDir()}
	for _, d := range want {
		if seen[d] == 0 {
			t.Fatalf("missing turbine shoot dir %s in 200 seeds (seen %v)", hex.DisplayDirName(d), seen)
		}
	}
}

func TestTurbineVoltageShootDirCanVaryPerChip(t *testing.T) {
	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{
		{Type: sim.ChipHeat, Pos: hex.Coord{Q: 3, R: 1}, Dir: hex.RotE.TravelDir()},
		{Type: sim.ChipHeat, Pos: hex.Coord{Q: 3, R: 3}, Dir: hex.RotE.TravelDir()},
	}
	mixed := false
	for seed := int64(1); seed <= 300; seed++ {
		_, snaps := sim.RunTrace(board.NewEmpty(), rand.New(rand.NewSource(seed)), cfg)
		tCoord := hex.Coord{Q: hex.TurbineCol, R: hex.TurbineRow}
		var dirs []int
		for _, snap := range snaps {
			if snap.Event != "Turbine" {
				continue
			}
			// Active heat is recorded; the newly created voltage is the last one at the turbine.
			for i := len(snap.Queue) - 1; i >= 0; i-- {
				c := snap.Queue[i]
				if c.Type == sim.ChipVoltage && c.Pos == tCoord {
					dirs = append(dirs, c.Dir)
					break
				}
			}
		}
		if len(dirs) >= 2 && dirs[0] != dirs[1] {
			mixed = true
			break
		}
	}
	if !mixed {
		t.Fatal("expected at least one run with different turbine shoot dirs per converted chip")
	}
}
