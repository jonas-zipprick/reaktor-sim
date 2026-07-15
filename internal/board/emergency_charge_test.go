package board_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/rules"
)

func TestSpendShiftBudgetEmergencyGeneratorFullChargeOnEroeffnungsfeier(t *testing.T) {
	for seed := int64(1); seed <= 500; seed++ {
		s := board.NewEmpty()
		month := rules.Month{EnergyID: "eroeffnungsfeier", FinanceID: "planwirtschaft"}
		_, err := board.SpendShiftBudget(rand.New(rand.NewSource(seed)), s, 0, 20, 1, month)
		if err != nil {
			continue
		}
		for _, c := range board.PlaceableSlots() {
			tile := s.Tiles[c.Q][c.R]
			if tile.Type != field.EmergencyGenerator {
				continue
			}
			if tile.Charge != 2 {
				t.Fatalf("seed %d: emergency generator at (%d,%d) charge=%d, want 2", seed, c.Q, c.R, tile.Charge)
			}
		}
	}
}
