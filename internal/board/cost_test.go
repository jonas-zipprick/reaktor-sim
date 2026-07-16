package board

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/rules"
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
	s, left, err := RandomWithPlayerCosts(rand.New(rand.NewSource(2)), 15, 0, 0, 0, rules.Month{})
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
	_, _, err := RandomWithPlayerCosts(rand.New(rand.NewSource(1)), -1, 5, 0, 0, rules.Month{})
	if err == nil {
		t.Fatal("expected error for negative player1 cost")
	}
}

func TestRandomWithPlayerCostsBothZero(t *testing.T) {
	s, _, err := RandomWithPlayerCosts(rand.New(rand.NewSource(1)), 0, 0, 0, 0, rules.Month{})
	if err != nil {
		t.Fatal(err)
	}
	if s.TotalCost() != 0 {
		t.Fatalf("expected empty board, got cost %d", s.TotalCost())
	}
}

func TestRandomWithPlayerCostsPlacesInteriorSlots(t *testing.T) {
	// Low budgets previously always spent on early edge slots (row-major order).
	interior := map[hex.Coord]bool{
		{Q: 1, R: 2}: true,
		{Q: 2, R: 2}: true,
		{Q: 3, R: 2}: true,
		{Q: 6, R: 2}: true,
		{Q: 7, R: 2}: true,
		{Q: 8, R: 2}: true,
	}
	seen := 0
	for seed := int64(1); seed <= 300; seed++ {
		s, _, err := RandomWithPlayerCosts(rand.New(rand.NewSource(seed)), 3, 3, 0, 2, rules.Month{})
		if err != nil {
			t.Fatal(err)
		}
		for c, ok := range interior {
			if !ok {
				continue
			}
			t := s.Tiles[c.Q][c.R]
			if t.Type != field.Empty {
				seen++
				break
			}
		}
		if seen >= 5 {
			return
		}
	}
	t.Fatalf("interior slots filled in only %d/300 boards, want some middle placements", seen)
}
