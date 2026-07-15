package seedsearch

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
)

// LoadOutcomes returns all outcomes for this shift, loading from disk when spilled.
func (sr ShiftResult) LoadOutcomes() ([]Outcome, error) {
	if len(sr.Outcomes) > 0 {
		return sr.Outcomes, nil
	}
	if sr.spillPath == "" {
		return nil, nil
	}
	return readOutcomes(sr.spillPath)
}

// Len returns the number of outcomes without loading spilled data from disk.
func (sr ShiftResult) Len() int {
	if len(sr.Outcomes) > 0 {
		return len(sr.Outcomes)
	}
	return sr.outcomeCount
}

// Cleanup removes the on-disk spill directory created during Scan.
func (r ScanResult) Cleanup() error {
	if r.spillDir == "" {
		return nil
	}
	return os.RemoveAll(r.spillDir)
}

func processMemoryUse() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// Sys tracks total memory obtained from the OS and correlates better with
	// Task Manager / RSS than heap in-use alone.
	return m.Sys
}

func releaseMemory() {
	runtime.GC()
	debug.FreeOSMemory()
}

func shouldSpillToDisk(opts Options) bool {
	if opts.SpillDir == "" {
		return false
	}
	if opts.SpillMemoryThreshold == 0 {
		return true
	}
	return processMemoryUse() >= opts.SpillMemoryThreshold
}

func spillShiftResult(sr *ShiftResult, opts Options) error {
	if len(sr.Outcomes) == 0 {
		if sr.spillPath != "" && sr.outcomeCount == 0 {
			sr.outcomeCount = countSpilledOutcomes(sr.spillPath)
		}
		return nil
	}
	if sr.outcomeCount == 0 {
		sr.outcomeCount = len(sr.Outcomes)
	}
	if !shouldSpillToDisk(opts) {
		return nil
	}
	if err := os.MkdirAll(opts.SpillDir, 0o755); err != nil {
		return fmt.Errorf("spill dir %s: %w", opts.SpillDir, err)
	}
	path := filepath.Join(opts.SpillDir, fmt.Sprintf("shift%d.gob", sr.Shift))
	if err := writeOutcomes(path, sr.Outcomes); err != nil {
		return err
	}
	sr.Outcomes = nil
	sr.spillPath = path
	releaseMemory()
	return nil
}

func prunePreviousShift(sr *ShiftResult, child ShiftResult, opts Options) error {
	needed, err := collectNeededParentFPs(child)
	if err != nil {
		return err
	}
	if len(needed) == 0 {
		return nil
	}
	pruned, err := filterOutcomes(*sr, needed)
	if err != nil {
		return err
	}
	sr.Outcomes = pruned
	sr.outcomeCount = len(pruned)
	sr.spillPath = ""
	if !shouldSpillToDisk(opts) {
		return nil
	}
	return spillShiftResult(sr, opts)
}

func collectNeededParentFPs(child ShiftResult) (map[string]struct{}, error) {
	needed := make(map[string]struct{})
	if len(child.Outcomes) > 0 {
		for _, o := range child.Outcomes {
			if o.PrevBoardFingerprint != "" {
				needed[o.PrevBoardFingerprint] = struct{}{}
			}
		}
		return needed, nil
	}
	if child.spillPath == "" {
		return needed, nil
	}
	err := readOutcomeBatches(child.spillPath, func(batch []Outcome) error {
		for _, o := range batch {
			if o.PrevBoardFingerprint != "" {
				needed[o.PrevBoardFingerprint] = struct{}{}
			}
		}
		return nil
	})
	return needed, err
}

func filterOutcomes(sr ShiftResult, needed map[string]struct{}) ([]Outcome, error) {
	if len(sr.Outcomes) > 0 {
		pruned := make([]Outcome, 0, len(needed))
		for _, o := range sr.Outcomes {
			if _, ok := needed[o.BoardFingerprint]; ok {
				pruned = append(pruned, o)
			}
		}
		return pruned, nil
	}
	if sr.spillPath == "" {
		return nil, nil
	}
	pruned := make([]Outcome, 0, len(needed))
	err := readOutcomeBatches(sr.spillPath, func(batch []Outcome) error {
		for _, o := range batch {
			if _, ok := needed[o.BoardFingerprint]; ok {
				pruned = append(pruned, o)
			}
		}
		return nil
	})
	return pruned, err
}

func countSpilledOutcomes(path string) int {
	n := 0
	_ = readOutcomeBatches(path, func(batch []Outcome) error {
		n += len(batch)
		return nil
	})
	return n
}

func writeOutcomes(path string, outcomes []Outcome) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("spill write %s: %w", path, err)
	}
	defer f.Close()
	if err := gob.NewEncoder(f).Encode(outcomes); err != nil {
		return fmt.Errorf("spill encode %s: %w", path, err)
	}
	return nil
}

func writeOutcomeBatches(path string, batches [][]Outcome) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("spill write %s: %w", path, err)
	}
	defer f.Close()
	enc := gob.NewEncoder(f)
	for _, batch := range batches {
		if err := enc.Encode(batch); err != nil {
			return fmt.Errorf("spill encode %s: %w", path, err)
		}
	}
	return nil
}

func readOutcomeBatches(path string, fn func([]Outcome) error) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("spill read %s: %w", path, err)
	}
	defer f.Close()
	dec := gob.NewDecoder(f)
	for {
		var batch []Outcome
		err := dec.Decode(&batch)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("spill decode %s: %w", path, err)
		}
		if err := fn(batch); err != nil {
			return err
		}
	}
}

func readOutcomes(path string) ([]Outcome, error) {
	var outcomes []Outcome
	err := readOutcomeBatches(path, func(batch []Outcome) error {
		outcomes = append(outcomes, batch...)
		return nil
	})
	return outcomes, err
}
