package board_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/rules"
)

func TestRandomWithPlayerCostsCanLeaveLeftover(t *testing.T) {
	const budget = 6
	seenLeftover := false
	for seed := int64(1); seed <= 200; seed++ {
		_, left, err := board.RandomWithPlayerCosts(rand.New(rand.NewSource(seed)), budget, 0, 0, rules.Month{})
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
	left, err := board.SpendShiftBudget(rand.New(rand.NewSource(99)), s, 3, 0, 0, rules.Month{})
	if err != nil {
		t.Fatal(err)
	}
	if left.Player1 < 0 || left.Player1 > 3 {
		t.Fatalf("leftover = %d, want 0..3", left.Player1)
	}
}

func TestSpendShiftBudgetReservesRepairMoneyWhenHighDamage(t *testing.T) {
	base := board.NewEmpty()
	base.Damage = [4]int{2, 1, 1, 0} // 4 total across board
	month := rules.Month{FinanceID: "schwerindustrie"}

	var totalLeft float64
	const runs = 300
	for i := 0; i < runs; i++ {
		s := base.Clone()
		left, err := board.SpendShiftBudget(rand.New(rand.NewSource(int64(i))), s, 0, 8, 0, month)
		if err != nil {
			t.Fatal(err)
		}
		totalLeft += float64(left.Player2)
	}
	avg := totalLeft / runs
	if avg < 5.5 {
		t.Fatalf("avg leftover %.2f with 4 total damage, want repair reserve (>= 5.5)", avg)
	}
}

func TestSpendShiftBudgetReservesEmitterRepairOnTotalBoardDamage(t *testing.T) {
	base := board.NewEmpty()
	base.EmitterDamage = 2
	base.Damage = [4]int{1, 1, 0, 0} // 4 total, only 2 on P1
	month := rules.Month{FinanceID: "schwerindustrie"}

	var totalLeft float64
	const runs = 300
	for i := 0; i < runs; i++ {
		s := base.Clone()
		left, err := board.SpendShiftBudget(rand.New(rand.NewSource(int64(i))), s, 6, 0, 0, month)
		if err != nil {
			t.Fatal(err)
		}
		totalLeft += float64(left.Player1)
	}
	avg := totalLeft / runs
	if avg < 3.5 {
		t.Fatalf("avg P1 leftover %.2f with 4 total damage, want emitter repair reserve (>= 3.5)", avg)
	}
}

func TestSpendShiftBudgetNoRepairReserveWhenTotalDamageLow(t *testing.T) {
	base := board.NewEmpty()
	base.EmitterDamage = 1
	base.Damage = [4]int{1, 1, 0, 0} // 3 total
	month := rules.Month{FinanceID: "schwerindustrie"}

	var totalLeft float64
	const runs = 300
	for i := 0; i < runs; i++ {
		s := base.Clone()
		left, err := board.SpendShiftBudget(rand.New(rand.NewSource(int64(i))), s, 0, 8, 0, month)
		if err != nil {
			t.Fatal(err)
		}
		totalLeft += float64(left.Player2)
	}
	avg := totalLeft / runs
	if avg > 5.5 {
		t.Fatalf("avg leftover %.2f with 3 total damage, want no strong reserve (<= 5.5)", avg)
	}
}

func TestSpendShiftBudgetRepairHeuristicDisabledWithoutRepairs(t *testing.T) {
	base := board.NewEmpty()
	base.Damage = [4]int{6, 0, 0, 0}
	month := rules.Month{FinanceID: "wettruesten"}

	var totalLeft float64
	const runs = 300
	for i := 0; i < runs; i++ {
		s := base.Clone()
		left, err := board.SpendShiftBudget(rand.New(rand.NewSource(int64(i))), s, 0, 8, 0, month)
		if err != nil {
			t.Fatal(err)
		}
		totalLeft += float64(left.Player2)
	}
	avg := totalLeft / runs
	if avg < 3 || avg > 5 {
		t.Fatalf("avg leftover %.2f without repairs allowed, want ~4 (uniform)", avg)
	}
}
