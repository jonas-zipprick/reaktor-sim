package seedsearch_test

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/energy"
	"github.com/jonas/reaktor-sim/internal/finance"
	"github.com/jonas/reaktor-sim/internal/seedsearch"
)

func TestLeftoverCarriesBetweenShifts(t *testing.T) {
	fin, _ := finance.ByID("schwerindustrie")
	card, _ := energy.ByID("eroeffnungsfeier")
	opts := seedsearch.Options{
		Runs: 20, EnergyCard: card, Finance: fin, Shifts: 5, MonthFilter: 1,
	}
	chain, err := seedsearch.EvaluateChain(7, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(chain) != 5 {
		t.Fatalf("chain length = %d, want 5", len(chain))
	}
	for i := 1; i < len(chain); i++ {
		prev, cur := chain[i-1], chain[i]
		if prev.EndLeftover != cur.StartLeftover {
			t.Fatalf("shift %d carry: endLeft %+v != next startLeft %+v",
				prev.Shift, prev.EndLeftover, cur.StartLeftover)
		}
	}
}

func TestCampaignMoneyBalancesMonthBudget(t *testing.T) {
	fin, _ := finance.ByID("schwerindustrie")
	card, _ := energy.ByID("eroeffnungsfeier")
	opts := seedsearch.Options{
		Runs: 50, EnergyCard: card, Finance: fin, Shifts: 5, ShiftKeep: 3, MonthFilter: 1,
	}
	chain, err := seedsearch.EvaluateChain(46056, opts)
	if err != nil {
		t.Fatal(err)
	}
	cm := seedsearch.CampaignMoneyFromChain(chain, fin)
	last := chain[len(chain)-1]
	cases := []struct {
		label   string
		budget  int
		board   int
		rebuild int
		repair  int
		saved   int
	}{
		{"P1", cm.MonthBudget[0], last.BoardCosts.Player1, int(cm.RebuildF[0] + 0.5), int(cm.AvgRepairF[0] + 0.5), int(last.AvgSavedP1 + 0.5)},
		{"P2", cm.MonthBudget[1], last.BoardCosts.Player2, int(cm.RebuildF[1] + 0.5), int(cm.AvgRepairF[1] + 0.5), int(last.AvgSavedP2 + 0.5)},
	}
	for _, tc := range cases {
		accounted := tc.board + tc.rebuild + tc.repair + tc.saved
		if accounted != tc.budget {
			t.Fatalf("%s budget=%d accounted=%d (board=%d rebuild=%d repair=%d saved=%d)",
				tc.label, tc.budget, accounted, tc.board, tc.rebuild, tc.repair, tc.saved)
		}
	}
}
