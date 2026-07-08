package sim_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestIgniterDeflectOnlyValidShootDirs(t *testing.T) {
	emitter := hex.Coord{Q: hex.EmitterCol, R: hex.EmitterRow}
	for seed := int64(0); seed < 100; seed++ {
		s := board.NewEmpty()
		cfg := sim.DefaultConfig()
		cfg.InitialHeat = 0
		cfg.InitialChips = []sim.Chip{{
			Type: sim.ChipHeat,
			Pos:  hex.Coord{Q: 1, R: 1},
			Dir:  hex.RotW.TravelDir(),
		}}
		_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(seed)), cfg)
		for _, snap := range snaps {
			if snap.Event != "Zuender-Abpraller" {
				continue
			}
			for _, c := range snap.Queue {
				if c.Pos != emitter {
					continue
				}
				if !hex.ValidShootTravelDir(c.Dir) {
					t.Fatalf("seed %d: invalid shoot dir %s from emitter", seed, hex.DisplayDirName(c.Dir))
				}
			}
		}
	}
}

func TestTurbineVoltageOnlyValidShootDirs(t *testing.T) {
	tCoord := hex.Coord{Q: hex.TurbineCol, R: hex.TurbineRow}
	for seed := int64(0); seed < 100; seed++ {
		s := board.NewEmpty()
		cfg := sim.DefaultConfig()
		cfg.InitialHeat = 0
		cfg.InitialChips = []sim.Chip{{
			Type: sim.ChipHeat,
			Pos:  tCoord,
			Dir:  hex.RotE.TravelDir(),
		}}
		_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(seed)), cfg)
		for _, snap := range snaps {
			if snap.Event != "Turbine" {
				continue
			}
			for _, c := range snap.Queue {
				if c.Pos != tCoord || c.Type != sim.ChipVoltage {
					continue
				}
				if !hex.ValidShootTravelDir(c.Dir) {
					t.Fatalf("seed %d: invalid turbine shot %s", seed, hex.DisplayDirName(c.Dir))
				}
			}
		}
	}
}
