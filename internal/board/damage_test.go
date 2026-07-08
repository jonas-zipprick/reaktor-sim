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
	z, ok := s.AddWallDamage(hex.Coord{Q: 8, R: 1}, hex.RotE.TravelDir(), rand.New(rand.NewSource(1)))
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
