package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/graph"
	"github.com/jonas/reaktor-sim/internal/sim"
	"gopkg.in/yaml.v3"
)

func TestWriteRunTraceAlwaysCreatesTraceYAML(t *testing.T) {
	dir := t.TempDir()
	b := board.NewEmpty()
	b.ApplyDemands(board.ShiftDemands{Industry: 1, Plant: 1})
	snaps := []sim.Snapshot{{
		Step:      0,
		Event:     "Start",
		Narrative: "Schichtbeginn.",
		Board:     b,
		Graph:     graph.BuildFlow(b, nil),
	}}
	meta := TraceMeta{
		EnergyCardID:    "eroeffnungsfeier",
		EnergyCardName:  "Eroeffnungsfeier",
		EnergyCardLevel: 1,
		Shift:           1,
		Costs:           b.PlayerCosts(),
	}

	if err := WriteRunTrace(1, meta, snaps, dir); err != nil {
		t.Fatalf("WriteRunTrace: %v", err)
	}

	tracePath := filepath.Join(dir, "run1", "trace.yaml")
	data, err := os.ReadFile(tracePath)
	if err != nil {
		t.Fatalf("read trace.yaml: %v", err)
	}
	var doc traceYAML
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse trace.yaml: %v", err)
	}
	if doc.Run != 1 || len(doc.Events) != 1 {
		t.Fatalf("doc = %+v", doc)
	}
	if doc.Setup.EnergyCard.ID != "eroeffnungsfeier" || doc.Setup.Shift != 1 {
		t.Fatalf("setup = %+v", doc.Setup)
	}
	if doc.Events[0].Narrative != "Schichtbeginn." {
		t.Fatalf("narrative = %q", doc.Events[0].Narrative)
	}
}

func TestWriteRunTraceEmptySnapshots(t *testing.T) {
	dir := t.TempDir()
	meta := TraceMeta{EnergyCardID: "test", Shift: 1}
	if err := WriteRunTrace(1, meta, nil, dir); err != nil {
		t.Fatalf("WriteRunTrace: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "run1", "trace.yaml"))
	if err != nil {
		t.Fatalf("read trace.yaml: %v", err)
	}
	var doc traceYAML
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Steps != 0 || len(doc.Events) != 0 {
		t.Fatalf("doc = %+v", doc)
	}
}

func TestWriteTraceIndexYAML(t *testing.T) {
	dir := t.TempDir()
	entries := []TraceIndexEntry{
		{Run: 1, Kind: TraceKindFirst, Steps: 8, Dir: filepath.Join(dir, "run1")},
		{Run: 1, Kind: TraceKindLoop, MonteCarloRun: 5, Steps: 101, Dir: filepath.Join(dir, "loop1")},
	}
	if err := WriteTraceIndex(dir, entries); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "trace_index.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var doc traceIndexYAML
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Runs) != 2 || doc.Runs[0].Run != 1 {
		t.Fatalf("index = %+v", doc)
	}
	if doc.Runs[0].Kind != string(TraceKindFirst) {
		t.Fatalf("first kind = %q, want %q", doc.Runs[0].Kind, TraceKindFirst)
	}
	if doc.Runs[1].Kind != string(TraceKindLoop) || doc.Runs[1].MonteCarloRun != 5 {
		t.Fatalf("loop entry = %+v", doc.Runs[1])
	}
}

func TestTraceDocumentOutcome(t *testing.T) {
	b := board.NewEmpty()
	b.ApplyDemands(board.ShiftDemands{Industry: 1, Residential: 1})
	snaps := []sim.Snapshot{{Event: "Ende", Narrative: "Ende", Board: b}}
	meta := TraceMeta{EnergyCardID: "test", Shift: 1}
	doc := buildTraceYAML(3, 0, meta, snaps)
	if !strings.Contains(doc.Outcome, "Schicht normal beendet") {
		t.Fatalf("outcome = %q", doc.Outcome)
	}
	if !strings.Contains(doc.Outcome, "Industrie=1") || !strings.Contains(doc.Outcome, "Wohnviertel=1") {
		t.Fatalf("outcome missing demands: %q", doc.Outcome)
	}

	snapsDone := []sim.Snapshot{{Event: "Ende", Narrative: "Ende", Board: board.NewEmpty()}}
	docDone := buildTraceYAML(1, 0, meta, snapsDone)
	if !strings.Contains(docDone.Outcome, "keine") {
		t.Fatalf("outcome = %q, want keine remaining demands", docDone.Outcome)
	}

	snapsLost := []sim.Snapshot{{Event: "verloren", Narrative: "verloren"}}
	docLost := buildTraceYAML(2, 0, meta, snapsLost)
	if !strings.Contains(docLost.Outcome, "Kritische Masse") {
		t.Fatalf("outcome = %q", docLost.Outcome)
	}
}

func TestWriteBoardYAML(t *testing.T) {
	dir := t.TempDir()
	state := board.NewEmpty()
	state.ApplyDemands(board.ShiftDemands{Industry: 1, Rail: 2})
	path := filepath.Join(dir, "spielfeld.yaml")
	if err := WriteBoardYAML(state, path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var doc boardYAML
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Demands) < 2 {
		t.Fatalf("demands = %+v", doc.Demands)
	}
}
