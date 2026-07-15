package board

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/rules"
)

func TestSpendShiftBudgetRemovesAndPlaces(t *testing.T) {
	prev := Random(rand.New(rand.NewSource(42)), 0)
	unchanged := prev.Clone()

	if _, err := SpendShiftBudget(rand.New(rand.NewSource(3)), prev, 5, 0, 0, rules.Month{}); err != nil {
		t.Fatal(err)
	}
	if prev.PlayerCosts() == unchanged.PlayerCosts() {
		t.Fatal("expected board to change after spending budget")
	}
}

func TestSpendShiftBudgetDeterministic(t *testing.T) {
	a := Random(rand.New(rand.NewSource(5)), 0)
	b := Random(rand.New(rand.NewSource(5)), 0)
	if _, err := SpendShiftBudget(rand.New(rand.NewSource(3)), a, 4, 2, 0, rules.Month{}); err != nil {
		t.Fatal(err)
	}
	if _, err := SpendShiftBudget(rand.New(rand.NewSource(3)), b, 4, 2, 0, rules.Month{}); err != nil {
		t.Fatal(err)
	}
	if a.PlayerCosts() != b.PlayerCosts() {
		t.Fatalf("non-deterministic budget spend: %+v vs %+v", a.PlayerCosts(), b.PlayerCosts())
	}
}

func TestSpendShiftBudgetZeroLeavesBoard(t *testing.T) {
	prev := Random(rand.New(rand.NewSource(99)), 0)
	before := prev.Clone()
	if _, err := SpendShiftBudget(rand.New(rand.NewSource(1)), prev, 0, 0, 0, rules.Month{}); err != nil {
		t.Fatal(err)
	}
	if prev.PlayerCosts() != before.PlayerCosts() {
		t.Fatalf("budget 0 changed board: %+v -> %+v", before.PlayerCosts(), prev.PlayerCosts())
	}
}
