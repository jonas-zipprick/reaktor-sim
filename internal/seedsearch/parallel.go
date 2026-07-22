package seedsearch

import (
	"encoding/gob"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/rules"
)

const spillCheckInterval = 64

func scanWorkers() int {
	n := runtime.GOMAXPROCS(0)
	if n < 1 {
		return 1
	}
	return n
}

type scanTracker struct {
	progress  ProgressFunc
	done      atomic.Int64
	totalWork int64
	shift     int32
	skipped   atomic.Int64
	errMu     sync.Mutex
	err       error
}

func newScanTracker(progress ProgressFunc, total int64) *scanTracker {
	if total < 1 {
		total = 1
	}
	return &scanTracker{progress: progress, totalWork: total}
}

func (t *scanTracker) setShift(shift int) {
	t.shift = int32(shift)
	t.report()
}

func (t *scanTracker) finish(dup bool) {
	if dup {
		t.skipped.Add(1)
	}
	t.done.Add(1)
	t.report()
}

func (t *scanTracker) report() {
	if t.progress == nil {
		return
	}
	t.progress(t.done.Load(), t.totalWork, int(t.shift))
}

func (t *scanTracker) setErr(err error) {
	if err == nil {
		return
	}
	t.errMu.Lock()
	if t.err == nil {
		t.err = err
	}
	t.errMu.Unlock()
}

func (t *scanTracker) error() error {
	t.errMu.Lock()
	defer t.errMu.Unlock()
	return t.err
}

type outcomeCollector struct {
	mu             sync.Mutex
	seen           map[string]struct{}
	items          []Outcome
	opts           Options
	shift          int
	spillPath      string
	spillFile      *os.File
	spillEnc       *gob.Encoder
	totalCount     int
	addsSinceCheck int
}

func newOutcomeCollector(opts Options, shift int) *outcomeCollector {
	return &outcomeCollector{
		seen:  make(map[string]struct{}),
		opts:  opts,
		shift: shift,
	}
}

func (c *outcomeCollector) tryClaim(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, dup := c.seen[key]; dup {
		return false
	}
	c.seen[key] = struct{}{}
	return true
}

func (c *outcomeCollector) add(o Outcome) error {
	c.mu.Lock()
	c.items = append(c.items, o)
	c.totalCount++
	c.addsSinceCheck++
	shouldCheck := c.opts.SpillDir != "" && c.addsSinceCheck >= spillCheckInterval
	c.mu.Unlock()
	if shouldCheck {
		return c.maybeFlush()
	}
	return nil
}

func (c *outcomeCollector) maybeFlush() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.addsSinceCheck = 0
	if c.opts.SpillDir == "" || len(c.items) == 0 {
		return nil
	}
	if !shouldSpillToDisk(c.opts) {
		return nil
	}
	return c.flushLocked()
}

func (c *outcomeCollector) flushLocked() error {
	if len(c.items) == 0 {
		return nil
	}
	if c.spillPath == "" {
		if err := os.MkdirAll(c.opts.SpillDir, 0o755); err != nil {
			return fmt.Errorf("spill dir %s: %w", c.opts.SpillDir, err)
		}
		c.spillPath = filepath.Join(c.opts.SpillDir, fmt.Sprintf("shift%d.gob", c.shift))
	}
	if c.spillFile == nil {
		f, err := os.OpenFile(c.spillPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return fmt.Errorf("spill open %s: %w", c.spillPath, err)
		}
		c.spillFile = f
		c.spillEnc = gob.NewEncoder(f)
	}
	if err := c.spillEnc.Encode(c.items); err != nil {
		return fmt.Errorf("spill encode %s: %w", c.spillPath, err)
	}
	c.items = c.items[:0]
	releaseMemory()
	return nil
}

func (c *outcomeCollector) closeSpill() {
	if c.spillFile != nil {
		_ = c.spillFile.Close()
		c.spillFile = nil
		c.spillEnc = nil
	}
}

