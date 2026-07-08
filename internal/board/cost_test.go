package board

import (
	"math/rand"
	"testing"
)

func TestPlayerCostsSplitByHalf(t *testing.T) {
	s, err := RandomWithPlayerCosts(rand.New(rand.NewSource(1)), 12, 8)
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
	s, err := RandomWithPlayerCosts(rand.New(rand.NewSource(2)), 15, 0)
	if err != nil {
		t.Fatal(err)
	}
	costs := s.PlayerCosts()
	if costs.Player1 != 15 || costs.Player2 != 0 {
		t.Fatalf("costs = %+v, want P1=15 P2=0", costs)
	}
}

func TestRandomWithPlayerCostsNegative(t *testing.T) {
	_, err := RandomWithPlayerCosts(rand.New(rand.NewSource(1)), -1, 5)
	if err == nil {
		t.Fatal("expected error for negative player1 cost")
	}
}

func TestRandomWithPlayerCostsBothZero(t *testing.T) {
	s, err := RandomWithPlayerCosts(rand.New(rand.NewSource(1)), 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if s.TotalCost() != 0 {
		t.Fatalf("expected empty board, got cost %d", s.TotalCost())
	}
}
