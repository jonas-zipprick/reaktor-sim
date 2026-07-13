package seedsearch_test

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/seedsearch"
)

func TestPickUniqueOutcomesSkipsSeenFingerprints(t *testing.T) {
	outcomes := []seedsearch.Outcome{
		{Seed: 1, BoardFingerprint: "b2_AAA", Wins: 10},
		{Seed: 2, BoardFingerprint: "b2_AAA", Wins: 8},
		{Seed: 3, BoardFingerprint: "b2_BBB", Wins: 7},
		{Seed: 4, BoardFingerprint: "b2_CCC", Wins: 6},
	}
	seen := map[string]struct{}{"b2_BBB": {}}

	got := seedsearch.PickUniqueOutcomes(outcomes, seedsearch.TopWins, 2, seen)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Seed != 1 || got[1].Seed != 4 {
		t.Fatalf("got seeds %d,%d, want 1,4", got[0].Seed, got[1].Seed)
	}
	if _, ok := seen["b2_AAA"]; !ok {
		t.Fatal("b2_AAA should be marked seen")
	}
	if _, ok := seen["b2_CCC"]; !ok {
		t.Fatal("b2_CCC should be marked seen")
	}
}

func TestPickUniqueOutcomesRespectsKeepAcrossTables(t *testing.T) {
	outcomes := []seedsearch.Outcome{
		{Seed: 1, BoardFingerprint: "b2_AAA", Wins: 10},
		{Seed: 2, BoardFingerprint: "b2_BBB", Wins: 9},
		{Seed: 3, BoardFingerprint: "b2_CCC", Wins: 8},
	}
	seen := make(map[string]struct{})

	first := seedsearch.PickUniqueOutcomes(outcomes, seedsearch.TopWins, 1, seen)
	second := seedsearch.PickUniqueOutcomes(outcomes, seedsearch.TopAllDemandsNoDamage, 1, seen)

	if len(first) != 1 || first[0].Seed != 1 {
		t.Fatalf("first table = %+v", first)
	}
	if len(second) != 1 || second[0].Seed != 2 {
		t.Fatalf("second table should skip b2_AAA, got %+v", second)
	}
}
