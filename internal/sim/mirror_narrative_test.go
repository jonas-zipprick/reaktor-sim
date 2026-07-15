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

func TestMirrorNarrativeIgnoresQueuedChipsOnSameCell(t *testing.T) {
	mirror := hex.Coord{Q: 2, R: 1}
	s := board.NewEmpty()
	s.Tiles[mirror.Q][mirror.R] = field.NewTile(field.Mirror, hex.RotNW, 0)

	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipHeat,
		Pos:  hex.Coord{Q: 1, R: 1},
		Dir:  hex.RotE.TravelDir(),
	}}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(1)), cfg)
	for _, snap := range snaps {
		if snap.Event != "Feldreaktion" {
			continue
		}
		if !strings.Contains(snap.Narrative, "Spiegel") {
			continue
		}
		if strings.Contains(snap.Narrative, "Richtungs-Wuerfe") {
			t.Fatalf("mirror narrative must be deterministic, got %q", snap.Narrative)
		}
		if !strings.Contains(snap.Narrative, "gelenkt") && !strings.Contains(snap.Narrative, "fliegt durch") {
			t.Fatalf("unexpected mirror narrative: %q", snap.Narrative)
		}
		return
	}
	t.Fatal("expected mirror field reaction")
}

// A mirror that deflects a heat chip into an immediately adjacent outer wall
// must bounce it back through the mirror toward its origin:
// origin -> E -> mirror -> NW -> wall -> SE -> mirror -> W -> origin.
func TestMirrorBouncesChipBackOffAdjacentWall(t *testing.T) {
	mirror := hex.Coord{Q: 3, R: 0}
	origin := hex.Coord{Q: 2, R: 0}
	s := board.NewEmpty()
	s.Tiles[mirror.Q][mirror.R] = field.NewTile(field.Mirror, hex.RotNE, 0)

	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipHeat,
		Pos:  origin,
		Dir:  hex.RotE.TravelDir(),
	}}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(1)), cfg)
	returned := false
	for _, snap := range snaps {
		if snap.Event != "Waerme-Reflektion" || snap.Active == nil {
			continue
		}
		if snap.Active.Pos == mirror && snap.Active.Dir == hex.RotW.TravelDir() {
			returned = true
			break
		}
	}
	if !returned {
		t.Fatal("mirror bounce off adjacent wall must re-deflect the chip west toward its origin")
	}
}
