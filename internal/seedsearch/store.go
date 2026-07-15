package seedsearch

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
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

func spillShiftResult(sr *ShiftResult, spillDir string) error {
	if spillDir == "" || len(sr.Outcomes) == 0 {
		return nil
	}
	if err := os.MkdirAll(spillDir, 0o755); err != nil {
		return fmt.Errorf("spill dir %s: %w", spillDir, err)
	}
	path := filepath.Join(spillDir, fmt.Sprintf("shift%d.gob", sr.Shift))
	if err := writeOutcomes(path, sr.Outcomes); err != nil {
		return err
	}
	sr.outcomeCount = len(sr.Outcomes)
	sr.Outcomes = nil
	sr.spillPath = path
	return nil
}

func prunePreviousShift(sr *ShiftResult, childOutcomes []Outcome, spillDir string) error {
	prev, err := sr.LoadOutcomes()
	if err != nil {
		return err
	}
	pruned := pruneShiftOutcomes(prev, childOutcomes)
	sr.Outcomes = pruned
	sr.outcomeCount = len(pruned)
	if spillDir == "" {
		return nil
	}
	return spillShiftResult(sr, spillDir)
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

func readOutcomes(path string) ([]Outcome, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("spill read %s: %w", path, err)
	}
	defer f.Close()
	var outcomes []Outcome
	if err := gob.NewDecoder(f).Decode(&outcomes); err != nil {
		return nil, fmt.Errorf("spill decode %s: %w", path, err)
	}
	return outcomes, nil
}
