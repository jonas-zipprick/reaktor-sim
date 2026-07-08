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
	s.Tiles[1][1] = field.NewTile(field.CoalChamber, 0, 0)

	cfg := sim.DefaultConfig()
	cfg.InitialHeat = 0
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
	cfg.MixedEmitterTrigger = false
	cfg.Trace = true

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(2)), cfg)
	if !strings.Contains(snaps[0].Narrative, "Richtung E") {
		t.Fatalf("east shot should be display dir E, got %q", snaps[0].Narrative)
	}
}

func TestZuenderAbprallerNarrativeIncludesDirection(t *testing.T) {
	s := board.NewEmpty()
	cfg := sim.DefaultConfig()
	cfg.InitialHeat = 0
	cfg.MixedEmitterTrigger = false
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipHeat,
		Pos:  hex.Coord{Q: 1, R: 1},
		Dir:  hex.RotW.TravelDir(),
	}}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(1)), cfg)
	for _, snap := range snaps {
		if snap.Event != "Zuender-Abpraller" {
			continue
		}
		if !strings.Contains(snap.Narrative, "wieder in Richtung") {
			t.Fatalf("expected direction in narrative, got %q", snap.Narrative)
		}
		if strings.Contains(snap.Narrative, "Wärme-Chip") {
			for _, rot := range hex.ShootRotations {
				if strings.Contains(snap.Narrative, "Richtung "+rot.String()+" ") ||
					strings.HasSuffix(snap.Narrative, "Richtung "+rot.String()+" ab.") {
					return
				}
			}
			t.Fatalf("narrative has invalid shoot direction: %q", snap.Narrative)
		}
		t.Fatalf("expected Wärme-Chip mention, got %q", snap.Narrative)
	}
	t.Fatal("no Zuender-Abpraller event in trace")
}
