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
		_, left, err := board.RandomWithPlayerCosts(rand.New(rand.NewSource(seed)), budget, 0, 0, 0, rules.Month{})
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
	res, err := board.SpendShiftBudget(rand.New(rand.NewSource(99)), s, 3, 0, 0, 0, rules.Month{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Leftover.Player1 < 0 || res.Leftover.Player1 > 3 {
		t.Fatalf("leftover = %d, want 0..3", res.Leftover.Player1)
	}
}

func TestSpendShiftBudgetRepairsHighDamage(t *testing.T) {
	base := board.NewEmpty()
	base.Damage = [4]int{2, 1, 1, 0}
	month := rules.Month{FinanceID: "schwerindustrie"}

	var totalRepair float64
	const runs = 500
	for i := 0; i < runs; i++ {
		s := base.Clone()
		res, err := board.SpendShiftBudget(rand.New(rand.NewSource(int64(i))), s, 0, 8, 0, 0, month)
		if err != nil {
			t.Fatal(err)
		}
		totalRepair += float64(res.TotalRepairP2())
	}
	avg := totalRepair / runs
	if avg < 0.3 {
		t.Fatalf("avg grid repair %.2f with 4 damage, want > 0.3", avg)
	}
}

func TestSpendShiftBudgetRepairsEmitterDamage(t *testing.T) {
	base := board.NewEmpty()
	base.EmitterDamage = 2
	base.Damage = [4]int{1, 1, 0, 0}
	month := rules.Month{FinanceID: "schwerindustrie"}

	var totalRepair float64
	const runs = 500
	for i := 0; i < runs; i++ {
		s := base.Clone()
		res, err := board.SpendShiftBudget(rand.New(rand.NewSource(int64(i))), s, 6, 0, 0, 0, month)
		if err != nil {
			t.Fatal(err)
		}
		totalRepair += float64(res.TotalRepairP1())
	}
	avg := totalRepair / runs
	if avg < 0.3 {
		t.Fatalf("avg reactor repair %.2f with 2 emitter damage, want > 0.3", avg)
	}
}

func TestSpendShiftBudgetRepairDisabledWithoutRepairs(t *testing.T) {
	base := board.NewEmpty()
	base.Damage = [4]int{6, 0, 0, 0}
	month := rules.Month{FinanceID: "wettruesten"}

	for i := 0; i < 100; i++ {
		s := base.Clone()
		res, err := board.SpendShiftBudget(rand.New(rand.NewSource(int64(i))), s, 0, 8, 0, 0, month)
		if err != nil {
			t.Fatal(err)
		}
		if res.TotalRepairP2() != 0 {
			t.Fatalf("wettruesten should not allow repair, got %d", res.TotalRepairP2())
		}
	}
}

func TestRandomWithPlayerCostsMinFirstShiftFieldSpend(t *testing.T) {
	const budget = 6
	for seed := int64(1); seed <= 200; seed++ {
		_, left, err := board.RandomWithPlayerCosts(rand.New(rand.NewSource(seed)), budget, budget, 0, board.MinFirstShiftFieldSpend, rules.Month{})
		if err != nil {
			t.Fatal(err)
		}
		spentP1 := budget - left.Player1
		spentP2 := budget - left.Player2
		if spentP1 < board.MinFirstShiftFieldSpend {
			t.Fatalf("seed %d: P1 spent %d Geld, want >= %d", seed, spentP1, board.MinFirstShiftFieldSpend)
		}
		if spentP2 < board.MinFirstShiftFieldSpend {
			t.Fatalf("seed %d: P2 spent %d Geld, want >= %d", seed, spentP2, board.MinFirstShiftFieldSpend)
		}
	}
}

func TestSpendShiftBudgetMinFirstShiftFieldSpend(t *testing.T) {
	const budget = 5
	for seed := int64(1); seed <= 200; seed++ {
		s := board.NewEmpty()
		res, err := board.SpendShiftBudget(rand.New(rand.NewSource(seed)), s, budget, budget, 0, board.MinFirstShiftFieldSpend, rules.Month{})
		if err != nil {
			t.Fatal(err)
		}
		spentP1 := budget - res.Leftover.Player1 - res.TotalRepairP1()
		spentP2 := budget - res.Leftover.Player2 - res.TotalRepairP2()
		if spentP1 < board.MinFirstShiftFieldSpend {
			t.Fatalf("seed %d: P1 field spend %d Geld, want >= %d", seed, spentP1, board.MinFirstShiftFieldSpend)
		}
		if spentP2 < board.MinFirstShiftFieldSpend {
			t.Fatalf("seed %d: P2 field spend %d Geld, want >= %d", seed, spentP2, board.MinFirstShiftFieldSpend)
		}
	}
}

func TestSpendShiftBudgetBiasesTowardSpending(t *testing.T) {
	const budget = 10
	month := rules.Month{FinanceID: "schwerindustrie"}
	var sumSpent, sumLeft float64
	const runs = 500
	for i := 0; i < runs; i++ {
		s := board.NewEmpty()
		res, err := board.SpendShiftBudget(rand.New(rand.NewSource(int64(i))), s, budget, 0, 0, 0, month)
		if err != nil {
			t.Fatal(err)
		}
		spent := budget - res.Leftover.Player1 - res.TotalRepairP1()
		sumSpent += float64(spent)
		sumLeft += float64(res.Leftover.Player1)
	}
	avgSpent := sumSpent / runs
	avgLeft := sumLeft / runs
	// Target uniform over 6..10 → expected spend ~8 when achievable.
	// Without bias (0..10) expected leftover would be ~5.
	if avgSpent < 6.5 {
		t.Fatalf("avg field spend %.2f, want >= 6.5 with 60%% floor", avgSpent)
	}
	if avgLeft > 3.5 {
		t.Fatalf("avg leftover %.2f, want <= 3.5 with 60%% spend bias", avgLeft)
	}
}
