package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPrepareChartsDirClearsPreviousOutputIncludingSpill(t *testing.T) {
	dir := t.TempDir()

	oldChart := filepath.Join(dir, "schicht_1")
	if err := os.MkdirAll(oldChart, 0o755); err != nil {
		t.Fatal(err)
	}
	spillDir := filepath.Join(dir, ".spill")
	if err := os.MkdirAll(spillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	spillFile := filepath.Join(spillDir, "shift1.gob")
	if err := os.WriteFile(spillFile, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := prepareChartsDir(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(oldChart); err == nil {
		t.Fatal("old chart dir should be removed")
	}
	if _, err := os.Stat(spillFile); err == nil {
		t.Fatal("stale spill file should be removed before a new scan")
	}
}
