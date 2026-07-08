package board

import (
	"fmt"

	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
)

// PlayerCosts holds placement costs per player half.
type PlayerCosts struct {
	Player1 int // Reaktor (Spieler 1)
	Player2 int // Stromnetz (Spieler 2)
}

// Total returns combined placement cost.
func (c PlayerCosts) Total() int {
	return c.Player1 + c.Player2
}

// String formats costs like the finance ministry cards.
func (c PlayerCosts) String() string {
	return fmt.Sprintf("Reaktor: %d Geld | Stromnetz: %d Geld", c.Player1, c.Player2)
}

// PlayerCosts sums placement costs per player half.
func (s *State) PlayerCosts() PlayerCosts {
	var costs PlayerCosts
	for _, c := range hex.AllBoardCoords {
		if c.Kind() != hex.CellSlot {
			continue
		}
		t := s.tileAt(c)
		if t == nil || t.Type == field.Empty {
			continue
		}
		n := t.Cost()
		if c.IsPlayer2() {
			costs.Player2 += n
		} else {
			costs.Player1 += n
		}
	}
	return costs
}

// TotalCost sums placement costs of all non-empty tiles on slots.
func (s *State) TotalCost() int {
	return s.PlayerCosts().Total()
}
