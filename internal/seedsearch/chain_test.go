package seedsearch_test

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/energy"
	"github.com/jonas/reaktor-sim/internal/seedsearch"
)

func TestTraceChainMultiShift(t *testing.T) {
	opts := baseOptions()
	opts.Runs = 2
	opts.Shifts = 3
	opts.ShiftKeep = 1
	scan, err := seedsearch.Scan(1, 12, opts, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(scan.Shifts) != 3 {
		t.Fatalf("shifts = %d", len(scan.Shifts))
	}
	final := scan.Shifts[2]
	if len(final.Outcomes) == 0 {
		t.Fatal("shift 3 empty")
	}
	top := seedsearch.TopWins(final.Outcomes, 1)
	if len(top) == 0 {
		t.Fatal("no top wins at shift 3")
	}
	chain, err := seedsearch.TraceChain(scan, top[0], opts.EnergyCard)
	if err != nil {
		t.Fatal(err)
	}
	if len(chain) != 3 {
		t.Fatalf("chain length = %d, want 3", len(chain))
	}
	for i, o := range chain {
		if o.Shift != i+1 {
			t.Fatalf("chain[%d].Shift = %d", i, o.Shift)
		}
	}
	if chain[2].BoardFingerprint != top[0].BoardFingerprint {
		t.Fatal("chain end mismatch")
	}
	if chain[1].BoardFingerprint != top[0].PrevBoardFingerprint {
		t.Fatal("chain parent link broken at shift 2")
	}
}

func TestTraceChainSingleShift(t *testing.T) {
	opts := baseOptions()
	scan, err := seedsearch.Scan(1, 5, opts, nil)
	if err != nil {
		t.Fatal(err)
	}
	top := seedsearch.TopWins(scan.Shifts[0].Outcomes, 1)
	if len(top) == 0 {
		t.Fatal("no outcomes")
	}
	chain, err := seedsearch.TraceChain(scan, top[0], energy.DefaultCard())
	if err != nil {
		t.Fatal(err)
	}
	if len(chain) != 1 || chain[0].Shift != 1 {
		t.Fatalf("chain = %+v", chain)
	}
}
