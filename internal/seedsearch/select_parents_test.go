package seedsearch

import "testing"

func TestSelectParentsSkipsLoopTableEntries(t *testing.T) {
	outcomes := []Outcome{
		{Seed: 1, BoardFingerprint: "b_loop", CarryBoardFingerprint: "c_loop", Loops: 8, Wins: 5, Runs: 10},
		{Seed: 2, BoardFingerprint: "b_win", CarryBoardFingerprint: "c_win", Loops: 0, Wins: 4, Runs: 10},
		{Seed: 3, BoardFingerprint: "b_clean", CarryBoardFingerprint: "c_clean", Loops: 0, Wins: 1, Runs: 10, AllDemandsNoDamage: 6},
	}

	parents := selectParents(outcomes, 1)
	got := make(map[string]bool)
	for _, p := range parents {
		got[p.prevFP] = true
	}
	if got["b_loop"] {
		t.Fatal("loop-table entry must not branch into next shift")
	}
	if !got["b_win"] {
		t.Fatal("expected next-ranked win board to branch")
	}
	if !got["b_clean"] {
		t.Fatal("expected clean-board table entry to branch")
	}
}
