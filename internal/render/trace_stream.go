package render

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jonas/reaktor-sim/internal/sim"
)

// TraceStreamWriter writes trace PNGs step-by-step to avoid holding all snapshots in memory.
type TraceStreamWriter struct {
	runDir        string
	pngName       func(index int) string
	captionMeta   func(snap sim.Snapshot) string
	events        []traceStepYAML
	stepIndex     int
	lastBoard     *sim.Snapshot
	lastEvent     string
	meta          TraceMeta
	run           int
	monteCarloRun int
	pngErrs       []error
}

func newTraceStreamWriter(outDir, runDirName string, run, monteCarloRun int, meta TraceMeta, pngName func(index int) string, captionMeta func(snap sim.Snapshot) string) (*TraceStreamWriter, error) {
	runDir := filepath.Join(outDir, runDirName)
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return nil, err
	}
	return &TraceStreamWriter{
		runDir:        runDir,
		pngName:       pngName,
		captionMeta:   captionMeta,
		meta:          meta,
		run:           run,
		monteCarloRun: monteCarloRun,
	}, nil
}

// NewRunTraceStreamWriter streams the primary run trace into runN/.
func NewRunTraceStreamWriter(outDir string, run int, meta TraceMeta) (*TraceStreamWriter, error) {
	return newTraceStreamWriter(outDir, fmt.Sprintf("run%d", run), run, 0, meta,
		func(i int) string { return fmt.Sprintf("graph_run%d_%03d.png", run, i) },
		func(snap sim.Snapshot) string {
			return fmt.Sprintf("Run %d | Schritt %d (Warteschlange: %d)", run, snap.Step, snap.QueueSize)
		},
	)
}

// NewLoopTraceStreamWriter streams a loop trace into loopN/.
func NewLoopTraceStreamWriter(outDir string, loopNum, monteCarloRun int, meta TraceMeta) (*TraceStreamWriter, error) {
	return newTraceStreamWriter(outDir, fmt.Sprintf("loop%d", loopNum), loopNum, monteCarloRun, meta,
		func(i int) string { return fmt.Sprintf("graph_loop%d_%03d.png", loopNum, i) },
		func(snap sim.Snapshot) string {
			return fmt.Sprintf("Loop-Trace %d | MC-Lauf %d | Schritt %d (Warteschlange: %d)",
				loopNum, monteCarloRun, snap.Step, snap.QueueSize)
		},
	)
}

// NewWinTraceStreamWriter streams a win trace into winN/.
func NewWinTraceStreamWriter(outDir string, winNum, monteCarloRun int, meta TraceMeta) (*TraceStreamWriter, error) {
	return newTraceStreamWriter(outDir, fmt.Sprintf("win%d", winNum), winNum, monteCarloRun, meta,
		func(i int) string { return fmt.Sprintf("graph_win%d_%03d.png", winNum, i) },
		func(snap sim.Snapshot) string {
			return fmt.Sprintf("Win-Trace %d | MC-Lauf %d | Schritt %d (Warteschlange: %d)",
				winNum, monteCarloRun, snap.Step, snap.QueueSize)
		},
	)
}

// Record implements sim trace streaming callbacks.
func (w *TraceStreamWriter) Record(snap sim.Snapshot) error {
	if w == nil {
		return nil
	}
	metaLine := w.captionMeta(snap)
	caption := ASCII(snap.Narrative) + "\n" + metaLine
	if err := WriteGraphPNG(snap.Board, snap.Graph, filepath.Join(w.runDir, w.pngName(w.stepIndex)), caption, ChipView{
		Queue:  snap.Queue,
		Active: snap.Active,
	}); err != nil {
		w.pngErrs = append(w.pngErrs, err)
	}
	w.events = append(w.events, traceStepYAML{
		Index:     w.stepIndex,
		Step:      snap.Step,
		Event:     snap.Event,
		Narrative: ASCII(snap.Narrative),
		QueueSize: snap.QueueSize,
	})
	w.stepIndex++
	w.lastEvent = snap.Event
	if snap.Board != nil {
		w.lastBoard = &snap
	}
	return nil
}

// Finish writes trace.yaml and returns any PNG errors encountered while streaming.
func (w *TraceStreamWriter) Finish() error {
	if w == nil {
		return nil
	}
	doc := traceYAML{
		Run:           w.run,
		MonteCarloRun: w.monteCarloRun,
		Setup:         buildTraceSetupYAML(w.meta),
		Steps:         len(w.events),
		Events:        w.events,
	}
	if w.lastBoard != nil {
		doc.Outcome = traceOutcomeNote(w.lastEvent, w.lastBoard.Board)
	}
	tracePath := filepath.Join(w.runDir, "trace.yaml")
	if err := writeYAML(tracePath, doc); err != nil {
		return fmt.Errorf("trace.yaml: %w", err)
	}
	w.events = nil
	w.lastBoard = nil
	return errors.Join(w.pngErrs...)
}

// Steps returns the number of streamed snapshots.
func (w *TraceStreamWriter) Steps() int {
	if w == nil {
		return 0
	}
	return w.stepIndex
}

// Dir returns the absolute trace directory.
func (w *TraceStreamWriter) Dir() string {
	if w == nil {
		return ""
	}
	dir, err := filepath.Abs(w.runDir)
	if err != nil {
		return w.runDir
	}
	return dir
}
