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

func TestGroundAbsorbsVoltageWithoutBurningOut(t *testing.T) {
	s := board.NewEmpty()
	pos := hex.Coord{Q: 6, R: 1}
	s.Tiles[pos.Q][pos.R] = field.NewTile(field.Ground, 0, 0)

	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  hex.Coord{Q: 5, R: 1},
		Dir:  hex.RotE.TravelDir(),
	}}

	rng := rand.New(rand.NewSource(1))
	_, snaps := sim.RunTrace(s, rng, cfg)
	found := false
	for _, snap := range snaps {
		if snap.Event != "Feldreaktion" {
			continue
		}
		if strings.Contains(snap.Narrative, "Erdung") {
			found = true
			if strings.Contains(snap.Narrative, "Ladung") {
				t.Fatalf("ground narrative should not mention charge, got %q", snap.Narrative)
			}
		}
	}
	if !found {
		t.Fatal("expected ground absorption narrative")
	}

	for i := 0; i < 10; i++ {
		cfg.InitialChips = []sim.Chip{{
			Type: sim.ChipVoltage,
			Pos:  hex.Coord{Q: 5, R: 1},
			Dir:  hex.RotE.TravelDir(),
		}}
		sim.Run(s, rng, cfg)
		tile := s.Tiles[pos.Q][pos.R]
		if tile.BurnedOut {
			t.Fatalf("ground burned out after %d voltage hits", i+1)
		}
		if tile.Charge != 0 {
			t.Fatalf("ground charge = %d after hit %d, want 0", tile.Charge, i+1)
		}
	}
}
