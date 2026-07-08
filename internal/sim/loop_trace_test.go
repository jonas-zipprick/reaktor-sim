package sim_test

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestLoopTraceRunIndices(t *testing.T) {
	results := []sim.Result{
		{},
		{StepLimitExceeded: true},
		{},
		{StepLimitExceeded: true},
		{StepLimitExceeded: true},
	}

	got := sim.LoopTraceRunIndices(results, 2)
	want := []int{2, 4}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}

	if sim.LoopTraceRunIndices(results, 0) != nil {
		t.Fatal("max 0 should return nil")
	}
	if sim.LoopTraceRunIndices(nil, 5) != nil {
		t.Fatal("nil results should return nil")
	}
}
