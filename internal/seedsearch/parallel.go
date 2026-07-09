package seedsearch

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/jonas/reaktor-sim/internal/board"
)

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
	mu    sync.Mutex
	seen  map[string]struct{}
	items []Outcome
}

func newOutcomeCollector() *outcomeCollector {
	return &outcomeCollector{seen: make(map[string]struct{})}
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

func (c *outcomeCollector) add(o Outcome) {
	c.mu.Lock()
	c.items = append(c.items, o)
	c.mu.Unlock()
}

func (c *outcomeCollector) snapshot() []Outcome {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]Outcome(nil), c.items...)
}

func scanShift1(from, to int64, opts Options, tracker *scanTracker) ([]Outcome, error) {
	seeds := make([]int64, 0, to-from+1)
	for seed := from; seed <= to; seed++ {
		seeds = append(seeds, seed)
	}
	collector := newOutcomeCollector()
	workers := scanWorkers()
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
				state, endLeft, err := buildInitialBoard(rng, opts)
				if err != nil {
					tracker.setErr(fmt.Errorf("seed %d: %w", seed, err))
					return
				}
				fp := board.Fingerprint(state)
				if !collector.tryClaim(fp) {
					tracker.finish(true)
					continue
				}
				out := evaluateShift(seed, state, opts, 1, "", [4]int{}, [4]int{}, board.PlayerLeftover{}, endLeft)
				collector.add(out)
				tracker.finish(false)
			}
		}()
	}
	for _, seed := range seeds {
		jobs <- seed
	}
	close(jobs)
	wg.Wait()
	if err := tracker.error(); err != nil {
		return nil, err
	}
	return collector.snapshot(), nil
}

type branchJob struct {
	parent parentBoard
	seed int64
}

func scanShiftBranch(k int, from, to int64, parents []parentBoard, opts Options, tracker *scanTracker) ([]Outcome, error) {
	jobs := make([]branchJob, 0, len(parents)*int(to-from+1))
	for _, p := range parents {
		for seed := from; seed <= to; seed++ {
			jobs = append(jobs, branchJob{parent: p, seed: seed})
		}
	}
	collector := newOutcomeCollector()
	workers := scanWorkers()
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
				rng := rand.New(rand.NewSource(job.seed))
				budgetP1 := opts.Finance.ReactorBudget + job.parent.leftover.Player1
				budgetP2 := opts.Finance.GridBudget + job.parent.leftover.Player2
				endLeft, err := board.SpendShiftBudget(rng, base, budgetP1, budgetP2, opts.MonthFilter)
				if err != nil {
					tracker.setErr(fmt.Errorf("schicht %d: %w", k, err))
					return
				}
				key := board.Fingerprint(base) + carryKey(job.parent.demand, job.parent.damage, job.parent.leftover)
				if !collector.tryClaim(key) {
					tracker.finish(true)
					continue
				}
				out := evaluateShift(job.seed, base, opts, k, job.parent.prevFP, job.parent.demand, job.parent.damage, job.parent.leftover, endLeft)
				collector.add(out)
				tracker.finish(false)
			}
		}()
	}
	for _, job := range jobs {
		work <- job
	}
	close(work)
	wg.Wait()
	if err := tracker.error(); err != nil {
		return nil, err
	}
	return collector.snapshot(), nil
}
