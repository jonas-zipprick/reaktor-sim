package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestShiftBoardDirsLookupByPrevBoard(t *testing.T) {
	m := make(shiftBoardDirs)
	prevFP := "b2_prevboard"
	prevDir := "/tmp/Schicht 4 (seed27176, b2_prevboard)"
	m.register(4, prevFP, prevDir)

	got, ok := m.lookup(4, prevFP)
	if !ok || got != prevDir {
		t.Fatalf("lookup = %q, %v; want %q, true", got, ok, prevDir)
	}
	if _, ok := m.lookup(4, "b2_other"); ok {
		t.Fatal("unexpected hit for other fingerprint")
	}
}

func TestLinkToPrevShift(t *testing.T) {
	dir := t.TempDir()
	prev := filepath.Join(dir, "Schicht 4 (seed1, board4)")
	cur := filepath.Join(dir, "Schicht 5 (seed1, board5)")
	if err := os.MkdirAll(prev, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cur, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := linkToPrevShift(cur, prev); err != nil {
		t.Fatal(err)
	}

	linkPath := filepath.Join(cur, prevShiftLinkName)
	prevAbs, err := filepath.Abs(prev)
	if err != nil {
		t.Fatal(err)
	}

	resolved, err := resolvePrevShiftDir(linkPath)
	if err != nil {
		t.Fatal(err)
	}
	resolvedAbs, err := filepath.Abs(resolved)
	if err != nil {
		t.Fatal(err)
	}
	if resolvedAbs != prevAbs {
		t.Fatalf("resolved = %q, want %q", resolvedAbs, prevAbs)
	}

	if runtime.GOOS != "windows" {
		target, err := os.Readlink(linkPath)
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join("..", filepath.Base(prev))
		if target != want {
			t.Fatalf("symlink target = %q, want %q", target, want)
		}
	}
}
