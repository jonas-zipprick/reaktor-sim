package board_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/hex"
)

func TestAddWallDamageWhenNoDemand(t *testing.T) {
	s := board.NewEmpty()
	s.ApplyDemands(board.ShiftDemands{Residential: 0})
	z, ok := s.AddWallDamage(hex.Coord{Q: 8, R: 2}, hex.RotE.TravelDir(), rand.New(rand.NewSource(1)))
	if !ok {
		t.Fatal("expected damage")
	}
	if z != board.ZoneResidential {
		t.Fatalf("zone = %v", z)
	}
	if s.TotalDamage(board.ZoneResidential) != 1 {
		t.Fatalf("damage = %d", s.TotalDamage(board.ZoneResidential))
	}
}

func TestApplyDemandsAccumulates(t *testing.T) {
	s := board.NewEmpty()
	s.ApplyDemands(board.ShiftDemands{Residential: 1})
	s.ApplyDemands(board.ShiftDemands{Residential: 2})
	if s.TotalDemand(board.ZoneResidential) != 3 {
		t.Fatalf("demand = %d, want 3", s.TotalDemand(board.ZoneResidential))
	}
}

func TestBorderDamageEvent(t *testing.T) {
	got := board.BorderDamageEvent(board.ZoneRail)
	z, ok := board.ZoneFromBorderDamageEvent(got)
	if !ok || z != board.ZoneRail {
		t.Fatalf("parse %q -> %v %v", got, z, ok)
	}
}

func TestReactorRepairBudget(t *testing.T) {
	s := board.NewEmpty()
	s.EmitterDamage = 3
	if got := board.ReactorRepairBudget(0, s); got != 0 {
		t.Fatalf("no leftover = %d, want 0", got)
	}
	if got := board.ReactorRepairBudget(2, s); got != 2 {
		t.Fatalf("limited by money = %d, want 2", got)
	}
	if got := board.ReactorRepairBudget(10, s); got != 3 {
		t.Fatalf("limited by damage = %d, want 3", got)
	}
}

func TestRepairEmitterDamage(t *testing.T) {
	s := board.NewEmpty()
	s.EmitterDamage = 4
	spent := s.RepairEmitterDamage(2)
	if spent != 2 || s.EmitterDamage != 2 {
		t.Fatalf("spent=%d damage=%d, want 2/2", spent, s.EmitterDamage)
	}
}

func TestGridRepairBudget(t *testing.T) {
	s := board.NewEmpty()
	s.Damage = [4]int{2, 1, 0, 0}
	if got := board.GridRepairBudget(0, s); got != 0 {
		t.Fatalf("no leftover = %d, want 0", got)
	}
	if got := board.GridRepairBudget(2, s); got != 2 {
		t.Fatalf("limited by money = %d, want 2", got)
	}
	if got := board.GridRepairBudget(10, s); got != 3 {
		t.Fatalf("limited by damage = %d, want 3", got)
	}
}

func TestRepairRandomDamage(t *testing.T) {
	s := board.NewEmpty()
	s.Damage = [4]int{2, 1, 0, 1}
	rng := rand.New(rand.NewSource(1))
	spent := s.RepairRandomDamage(rng, 2)
	if spent != 2 {
		t.Fatalf("spent = %d, want 2", spent)
	}
	if s.TotalPlayer2Damage() != 2 {
		t.Fatalf("remaining damage = %d, want 2", s.TotalPlayer2Damage())
	}
}
