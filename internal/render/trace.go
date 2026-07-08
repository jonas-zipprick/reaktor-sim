package render

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jonas/reaktor-sim/internal/sim"
)

// WriteRunTrace saves trace.txt and one graph PNG per simulation snapshot.
// trace.txt is written before PNGs so each run folder always has a log even
// when image export fails or is interrupted.
func WriteRunTrace(run int, snaps []sim.Snapshot, outDir string) error {
	runDir := filepath.Join(outDir, fmt.Sprintf("run%d", run))
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return err
	}

	summary := formatRunTrace(run, snaps)
	tracePath := filepath.Join(runDir, "trace.txt")
	if err := os.WriteFile(tracePath, []byte(summary), 0o644); err != nil {
		return fmt.Errorf("trace.txt: %w", err)
	}

	var pngErrs []error
	for i, snap := range snaps {
		name := fmt.Sprintf("graph_run%d_%03d.png", run, i)
		meta := fmt.Sprintf("Run %d | Schritt %d (Warteschlange: %d)", run, snap.Step, snap.QueueSize)
		caption := ASCII(snap.Narrative) + "\n" + meta
		if err := WriteGraphPNG(snap.Board, snap.Graph, filepath.Join(runDir, name), caption, ChipView{
			Queue:  snap.Queue,
			Active: snap.Active,
		}); err != nil {
			pngErrs = append(pngErrs, fmt.Errorf("%s: %w", name, err))
		}
	}
	return errors.Join(pngErrs...)
}

// TraceIndexEntry records one traced Monte-Carlo run.
type TraceIndexEntry struct {
	Run   int
	Steps int
	Dir   string
}

// WriteTraceIndex lists all trace run folders at outDir/trace_index.txt.
func WriteTraceIndex(outDir string, entries []TraceIndexEntry) error {
	if len(entries) == 0 {
		return nil
	}
	var b strings.Builder
	b.WriteString("Trace-Laeufe (output/runN/)\n\n")
	for _, e := range entries {
		b.WriteString(fmt.Sprintf("run%d: %d Schritte — %s/trace.txt\n", e.Run, e.Steps, e.Dir))
	}
	path := filepath.Join(outDir, "trace_index.txt")
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func formatRunTrace(run int, snaps []sim.Snapshot) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Run %d – Simulationstrace (%d Schritte)\n", run, len(snaps)))
	if len(snaps) == 0 {
		b.WriteString("\nKeine Snapshots aufgezeichnet.\n")
		return b.String()
	}
	if note := traceOutcomeNote(snaps[len(snaps)-1].Event); note != "" {
		b.WriteString(note + "\n")
	}
	b.WriteByte('\n')

	for _, snap := range snaps {
		meta := fmt.Sprintf("Run %d | Schritt %d (Warteschlange: %d)", run, snap.Step, snap.QueueSize)
		b.WriteString(ASCII(snap.Narrative) + "\n" + meta + "\n")
	}
	return b.String()
}

func traceOutcomeNote(lastEvent string) string {
	switch lastEvent {
	case "verloren":
		return "Ergebnis: Kritische Masse ueberschritten — Simulation sofort beendet."
	case "Schrittlimit":
		return "Ergebnis: Schrittlimit erreicht — Simulation abgebrochen."
	case "Ende":
		return "Ergebnis: Schicht normal beendet (Warteschlange leer)."
	default:
		return ""
	}
}
