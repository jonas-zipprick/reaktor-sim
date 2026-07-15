package board_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/rules"
)

func TestSpendShiftBudgetNoCoalOnTurbineColumn(t *testing.T) {
	for seed := int64(1); seed <= 500; seed++ {
		state := board.NewEmpty()
		_, err := board.SpendShiftBudget(rand.New(rand.NewSource(seed)), state, 20, 0, 0, rules.Month{})
		if err != nil {
			t.Fatal(err)
		}
		for _, c := range hex.AllBoardCoords {
			if c.Q != hex.TurbineCol || c.IsTurbine() {
				continue
			}
			if state.Tiles[c.Q][c.R].Type == field.CoalChamber {
				t.Fatalf("seed %d: coal placed on turbine column (%d,%d)", seed, c.Q, c.R)
			}
		}
	}
}

func TestRandomWithPlayerCostsRespectsSector(t *testing.T) {
	for seed := int64(1); seed <= 200; seed++ {
		s, _, err := board.RandomWithPlayerCosts(rand.New(rand.NewSource(seed)), 20, 20, 0, rules.Month{})
		if err != nil {
			t.Fatal(err)
		}
		for _, c := range hex.AllBoardCoords {
			tile := s.Tiles[c.Q][c.R]
			if tile.Type == field.Empty {
				continue
			}
			if !field.AllowedOnCell(tile.Type, c) {
				t.Fatalf("seed %d: %s on (%d,%d)", seed, field.Catalog[tile.Type].Name, c.Q, c.R)
			}
		}
	}
}
