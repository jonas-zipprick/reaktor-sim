package sim_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestCriticalMassPerSide(t *testing.T) {
	s := board.NewEmpty()
	s.ApplyDemands(board.ShiftDemands{})
	cfg := sim.DefaultConfig()
	cfg.InitialHeat = 0
	chips := make([]sim.Chip, 0, 9)
	for i := 0; i < 9; i++ {
		chips = append(chips, sim.Chip{
			Type: sim.ChipHeat,
			Pos:  hex.Coord{Q: 1, R: 1},
			Dir:  0,
		})
	}
	cfg.InitialChips = chips

	res, snaps := sim.RunTrace(s, rand.New(rand.NewSource(1)), cfg)
	if !res.CriticalFailure {
		t.Fatal("expected critical failure with 9 chips on player 1 side")
	}
	if len(snaps) == 0 || snaps[len(snaps)-1].Event != "verloren" {
		t.Fatalf("expected final trace event verloren, got %#v", snaps)
	}
	for i, snap := range snaps {
		if snap.Event == "verloren" && i != len(snaps)-1 {
			t.Fatalf("verloren at index %d but trace has %d more snapshots", i, len(snaps)-i-1)
		}
	}
}

func TestNineStoredVoltageOnPlayer2Loses(t *testing.T) {
	s := board.NewEmpty()
	s.ApplyDemands(board.ShiftDemands{})
	s.Tiles[6][1].StoredVoltage = 9

	cfg := sim.DefaultConfig()
	cfg.InitialHeat = 0

	res := sim.Run(s, rand.New(rand.NewSource(2)), cfg)
	if !res.CriticalFailure {
		t.Fatal("expected critical failure with 9 stored voltage on player 2")
	}
}
