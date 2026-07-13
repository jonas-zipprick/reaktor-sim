package charts_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/charts"
	"github.com/jonas/reaktor-sim/internal/seedsearch"
)

func TestOutcomesWithPositiveWinrate(t *testing.T) {
	outcomes := []seedsearch.Outcome{
		{Seed: 1, AllDemandsMax1Damage: 0, Runs: 10},
		{Seed: 2, AllDemandsMax1Damage: 1, Runs: 10},
		{Seed: 3, AllDemandsMax1Damage: 0, Runs: 10},
	}
	filtered := charts.OutcomesWithPositiveWinrate(outcomes)
	if len(filtered) != 1 || filtered[0].Seed != 2 {
		t.Fatalf("filtered = %+v, want seed 2 only", filtered)
	}
}

func TestWriteSeedsearchCharts(t *testing.T) {
	dir := t.TempDir()
	outcomes := []seedsearch.Outcome{
		{Seed: 1, Wins: 0, Runs: 10, AllDemandsMax1Damage: 0, AvgSteps: 12, BoardCosts: board.PlayerCosts{Player1: 20, Player2: 25}, MedianEndDemand: [4]int{0, 0, 0, 0}, MedianEndDamage: [4]int{0, 0, 0, 0}},
		{Seed: 2, Wins: 3, Runs: 10, AllDemandsMax1Damage: 5, AvgSteps: 12, BoardCosts: board.PlayerCosts{Player1: 20, Player2: 25}, MedianEndDemand: [4]int{1, 0, 0, 0}, MedianEndDamage: [4]int{1, 0, 0, 0}},
		{Seed: 3, Wins: 7, Runs: 10, AllDemandsMax1Damage: 8, AvgSteps: 8, BoardCosts: board.PlayerCosts{Player1: 22, Player2: 25}, MedianEndDemand: [4]int{0, 0, 0, 0}, MedianEndDamage: [4]int{0, 0, 0, 0}},
		{Seed: 4, Wins: 7, Runs: 10, AllDemandsMax1Damage: 9, AvgSteps: 15, BoardCosts: board.PlayerCosts{Player1: 20, Player2: 28}, MedianEndDemand: [4]int{2, 0, 0, 0}, MedianEndDamage: [4]int{2, 0, 0, 0}},
	}
	if err := charts.WriteSeedsearchCharts(dir, outcomes, 10); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{
		"verteilung_end_bedarf.png",
		"verteilung_end_schaden.png",
		"verteilung_winrate_all_demands_max1_damage.png",
		"verteilung_schritte.png",
		"verteilung_wins.png",
		"verteilung_loops.png",
	} {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("missing chart %s: %v", name, err)
		}
	}
}