func (c *outcomeCollector) finish() (ShiftResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.opts.SpillDir != "" && shouldSpillToDisk(c.opts) {
		if err := c.flushLocked(); err != nil {
			return ShiftResult{}, err
		}
	}
	c.closeSpill()
	sr := ShiftResult{
		Shift:        c.shift,
		spillPath:    c.spillPath,
		outcomeCount: c.totalCount,
	}
	if c.spillPath == "" {
		sr.Outcomes = append([]Outcome(nil), c.items...)
	}
	return sr, nil
}

func scanShift1(from, to int64, opts Options, tracker *scanTracker) (ShiftResult, error) {
	collector := newOutcomeCollector(opts, 1)
	workers := opts.workerCount()
	jobs := make(chan int64)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for seed := range jobs {
				if tracker.error() != nil {
					return
				}
				rng := rand.New(rand.NewSource(seed))
				state, endLeft, shift1Prev, err := prepareShift1Board(rng, opts)
				if err != nil {
					tracker.setErr(fmt.Errorf("seed %d: %w", seed, err))
					return
				}
				fp := board.Fingerprint(state)
				if !collector.tryClaim(fp) {
					tracker.finish(true)
					continue
				}
				out := evaluateShift(seed, state, opts, 1, shift1Prev, [4]int{}, [4]int{}, 0, board.PlayerLeftover{}, endLeft)
				if err := collector.add(out); err != nil {
					tracker.setErr(err)
					return
				}
				tracker.finish(false)
			}
		}()
	}
	go func() {
		for seed := from; seed <= to; seed++ {
			jobs <- seed
		}
		close(jobs)
	}()
	wg.Wait()
	if err := tracker.error(); err != nil {
		return ShiftResult{}, err
	}
	return collector.finish()
}

type branchJob struct {
	parent parentBoard
	seed   int64
}

func scanShiftBranch(k int, from, to int64, parents []parentBoard, opts Options, tracker *scanTracker) (ShiftResult, error) {
	collector := newOutcomeCollector(opts, k)
	workers := opts.workerCount()
	work := make(chan branchJob)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range work {
				if tracker.error() != nil {
					return
				}
				carryFP := job.parent.carryFP
				if carryFP == "" {
					carryFP = job.parent.prevFP
				}
				base, err := board.FromFingerprint(carryFP)
				if err != nil {
					tracker.setErr(fmt.Errorf("schicht %d, board %s: %w", k, carryFP, err))
					return
				}
				base.Damage = job.parent.damage
				base.EmitterDamage = job.parent.emitterDamage
				rng := rand.New(rand.NewSource(job.seed))
				budgetP1 := opts.Finance.ReactorBudget + job.parent.leftover.Player1
				budgetP2 := opts.Finance.GridBudget + job.parent.leftover.Player2
				spendRes, err := board.SpendShiftBudget(rng, base, budgetP1, budgetP2, opts.MonthFilter, 0, rules.Month{
					EnergyID:  opts.EnergyCard.ID,
					FinanceID: opts.Finance.ID,
				})
				if err != nil {
					tracker.setErr(fmt.Errorf("schicht %d: %w", k, err))
					return
				}
				key := board.Fingerprint(base) + carryKey(job.parent.demand, job.parent.damage, job.parent.emitterDamage, job.parent.leftover)
				if !collector.tryClaim(key) {
					tracker.finish(true)
					continue
				}
				out := evaluateShift(job.seed, base, opts, k, job.parent.prevFP, job.parent.demand, job.parent.damage, job.parent.emitterDamage, job.parent.leftover, spendRes)
				if err := collector.add(out); err != nil {
					tracker.setErr(err)
					return
				}
				tracker.finish(false)
			}
		}()
	}
	go func() {
		for _, p := range parents {
			for seed := from; seed <= to; seed++ {
				work <- branchJob{parent: p, seed: seed}
			}
		}
		close(work)
	}()
	wg.Wait()
	if err := tracker.error(); err != nil {
		return ShiftResult{}, err
	}
	return collector.finish()
}
