package sim_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestUraniumEmitsNeutronsOnNeutronHit(t *testing.T) {
	s := board.NewEmpty()
	uranium := hex.Coord{Q: 2, R: 1}
	s.Tiles[uranium.Q][uranium.R] = field.NewTile(field.UraniumPlate, 0, 0)

	cfg := sim.DefaultConfig()
	cfg.InitialHeat = 0
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipNeutron,
		Pos:  hex.Coord{Q: 1, R: 1},
		Dir:  hex.RotE.TravelDir(),
	}}
	cfg.Trace = true

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(1)), cfg)
	if len(snaps) < 2 {
		t.Fatalf("expected snapshots, got %d", len(snaps))
	}

	foundNeutrons := false
	consumed := false
	for _, snap := range snaps[1:] {
		for _, chip := range snap.Queue {
			if chip.Type == sim.ChipNeutron && chip.Pos == uranium {
				foundNeutrons = true
			}
		}
		if snap.Board.Tiles[uranium.Q][uranium.R].Charge < 2 {
			consumed = true
		}
	}
	if !foundNeutrons {
		t.Fatal("uranium should emit neutron chips after neutron hit")
	}
	if !consumed {
		t.Fatal("uranium should consume charge on neutron hit")
	}
}

func TestMixedEmitterCanReachUranium(t *testing.T) {
	s := board.NewEmpty()
	s.Tiles[1][1] = field.NewTile(field.UraniumPlate, 0, 0)

	cfg := sim.DefaultConfig()
	cfg.MixedEmitterTrigger = true
	cfg.StartDir = hex.RotE.TravelDir()

	neutronRuns := 0
	for seed := int64(1); seed < 200; seed++ {
		runCfg := cfg
		runCfg.Trace = true
		_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(seed)), runCfg)
		for _, snap := range snaps {
			for _, chip := range snap.Queue {
				if chip.Type == sim.ChipNeutron {
					neutronRuns++
					break
				}
			}
		}
	}
	if neutronRuns == 0 {
		t.Fatal("mixed emitter should produce neutrons when uranium is on path")
	}
}

func TestUraniumDeflectsHeatRandomly(t *testing.T) {
	uranium := hex.Coord{Q: 2, R: 1}
	dirs := map[int]bool{}

	for seed := int64(0); seed < 40; seed++ {
		s := board.NewEmpty()
		s.Tiles[uranium.Q][uranium.R] = field.NewTile(field.UraniumPlate, 0, 0)

		cfg := sim.DefaultConfig()
		cfg.InitialHeat = 0
		cfg.InitialChips = []sim.Chip{{
			Type: sim.ChipHeat,
			Pos:  hex.Coord{Q: 1, R: 1},
			Dir:  hex.RotE.TravelDir(),
		}}

		_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(seed)), cfg)
		for _, snap := range snaps {
			if snap.Event != "Feldreaktion" {
				continue
			}
			if snap.Board.Tiles[uranium.Q][uranium.R].Charge != 2 {
				t.Fatalf("seed %d: heat hit should not consume uranium charge", seed)
			}
			for _, chip := range snap.Queue {
				if chip.Type == sim.ChipHeat && chip.Pos == uranium {
					dirs[chip.Dir] = true
				}
			}
		}
	}
	if len(dirs) < 2 {
		t.Fatalf("uranium should deflect heat in varying directions, got %v", dirs)
	}
}
