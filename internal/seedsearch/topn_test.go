package seedsearch

import "testing"

func TestTopNMatchesFullSort(t *testing.T) {
	outcomes := []Outcome{
		{Seed: 1, Wins: 3},
		{Seed: 2, Wins: 9},
		{Seed: 3, Wins: 5},
		{Seed: 4, Wins: 9},
		{Seed: 5, Wins: 1},
	}
	got := TopWins(outcomes, 3)
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	if got[0].Seed != 2 || got[1].Seed != 4 || got[2].Seed != 3 {
		t.Fatalf("unexpected ranking: %+v", got)
	}
}

func TestPruneShiftOutcomes(t *testing.T) {
	parents := []Outcome{
		{Seed: 1, BoardFingerprint: "a"},
		{Seed: 2, BoardFingerprint: "b"},
		{Seed: 3, BoardFingerprint: "c"},
	}
	children := []Outcome{
		{Seed: 10, PrevBoardFingerprint: "b"},
		{Seed: 11, PrevBoardFingerprint: "b"},
	}
	pruned := pruneShiftOutcomes(parents, children)
	if len(pruned) != 1 || pruned[0].BoardFingerprint != "b" {
		t.Fatalf("pruned = %+v, want only b", pruned)
	}
}
