package sim_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestTraceEndsImmediatelyAfterVerloren(t *testing.T) {
	s := board.NewEmpty()
	cfg := sim.DefaultConfig()
	cfg.InitialChips = make([]sim.Chip, 9)
	for i := range cfg.InitialChips {
		cfg.InitialChips[i] = sim.Chip{
			Type: sim.ChipHeat,
			Pos:  hex.Coord{Q: 1, R: 1},
			Dir:  0,
		}
	}
	cfg.Trace = true

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(1)), cfg)
	if len(snaps) == 0 {
		t.Fatal("expected snapshots")
	}
	last := snaps[len(snaps)-1]
	if last.Event != "verloren" {
		t.Fatalf("last event = %q, want verloren", last.Event)
	}
	for i, snap := range snaps {
		if snap.Event == "verloren" && i != len(snaps)-1 {
			t.Fatalf("verloren at %d with %d trailing snapshots", i, len(snaps)-i-1)
		}
	}
}
