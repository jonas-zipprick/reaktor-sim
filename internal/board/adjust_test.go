package board

import (
	"math/rand"
	"testing"
)

func TestSpendShiftBudgetRemovesAndPlaces(t *testing.T) {
	prev := Random(rand.New(rand.NewSource(42)))
	unchanged := prev.Clone()

	if err := SpendShiftBudget(rand.New(rand.NewSource(7)), prev, 1, 0); err != nil {
		t.Fatal(err)
	}
	if prev.PlayerCosts() == unchanged.PlayerCosts() {
		t.Fatal("expected board to change after spending budget")
	}
}

func TestSpendShiftBudgetDeterministic(t *testing.T) {
	a := Random(rand.New(rand.NewSource(5)))
	b := Random(rand.New(rand.NewSource(5)))
	if err := SpendShiftBudget(rand.New(rand.NewSource(3)), a, 4, 2); err != nil {
		t.Fatal(err)
	}
	if err := SpendShiftBudget(rand.New(rand.NewSource(3)), b, 4, 2); err != nil {
		t.Fatal(err)
	}
	if a.PlayerCosts() != b.PlayerCosts() {
		t.Fatalf("non-deterministic budget spend: %+v vs %+v", a.PlayerCosts(), b.PlayerCosts())
	}
}

func TestSpendShiftBudgetZeroLeavesBoard(t *testing.T) {
	prev := Random(rand.New(rand.NewSource(99)))
	before := prev.Clone()
	if err := SpendShiftBudget(rand.New(rand.NewSource(1)), prev, 0, 0); err != nil {
		t.Fatal(err)
	}
	if prev.PlayerCosts() != before.PlayerCosts() {
		t.Fatalf("budget 0 changed board: %+v -> %+v", before.PlayerCosts(), prev.PlayerCosts())
	}
}
