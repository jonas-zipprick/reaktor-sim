package seedsearch_test

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/energy"
	"github.com/jonas/reaktor-sim/internal/finance"
	"github.com/jonas/reaktor-sim/internal/seedsearch"
)

func baseOptions() seedsearch.Options {
	return seedsearch.Options{
		Runs:       5,
		EnergyCard: energy.DefaultCard(),
		Finance:    finance.DefaultCard(),
		Shifts:     1,
	}
}

func TestEvaluateChainDeterministic(t *testing.T) {
	opts := baseOptions()
	opts.Shifts = 2
	a, err := seedsearch.EvaluateChain(42, opts)
	if err != nil {
		t.Fatal(err)
	}
	b, err := seedsearch.EvaluateChain(42, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("non-deterministic chain:\n%+v\n%+v", a, b)
	}
	if len(a) != 2 {
		t.Fatalf("chain length = %d, want 2", len(a))
	}
	if a[0].Shift != 1 || a[1].Shift != 2 {
		t.Fatalf("shift numbers = %d,%d", a[0].Shift, a[1].Shift)
	}
}

func TestEvaluateChainInitialBoardCosts(t *testing.T) {
	fin, ok := finance.ByID("schwerindustrie") // Reaktor 3 | Stromnetz 3
	if !ok {
		t.Fatal("finance card missing")
	}
	opts := baseOptions()
	opts.Runs = 1
	opts.Finance = fin
	chain, err := seedsearch.EvaluateChain(7, opts)
	if err != nil {
		t.Fatal(err)
	}
	spent := chain[0].BoardCosts
	left := chain[0].EndLeftover
	if spent.Player1+left.Player1 != 3 || spent.Player2+left.Player2 != 3 {
		t.Fatalf("shift-1 spend %+v + rest %+v != budget 3/3", spent, left)
	}
}

