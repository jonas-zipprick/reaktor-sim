package sim_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestIgniterDestroysChip(t *testing.T) {
	emitter := hex.Coord{Q: hex.EmitterCol, R: hex.EmitterRow}
	s := board.NewEmpty()
	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipHeat,
		Pos:  hex.Coord{Q: 1, R: 1},
		Dir:  hex.RotW.TravelDir(),
	}}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(1)), cfg)
	destroyed := false
	for _, snap := range snaps {
		if snap.Event != "Zuender-Treffer" {
			continue
		}
		destroyed = true
		for _, c := range snap.Queue {
			if c.Pos == emitter {
				t.Fatalf("chip should not remain at emitter: %+v", c)
			}
		}
	}
	if !destroyed {
		t.Fatal("expected Zuender-Treffer event")
	}
}

func TestTurbineVoltageOnlyValidShootDirs(t *testing.T) {
	tCoord := hex.Coord{Q: hex.TurbineCol, R: hex.TurbineRow}
	for seed := int64(0); seed < 100; seed++ {
		s := board.NewEmpty()
		cfg := sim.DefaultConfig()
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
