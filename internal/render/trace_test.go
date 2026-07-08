package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/graph"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestWriteRunTraceAlwaysCreatesTraceTxt(t *testing.T) {
	dir := t.TempDir()
	snaps := []sim.Snapshot{{
		Step:      0,
		Event:     "Start",
		Narrative: "Schichtbeginn.",
		Board:     board.NewEmpty(),
		Graph:     graph.BuildFlow(board.NewEmpty(), nil),
	}}

	if err := WriteRunTrace(1, snaps, dir); err != nil {
		t.Fatalf("WriteRunTrace: %v", err)
	}

	tracePath := filepath.Join(dir, "run1", "trace.txt")
	data, err := os.ReadFile(tracePath)
	if err != nil {
		t.Fatalf("read trace.txt: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, "Run 1") {
		t.Fatalf("trace.txt missing run header: %q", text)
	}
	if !strings.Contains(text, "Schichtbeginn") {
		t.Fatalf("trace.txt missing narrative: %q", text)
	}
}

func TestWriteRunTraceEmptySnapshots(t *testing.T) {
	dir := t.TempDir()
	if err := WriteRunTrace(1, nil, dir); err != nil {
		t.Fatalf("WriteRunTrace: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "run1", "trace.txt"))
	if err != nil {
		t.Fatalf("read trace.txt: %v", err)
	}
	if !strings.Contains(string(data), "Keine Snapshots") {
		t.Fatal("expected empty-snapshot note in trace.txt")
	}
}

func TestWriteTraceIndex(t *testing.T) {
	dir := t.TempDir()
	entries := []TraceIndexEntry{
		{Run: 1, Steps: 8, Dir: filepath.Join(dir, "run1")},
		{Run: 2, Steps: 5, Dir: filepath.Join(dir, "run2")},
	}
	if err := WriteTraceIndex(dir, entries); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "trace_index.txt"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, "run1:") || !strings.Contains(text, "run2:") {
		t.Fatalf("index = %q", text)
	}
}

func TestFormatRunTraceOutcomeNote(t *testing.T) {
	snaps := []sim.Snapshot{{Event: "verloren", Narrative: "verloren"}}
	text := formatRunTrace(3, snaps)
	if !strings.Contains(text, "Kritische Masse") {
		t.Fatalf("missing outcome note: %q", text)
	}
}
