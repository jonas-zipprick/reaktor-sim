package render

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/sim"
)

// TraceMeta records simulation setup written into trace.yaml.
type TraceMeta struct {
	EnergyCardID    string
	EnergyCardName  string
	EnergyCardLevel int
	Shift           int
	Costs           board.PlayerCosts
}

// WriteRunTrace saves trace.yaml and one graph PNG per simulation snapshot.
// trace.yaml is written before PNGs so each run folder always has a log even
// when image export fails or is interrupted.
func WriteRunTrace(run int, meta TraceMeta, snaps []sim.Snapshot, outDir string) error {
	runDir := filepath.Join(outDir, fmt.Sprintf("run%d", run))
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return err
	}

	tracePath := filepath.Join(runDir, "trace.yaml")
	if err := writeYAML(tracePath, buildTraceYAML(run, 0, meta, snaps)); err != nil {
		return fmt.Errorf("trace.yaml: %w", err)
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

// WriteLoopTrace saves trace.yaml and graph PNGs for a Monte-Carlo run that
// hit the step limit. loopNum is the sequential loop-trace index (1-based);
// monteCarloRun is the original Monte-Carlo run number (1-based).
func WriteLoopTrace(loopNum, monteCarloRun int, meta TraceMeta, snaps []sim.Snapshot, outDir string) error {
	runDir := filepath.Join(outDir, fmt.Sprintf("loop%d", loopNum))
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return err
	}

	tracePath := filepath.Join(runDir, "trace.yaml")
	if err := writeYAML(tracePath, buildTraceYAML(loopNum, monteCarloRun, meta, snaps)); err != nil {
		return fmt.Errorf("trace.yaml: %w", err)
	}

	var pngErrs []error
	for i, snap := range snaps {
		name := fmt.Sprintf("graph_loop%d_%03d.png", loopNum, i)
		meta := fmt.Sprintf("Loop-Trace %d | MC-Lauf %d | Schritt %d (Warteschlange: %d)",
			loopNum, monteCarloRun, snap.Step, snap.QueueSize)
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

// WriteWinTrace saves trace.yaml and graph PNGs for a Monte-Carlo run where all
// demands were fulfilled. winNum is the sequential win-trace index (1-based);
// monteCarloRun is the original Monte-Carlo run number (1-based).
func WriteWinTrace(winNum, monteCarloRun int, meta TraceMeta, snaps []sim.Snapshot, outDir string) error {
	runDir := filepath.Join(outDir, fmt.Sprintf("win%d", winNum))
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return err
	}

	tracePath := filepath.Join(runDir, "trace.yaml")
	if err := writeYAML(tracePath, buildTraceYAML(winNum, monteCarloRun, meta, snaps)); err != nil {
		return fmt.Errorf("trace.yaml: %w", err)
	}

	var pngErrs []error
	for i, snap := range snaps {
		name := fmt.Sprintf("graph_win%d_%03d.png", winNum, i)
		captionMeta := fmt.Sprintf("Win-Trace %d | MC-Lauf %d | Schritt %d (Warteschlange: %d)",
			winNum, monteCarloRun, snap.Step, snap.QueueSize)
		caption := ASCII(snap.Narrative) + "\n" + captionMeta
		if err := WriteGraphPNG(snap.Board, snap.Graph, filepath.Join(runDir, name), caption, ChipView{
			Queue:  snap.Queue,
			Active: snap.Active,
		}); err != nil {
			pngErrs = append(pngErrs, fmt.Errorf("%s: %w", name, err))
		}
	}
	return errors.Join(pngErrs...)
}

// TraceKind labels how a trace was selected in trace_index.yaml.
type TraceKind string

const (
	TraceKindFirst TraceKind = "trace-first"
	TraceKindLoop  TraceKind = "trace-loop"
	TraceKindWin   TraceKind = "trace-win"
)

// TraceIndexEntry records one traced Monte-Carlo run.
type TraceIndexEntry struct {
	Run           int
	MonteCarloRun int
	Kind          TraceKind
	Steps         int
	Dir           string
}

// WriteTraceIndex lists all trace run folders at outDir/trace_index.yaml.
func WriteTraceIndex(outDir string, entries []TraceIndexEntry) error {
	if len(entries) == 0 {
		return nil
	}
	return writeYAML(filepath.Join(outDir, "trace_index.yaml"), buildTraceIndexYAML(entries))
}

func traceOutcomeNote(lastEvent string, boardState *board.State) string {
	var base string
	switch lastEvent {
	case "verloren":
		base = "Kritische Masse ueberschritten — Simulation sofort beendet."
	case "Schrittlimit":
		base = "Schrittlimit erreicht — Simulation abgebrochen."
	case "Ende":
		base = "Schicht normal beendet (Warteschlange leer)."
	default:
		return ""
	}
	if boardState == nil {
		return base
	}
	return base + " Uebrige Bedarfe: " + formatRemainingDemands(boardState) + ". Schaden: " + formatRemainingDamage(boardState) + "."
}

func formatRemainingDamage(state *board.State) string {
	zones := []board.Zone{
		board.ZoneIndustry,
		board.ZoneResidential,
		board.ZoneRail,
		board.ZonePlant,
	}
	parts := make([]string, 0, len(zones))
	for _, z := range zones {
		if n := state.TotalDamage(z); n > 0 {
			parts = append(parts, fmt.Sprintf("%s=%d", z.String(), n))
		}
	}
	if len(parts) == 0 {
		return "keiner"
	}
	return strings.Join(parts, ", ")
}

func formatRemainingDemands(state *board.State) string {
	zones := []board.Zone{
		board.ZoneIndustry,
		board.ZoneResidential,
		board.ZoneRail,
		board.ZonePlant,
	}
	parts := make([]string, 0, len(zones))
	for _, z := range zones {
		if n := state.TotalDemand(z); n > 0 {
			parts = append(parts, fmt.Sprintf("%s=%d", z.String(), n))
		}
	}
	if len(parts) == 0 {
		return "keine"
	}
	return strings.Join(parts, ", ")
}