func TestScanMultiShiftStructure(t *testing.T) {
	opts := baseOptions()
	opts.Runs = 2
	opts.Shifts = 3
	scan, err := seedsearch.Scan(1, 4, opts, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(scan.Shifts) != 3 {
		t.Fatalf("shift results = %d, want 3", len(scan.Shifts))
	}
	for i, sr := range scan.Shifts {
		if sr.Shift != i+1 {
			t.Fatalf("shift[%d].Shift = %d", i, sr.Shift)
		}
		for _, o := range sr.Outcomes {
			if o.Shift != i+1 {
				t.Fatalf("outcome shift = %d, want %d", o.Shift, i+1)
			}
		}
	}
}

func TestScanWithStartBoard(t *testing.T) {
	base := board.Random(rand.New(rand.NewSource(99)), 0)
	fp := board.Fingerprint(base)

	opts := baseOptions()
	opts.Runs = 2
	opts.StartBoardFingerprint = fp
	scan, err := seedsearch.Scan(1, 4, opts, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(scan.Shifts[0].Outcomes) == 0 {
		t.Fatal("expected outcomes with start board")
	}
	anyChanged := false
	for _, o := range scan.Shifts[0].Outcomes {
		if o.PrevBoardFingerprint != fp {
			t.Fatalf("prev board = %q, want %q", o.PrevBoardFingerprint, fp)
		}
		if o.BoardFingerprint != fp {
			anyChanged = true
		}
	}
	if !anyChanged {
		t.Fatal("expected at least one seed to change the start board after purchases")
	}
}

func TestScanWithStartBoardInvalidFingerprint(t *testing.T) {
	opts := baseOptions()
	opts.StartBoardFingerprint = "not-a-board"
	_, err := seedsearch.EvaluateChain(1, opts)
	if err == nil {
		t.Fatal("expected error for invalid fingerprint")
	}
}

func TestAllDemandsMax1DamageRate(t *testing.T) {
	o := seedsearch.Outcome{AllDemandsMax1Damage: 7, Runs: 10}
	if got := o.AllDemandsMax1DamageRate(); got != 0.7 {
		t.Fatalf("rate = %v, want 0.7", got)
	}
}

func TestScanMultiShiftBranchesWithAllSeeds(t *testing.T) {
	opts := baseOptions()
	opts.Runs = 2
	opts.Shifts = 2
	opts.ShiftKeep = 1
	const from, to int64 = 1, 8
	scan, err := seedsearch.Scan(from, to, opts, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(scan.Shifts[0].Outcomes) == 0 {
		t.Fatal("shift 1 empty")
	}
	if len(scan.Shifts[1].Outcomes) == 0 {
		t.Fatal("shift 2 empty")
	}
	seeds := make(map[int64]bool)
	for _, o := range scan.Shifts[1].Outcomes {
		if o.PrevBoardFingerprint == "" {
			t.Fatalf("shift 2 outcome missing prev board: seed %d", o.Seed)
		}
		seeds[o.Seed] = true
	}
	if len(seeds) < 2 {
		t.Fatalf("shift 2 used only %d seed(s), want branching over multiple seeds: %v", len(seeds), seeds)
	}
}

func TestScanFiltersDuplicateBoards(t *testing.T) {
	opts := baseOptions()
	opts.Runs = 1
	const from, to int64 = 1, 200
	scan, err := seedsearch.Scan(from, to, opts, nil)
	if err != nil {
		t.Fatal(err)
	}
	unique := int64(len(scan.Shifts[0].Outcomes))
	if unique+scan.SkippedDuplicates != to-from+1 {
		t.Fatalf("unique(%d) + skipped(%d) != total(%d)", unique, scan.SkippedDuplicates, to-from+1)
	}
	seen := make(map[string]bool)
	for _, o := range scan.Shifts[0].Outcomes {
		if seen[o.BoardFingerprint] {
			t.Fatalf("duplicate fingerprint survived filtering: %s", o.BoardFingerprint)
		}
		seen[o.BoardFingerprint] = true
	}
}

func TestEstimateScanWork(t *testing.T) {
	opts := seedsearch.Options{Shifts: 1}
	if got := seedsearch.EstimateScanWork(1, 100, opts); got != 100 {
		t.Fatalf("single shift = %d, want 100", got)
	}
	opts.Shifts = 5
	opts.ShiftKeep = 1
	// 100 seeds + 4 branch shifts × 4 max parents × 100 seeds
	want := int64(100 + 4*4*100)
	if got := seedsearch.EstimateScanWork(1, 100, opts); got != want {
		t.Fatalf("5 shifts = %d, want %d", got, want)
	}
}

func TestTotalMedianEndTotals(t *testing.T) {
	o := seedsearch.Outcome{
		MedianEndDemand: [4]int{1, 2, 0, 1},
		MedianEndDamage: [4]int{0, 1, 1, 0},
	}
	if got := o.TotalMedianEndDemand(); got != 4 {
		t.Fatalf("TotalMedianEndDemand = %d, want 4", got)
	}
	if got := o.TotalMedianEndDamage(); got != 2 {
		t.Fatalf("TotalMedianEndDamage = %d, want 2", got)
	}
}

func TestWinningOnly(t *testing.T) {
	outcomes := []seedsearch.Outcome{
		{Seed: 1, Wins: 0},
		{Seed: 2, Wins: 3},
		{Seed: 3, Wins: 1},
	}
	got := seedsearch.WinningOnly(outcomes)
	if len(got) != 2 || got[0].Seed != 2 || got[1].Seed != 3 {
		t.Fatalf("winning only = %+v", got)
	}
}

func TestTopWinsAndLoops(t *testing.T) {
	outcomes := []seedsearch.Outcome{
		{Seed: 1, Wins: 2, Loops: 5, Runs: 10},
		{Seed: 2, Wins: 5, Loops: 1, Runs: 10},
		{Seed: 3, Wins: 5, Loops: 8, Runs: 10},
	}
	wins := seedsearch.TopWins(outcomes, 2)
	if len(wins) != 2 || wins[0].Seed != 2 || wins[1].Seed != 3 {
		t.Fatalf("top wins = %+v", wins)
	}
	loops := seedsearch.TopLoops(outcomes, 2)
	if len(loops) != 2 || loops[0].Seed != 3 || loops[1].Seed != 1 {
		t.Fatalf("top loops = %+v", loops)
	}
}

func TestTopDemandDamageCategories(t *testing.T) {
	outcomes := []seedsearch.Outcome{
		{Seed: 1, AllDemandsNoDamage: 1, Max1DemandNoDamage: 2, Max1DemandMax1Damage: 3, Runs: 10},
		{Seed: 2, AllDemandsNoDamage: 4, Max1DemandNoDamage: 4, Max1DemandMax1Damage: 4, Runs: 10},
		{Seed: 3, AllDemandsNoDamage: 4, Max1DemandNoDamage: 5, Max1DemandMax1Damage: 6, Runs: 10},
	}
	clean := seedsearch.TopAllDemandsNoDamage(outcomes, 2)
	if len(clean) != 2 || clean[0].Seed != 2 || clean[1].Seed != 3 {
		t.Fatalf("top all demands no damage = %+v", clean)
	}
	near := seedsearch.TopMax1DemandNoDamage(outcomes, 2)
	if len(near) != 2 || near[0].Seed != 3 || near[1].Seed != 2 {
		t.Fatalf("top max1 demand no damage = %+v", near)
	}
	low := seedsearch.TopMax1DemandMax1Damage(outcomes, 2)
	if len(low) != 2 || low[0].Seed != 3 || low[1].Seed != 2 {
		t.Fatalf("top max1 demand max1 damage = %+v", low)
	}
}

