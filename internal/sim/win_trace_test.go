package sim_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestAllDemandsMet(t *testing.T) {
	s := board.NewEmpty()
	s.ApplyDemands(board.ShiftDemands{Industry: 1, Residential: 1})
	if sim.AllDemandsMet(s) {
		t.Fatal("expected unmet demands")
	}
	rng := rand.New(rand.NewSource(1))
	s.TryConsumeZone(board.ZoneIndustry, rng)
	s.TryConsumeZone(board.ZoneResidential, rng)
	if !sim.AllDemandsMet(s) {
		t.Fatal("expected no demands after consumption")
	}
}

func TestWinTraceRunIndices(t *testing.T) {
	results := []sim.Result{
		{AllDemandsMet: true},
		{},
		{AllDemandsMet: true},
		{AllDemandsMet: true},
	}

	got := sim.WinTraceRunIndices(results, 2)
	want := []int{1, 3}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
