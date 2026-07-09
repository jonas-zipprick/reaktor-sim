package board

import (
	"math/rand"
	"testing"
)

func TestFingerprintRoundTrip(t *testing.T) {
	orig := Random(rand.New(rand.NewSource(123)), 0)
	fp := Fingerprint(orig)
	got, err := FromFingerprint(fp)
	if err != nil {
		t.Fatal(err)
	}
	if !tilesEqual(orig, got) {
		t.Fatalf("round-trip mismatch:\norig=%+v\ngot=%+v", orig.Tiles, got.Tiles)
	}
}

func TestFingerprintAfterShiftBudget(t *testing.T) {
	prev := Random(rand.New(rand.NewSource(42)), 0)
	if _, err := SpendShiftBudget(rand.New(rand.NewSource(7)), prev, 5, 3, 0); err != nil {
		t.Fatal(err)
	}
	got, err := FromFingerprint(Fingerprint(prev))
	if err != nil {
		t.Fatal(err)
	}
	if !tilesEqual(prev, got) {
		t.Fatal("round-trip after shift budget failed")
	}
}

func TestFromFingerprintRejectsInvalid(t *testing.T) {
	if _, err := FromFingerprint("invalid"); err == nil {
		t.Fatal("expected error for invalid prefix")
	}
	if _, err := FromFingerprint(fingerprintPrefix + "!!!"); err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func tilesEqual(a, b *State) bool {
	for q := 0; q < len(a.Tiles); q++ {
		for r := 0; r < len(a.Tiles[q]); r++ {
			if a.Tiles[q][r] != b.Tiles[q][r] {
				return false
			}
		}
	}
	return true
}
