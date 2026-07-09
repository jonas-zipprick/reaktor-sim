package sim_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestStepLimitAbortsRun(t *testing.T) {
	cfg := sim.DefaultConfig()
	cfg.MaxSteps = 3
	cfg.InitialChips = make([]sim.Chip, 8)
	for i := range cfg.InitialChips {
		cfg.InitialChips[i] = sim.Chip{
			Type: sim.ChipHeat,
			Pos:  hex.Coord{Q: 2, R: 1},
			Dir:  hex.RotE.TravelDir(),
		}
	}

	res, snaps := sim.RunTrace(board.NewEmpty(), rand.New(rand.NewSource(1)), cfg)
	if !res.StepLimitExceeded {
		t.Fatal("expected step limit exceeded")
	}
	if len(snaps) == 0 || snaps[len(snaps)-1].Event != "Schrittlimit" {
		t.Fatalf("last trace event = %q, want Schrittlimit", snaps[len(snaps)-1].Event)
	}
}

func TestDefaultMaxStepsIs500(t *testing.T) {
	if sim.DefaultConfig().MaxSteps != sim.MaxStepsPerRun {
		t.Fatalf("default MaxSteps = %d, want %d", sim.DefaultConfig().MaxSteps, sim.MaxStepsPerRun)
	}
	if sim.MaxStepsPerRun != 500 {
		t.Fatalf("MaxStepsPerRun = %d, want 500", sim.MaxStepsPerRun)
	}
}
