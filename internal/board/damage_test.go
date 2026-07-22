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

func TestRepairEmitterDamage(t *testing.T) {
	s := board.NewEmpty()
	s.EmitterDamage = 4
	spent := s.RepairEmitterDamage(2)
	if spent != 2 || s.EmitterDamage != 2 {
		t.Fatalf("spent=%d damage=%d, want 2/2", spent, s.EmitterDamage)
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

func TestRepairHalf(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	s := board.NewEmpty()
	s.EmitterDamage = 3
	spent := s.RepairHalf(rng, true, 5)
	if spent < 1 || spent > 3 {
		t.Fatalf("reactor repair spent = %d, want 1..3", spent)
	}
	if s.EmitterDamage != 3-spent {
		t.Fatalf("remaining emitter damage = %d, want %d", s.EmitterDamage, 3-spent)
	}

	s2 := board.NewEmpty()
	s2.Damage = [4]int{2, 1, 0, 0}
	spent2 := s2.RepairHalf(rand.New(rand.NewSource(7)), false, 10)
	if spent2 < 1 || spent2 > 3 {
		t.Fatalf("grid repair spent = %d, want 1..3", spent2)
	}
}

func TestMaybeRepairChances(t *testing.T) {
	const trials = 2000
	s := board.NewEmpty()
	s.Damage = [4]int{5, 0, 0, 0} // > threshold

	preCount := 0
	postCount := 0
	for i := 0; i < trials; i++ {
		clone := s.Clone()
		rng := rand.New(rand.NewSource(int64(i)))
		if clone.MaybeRepair(rng, false, 5, true) > 0 {
			preCount++
		}
		clone = s.Clone()
		rng = rand.New(rand.NewSource(int64(i + trials)))
		if clone.MaybeRepair(rng, false, 5, false) > 0 {
			postCount++
		}
	}
	preRate := float64(preCount) / trials
	postRate := float64(postCount) / trials
	if preRate < 0.40 || preRate > 0.60 {
		t.Fatalf("pre-purchase repair rate %.2f, want ~0.50 for high damage", preRate)
	}
	if postRate < 0.70 || postRate > 0.90 {
		t.Fatalf("post-purchase repair rate %.2f, want ~0.80 for high damage", postRate)
	}
}

func TestMaybeRepairLowDamage(t *testing.T) {
	const trials = 2000
	s := board.NewEmpty()
	s.Damage = [4]int{2, 0, 0, 0} // <= threshold

	preCount := 0
	postCount := 0
	for i := 0; i < trials; i++ {
		clone := s.Clone()
		rng := rand.New(rand.NewSource(int64(i)))
		if clone.MaybeRepair(rng, false, 5, true) > 0 {
			preCount++
		}
		clone = s.Clone()
		rng = rand.New(rand.NewSource(int64(i + trials)))
		if clone.MaybeRepair(rng, false, 5, false) > 0 {
			postCount++
		}
	}
	preRate := float64(preCount) / trials
	postRate := float64(postCount) / trials
	if preRate < 0.10 || preRate > 0.30 {
		t.Fatalf("pre-purchase repair rate %.2f, want ~0.20 for low damage", preRate)
	}
	if postRate < 0.30 || postRate > 0.50 {
		t.Fatalf("post-purchase repair rate %.2f, want ~0.40 for low damage", postRate)
	}
}
