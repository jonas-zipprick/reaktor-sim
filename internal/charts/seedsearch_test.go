package charts_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/charts"
	"github.com/jonas/reaktor-sim/internal/seedsearch"
)

func TestWriteSeedsearchCharts(t *testing.T) {
	dir := t.TempDir()
	outcomes := []seedsearch.Outcome{
		{Seed: 1, Wins: 0, Runs: 10, AvgSteps: 12, BoardCosts: board.PlayerCosts{Player1: 20, Player2: 25}},
		{Seed: 2, Wins: 3, Runs: 10, AvgSteps: 12, BoardCosts: board.PlayerCosts{Player1: 20, Player2: 25}},
		{Seed: 3, Wins: 7, Runs: 10, AvgSteps: 8, BoardCosts: board.PlayerCosts{Player1: 22, Player2: 25}},
		{Seed: 4, Wins: 7, Runs: 10, AvgSteps: 15, BoardCosts: board.PlayerCosts{Player1: 20, Player2: 28}},
	}
	if err := charts.WriteSeedsearchCharts(dir, outcomes, 10); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{
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
