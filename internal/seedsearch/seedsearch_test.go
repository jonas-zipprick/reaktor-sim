package seedsearch_test

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/energy"
	"github.com/jonas/reaktor-sim/internal/seedsearch"
)

func TestEvaluateDeterministic(t *testing.T) {
	opts := seedsearch.Options{
		Runs:         10,
		EnergyCardID: energy.DefaultCard().ID,
		Shift:        1,
	}
	a, err := seedsearch.Evaluate(42, opts)
	if err != nil {
		t.Fatal(err)
	}
	b, err := seedsearch.Evaluate(42, opts)
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatalf("non-deterministic: %+v vs %+v", a, b)
	}
	if a.Runs != 10 {
		t.Fatalf("unexpected runs: %+v", a)
	}
	if a.Wins+a.Loops > 10 {
		t.Fatalf("wins+loops exceed runs: %+v", a)
	}
}

func TestEvaluateFixedCosts(t *testing.T) {
	opts := seedsearch.Options{
		Runs:         5,
		EnergyCardID: energy.DefaultCard().ID,
		Shift:        1,
		CostP1:       15,
		CostP2:       20,
	}
	out, err := seedsearch.Evaluate(7, opts)
	if err != nil {
		t.Fatal(err)
	}
	if out.BoardCosts.Player1 != 15 || out.BoardCosts.Player2 != 20 {
		t.Fatalf("board costs = %+v, want P1=15 P2=20", out.BoardCosts)
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
