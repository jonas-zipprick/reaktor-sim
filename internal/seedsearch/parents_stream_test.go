package seedsearch

import (
	"path/filepath"
	"testing"
)

func TestSelectParentsStreamMatchesInMemory(t *testing.T) {
	outcomes := []Outcome{
		{Seed: 1, BoardFingerprint: "a", CarryBoardFingerprint: "a", Wins: 10, Loops: 5},
		{Seed: 2, BoardFingerprint: "b", CarryBoardFingerprint: "b", Wins: 8},
		{Seed: 3, BoardFingerprint: "c", CarryBoardFingerprint: "c", Wins: 6, AllDemandsNoDamage: 4},
		{Seed: 4, BoardFingerprint: "d", CarryBoardFingerprint: "d", Wins: 3, Max1DemandNoDamage: 2},
		{Seed: 5, BoardFingerprint: "e", CarryBoardFingerprint: "e", Wins: 1, Max1DemandMax1Damage: 1},
	}
	inMem := selectParents(outcomes, 1)

	dir := t.TempDir()
	path := filepath.Join(dir, "shift1.gob")
	if err := writeOutcomes(path, outcomes); err != nil {
		t.Fatal(err)
	}
	streamed, err := selectParentsStream(path, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(streamed) != len(inMem) {
		t.Fatalf("streamed parents = %d, in-memory = %d", len(streamed), len(inMem))
	}
	for i := range inMem {
		if streamed[i] != inMem[i] {
			t.Fatalf("parent[%d]: streamed %+v != in-memory %+v", i, streamed[i], inMem[i])
		}
	}
}

func TestAppendOutcomesMultiBatchRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "shift1.gob")
	batch1 := []Outcome{{Seed: 1, Shift: 1}, {Seed: 2, Shift: 1}}
	batch2 := []Outcome{{Seed: 3, Shift: 1}}
	if err := writeOutcomeBatches(path, [][]Outcome{batch1, batch2}); err != nil {
		t.Fatal(err)
	}
	got, err := readOutcomes(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("read %d outcomes, want 3", len(got))
	}
}
