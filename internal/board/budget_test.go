package board_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
)

func TestRandomWithPlayerCostsCanLeaveLeftover(t *testing.T) {
	const budget = 6
	seenLeftover := false
	for seed := int64(1); seed <= 200; seed++ {
		_, left, err := board.RandomWithPlayerCosts(rand.New(rand.NewSource(seed)), budget, 0, 0)
		if err != nil {
			t.Fatal(err)
		}
		if left.Player1 > 0 {
			seenLeftover = true
			break
		}
	}
	if !seenLeftover {
		t.Fatal("expected some seeds to leave reactor budget unspent")
	}
}

func TestSpendShiftBudgetReturnsLeftover(t *testing.T) {
	s := board.NewEmpty()
	left, err := board.SpendShiftBudget(rand.New(rand.NewSource(99)), s, 3, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if left.Player1 < 0 || left.Player1 > 3 {
		t.Fatalf("leftover = %d, want 0..3", left.Player1)
	}
}
