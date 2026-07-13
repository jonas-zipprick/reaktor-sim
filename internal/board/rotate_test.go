package board

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/rules"
)

func TestApplyRandomShiftRotationsChangesOrientableFields(t *testing.T) {
	s := NewEmpty()
	pos := hex.Coord{Q: 2, R: 1}
	s.Tiles[pos.Q][pos.R] = field.Tile{Type: field.Mirror, Orientation: hex.RotE}

	changed := false
	for seed := int64(0); seed < 200; seed++ {
		b := s.Clone()
		ApplyRandomShiftRotations(rand.New(rand.NewSource(seed)), b)
		if b.Tiles[pos.Q][pos.R].Orientation != hex.RotE {
			changed = true
			break
		}
	}
	if !changed {
		t.Fatal("expected at least one rotation to change mirror orientation")
	}
}

func TestApplyRandomShiftRotationsSkipsBurnedFields(t *testing.T) {
	s := NewEmpty()
	pos := hex.Coord{Q: 6, R: 1}
	s.Tiles[pos.Q][pos.R] = field.Tile{Type: field.Relay, Orientation: hex.RotE, BurnedOut: true}

	ApplyRandomShiftRotations(rand.New(rand.NewSource(1)), s)
	if s.Tiles[pos.Q][pos.R].Orientation != hex.RotE {
		t.Fatalf("burned relay orientation changed to %d", s.Tiles[pos.Q][pos.R].Orientation)
	}
}

func TestSpendShiftBudgetRotatesBeforeSpending(t *testing.T) {
	s := NewEmpty()
	pos := hex.Coord{Q: 2, R: 1}
	s.Tiles[pos.Q][pos.R] = field.Tile{Type: field.Mirror, Orientation: hex.RotE}

	if _, err := SpendShiftBudget(rand.New(rand.NewSource(4)), s, 0, 0, 0, rules.Month{}); err != nil {
		t.Fatal(err)
	}
	if s.Tiles[pos.Q][pos.R].Orientation == hex.RotE {
		t.Fatal("expected zero-budget shift to still apply random rotations")
	}
}
