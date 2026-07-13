package charts

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/stats"
)

func TestWriteCostChartStacked(t *testing.T) {
	dir := t.TempDir()
	report := stats.Report{
		Costs:      board.PlayerCosts{Player1: 15, Player2: 22},
		AvgRepairF: [2]float64{1.5, 0.5},
		AvgSavedF:  [2]float64{2.0, 1.5},
		Runs:       100,
	}
	path := filepath.Join(dir, "kosten.png")
	if err := writeCostChart(report, path); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Fatal("kosten.png is empty")
	}
}
