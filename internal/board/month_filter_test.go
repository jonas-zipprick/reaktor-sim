package board_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/rules"
)

func TestSpendShiftBudgetRespectsMonthFilter(t *testing.T) {
	s := board.NewEmpty()
	if _, err := board.SpendShiftBudget(rand.New(rand.NewSource(42)), s, 0, 8, 2, rules.Month{}); err != nil {
		t.Fatal(err)
	}
	for _, c := range hex.AllBoardCoords {
		if !c.IsPlayer2() {
			continue
		}
		tile := s.Tiles[c.Q][c.R]
		if tile.Type == field.Empty {
			continue
		}
		if !field.AvailableInMonth(tile.Type, 2) {
			t.Fatalf("placed unavailable field %v at month 2", tile.Type)
		}
	}
}

func TestRandomWithPlayerCostsRespectsMonthFilter(t *testing.T) {
	s, _, err := board.RandomWithPlayerCosts(rand.New(rand.NewSource(7)), 6, 6, 2, rules.Month{})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range hex.AllBoardCoords {
		tile := s.Tiles[c.Q][c.R]
		if tile.Type == field.Empty {
			continue
		}
		if !field.AvailableInMonth(tile.Type, 2) {
			t.Fatalf("placed unavailable field %v at month 2", tile.Type)
		}
	}
}
