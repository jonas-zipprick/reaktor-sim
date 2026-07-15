package seedsearch_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jonas/reaktor-sim/internal/energy"
	"github.com/jonas/reaktor-sim/internal/finance"
	"github.com/jonas/reaktor-sim/internal/seedsearch"
)

func TestScanSpillsResultsToDisk(t *testing.T) {
	dir := t.TempDir()
	opts := seedsearch.Options{
		Runs:       2,
		EnergyCard: energy.DefaultCard(),
		Finance:    finance.DefaultCard(),
		Shifts:     2,
		ShiftKeep:  1,
		SpillDir:   dir,
	}
	scan, err := seedsearch.Scan(1, 4, opts, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer scan.Cleanup()

	for _, sr := range scan.Shifts {
		if len(sr.Outcomes) != 0 {
			t.Fatalf("shift %d still held in RAM", sr.Shift)
		}
		if sr.Len() == 0 {
			t.Fatalf("shift %d empty", sr.Shift)
		}
		path := filepath.Join(dir, fmt.Sprintf("shift%d.gob", sr.Shift))
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("shift %d spill file missing: %v", sr.Shift, err)
		}
		outcomes, err := sr.LoadOutcomes()
		if err != nil {
			t.Fatal(err)
		}
		if len(outcomes) == 0 {
			t.Fatalf("shift %d empty after reload", sr.Shift)
		}
		for _, o := range outcomes {
			if o.Shift != sr.Shift {
				t.Fatalf("outcome shift = %d, want %d", o.Shift, sr.Shift)
			}
		}
	}
}

func TestScanWithoutSpillKeepsOutcomesInMemory(t *testing.T) {
	opts := seedsearch.Options{
		Runs:       2,
		EnergyCard: energy.DefaultCard(),
		Finance:    finance.DefaultCard(),
		Shifts:     1,
	}
	scan, err := seedsearch.Scan(1, 4, opts, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(scan.Shifts[0].Outcomes) == 0 {
		t.Fatal("expected in-memory outcomes")
	}
}
