package board

import (
	"math/rand"
	"testing"
)

func TestPlayerCostsSplitByHalf(t *testing.T) {
	s, err := randomWithPlayerCostsExact(rand.New(rand.NewSource(1)), 12, 8, 0)
	if err != nil {
		t.Fatal(err)
	}
	costs := s.PlayerCosts()
	if costs.Player1 != 12 {
		t.Fatalf("player1 = %d, want 12", costs.Player1)
	}
	if costs.Player2 != 8 {
		t.Fatalf("player2 = %d, want 8", costs.Player2)
	}
	if costs.Total() != 20 {
		t.Fatalf("total = %d, want 20", costs.Total())
	}
}

func TestRandomWithPlayerCostsEmptyHalf(t *testing.T) {
	s, left, err := RandomWithPlayerCosts(rand.New(rand.NewSource(2)), 15, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	costs := s.PlayerCosts()
	if costs.Player1 > 15 || costs.Player2 != 0 {
		t.Fatalf("costs = %+v, want P1<=15 P2=0", costs)
	}
	if left.Player1+costs.Player1 != 15 {
		t.Fatalf("spent %+v + leftover %d != budget 15", costs, left.Player1)
	}
}

func TestRandomWithPlayerCostsNegative(t *testing.T) {
	_, _, err := RandomWithPlayerCosts(rand.New(rand.NewSource(1)), -1, 5, 0)
	if err == nil {
		t.Fatal("expected error for negative player1 cost")
	}
}

func TestRandomWithPlayerCostsBothZero(t *testing.T) {
	s, _, err := RandomWithPlayerCosts(rand.New(rand.NewSource(1)), 0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if s.TotalCost() != 0 {
		t.Fatalf("expected empty board, got cost %d", s.TotalCost())
	}
}
