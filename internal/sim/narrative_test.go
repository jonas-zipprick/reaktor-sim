package sim_test

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestTraceNarrativeCoalReaction(t *testing.T) {
	s := board.NewEmpty()
	s.Tiles[1][2] = field.NewTile(field.CoalChamber, 0, 0)

	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipHeat,
		Pos:  hex.Coord{Q: hex.EmitterCol, R: hex.EmitterRow},
		Dir:  hex.RotE.TravelDir(),
	}}
	cfg.Trace = true

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(1)), cfg)
	if len(snaps) < 2 {
		t.Fatalf("snapshots: %d", len(snaps))
	}
	n := snaps[1].Narrative
	if n == "" {
		t.Fatal("expected narrative text")
	}
	if snaps[1].Narrative == snaps[1].Event {
		t.Fatalf("narrative should be descriptive, got %q", n)
	}
	if !strings.Contains(n, "Kohle") {
		t.Fatalf("expected coal mention, got %q", n)
	}
}

func TestTraceNarrativeStartUsesDisplayDir(t *testing.T) {
	s := board.NewEmpty()
	cfg := sim.DefaultConfig()
	cfg.StartDir = hex.RotE.TravelDir()
	cfg.Trace = true

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(2)), cfg)
	if !strings.Contains(snaps[0].Narrative, "Richtung E") {
		t.Fatalf("east shot should be display dir E, got %q", snaps[0].Narrative)
	}
}

func TestZuenderTrefferNarrative(t *testing.T) {
	s := board.NewEmpty()
	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipHeat,
		Pos:  hex.Coord{Q: 1, R: 2},
		Dir:  hex.RotW.TravelDir(),
	}}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(1)), cfg)
	for _, snap := range snaps {
		if snap.Event != "Zuender-Treffer" {
			continue
		}
		if !strings.Contains(snap.Narrative, "Zünder") || !strings.Contains(snap.Narrative, "vernichtet") {
			t.Fatalf("narrative = %q", snap.Narrative)
		}
		return
	}
	t.Fatal("no Zuender-Treffer event in trace")
}
