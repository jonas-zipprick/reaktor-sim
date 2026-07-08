// Package seedsearch evaluates Monte-Carlo outcomes across many board seeds.
package seedsearch

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/energy"
	"github.com/jonas/reaktor-sim/internal/sim"
)

// Options mirror the main simulator settings used per seed.
type Options struct {
	Runs              int
	InitialHeat       int
	InitialNeutron    int
	MixedEmitterTrigger bool
	EnergyCardID      string
	Shift             int
	CostP1            int
	CostP2            int
}

// ZoneTotals holds per-zone averages across Monte-Carlo runs (index = board.Zone).
type ZoneTotals [4]float64

// Outcome aggregates one seed's Monte-Carlo batch.
type Outcome struct {
	Seed         int64
	BoardCosts   board.PlayerCosts
	Wins         int
	Loops        int
	CriticalP1   int
	CriticalP2   int
	AvgEndDemand ZoneTotals
	AvgEndDamage ZoneTotals
	Runs         int
}

// WinRate returns the fraction of runs with all demands fulfilled.
func (o Outcome) WinRate() float64 {
	if o.Runs == 0 {
		return 0
	}
	return float64(o.Wins) / float64(o.Runs)
}

// LoopRate returns the fraction of runs that hit the step limit.
func (o Outcome) LoopRate() float64 {
	if o.Runs == 0 {
		return 0
	}
	return float64(o.Loops) / float64(o.Runs)
}

// EndDemandSummary formats average remaining demand as I2 W1 b0 R2.
func (o Outcome) EndDemandSummary() string {
	return formatZoneTotals(o.AvgEndDemand)
}

// EndDamageSummary formats average remaining damage; returns "-" if all zero.
func (o Outcome) EndDamageSummary() string {
	if o.AvgEndDamage.allZero() {
		return "-"
	}
	return formatZoneTotals(o.AvgEndDamage)
}

func (z ZoneTotals) allZero() bool {
	for _, v := range z {
		if v > 0.0001 {
			return false
		}
	}
	return true
}

func formatZoneTotals(totals ZoneTotals) string {
	zones := []board.Zone{
		board.ZoneIndustry,
		board.ZoneResidential,
		board.ZoneRail,
		board.ZonePlant,
	}
	parts := make([]string, 0, len(zones))
	for _, z := range zones {
		parts = append(parts, board.ZoneLetter(z)+formatAvg(totals[z]))
	}
	return strings.Join(parts, " ")
}

func formatAvg(v float64) string {
	if v < 0.05 {
		return "0"
	}
	rounded := float64(int(v + 0.5))
	if v >= rounded-0.05 && v <= rounded+0.05 {
		return fmt.Sprintf("%d", int(rounded))
	}
	return fmt.Sprintf("%.1f", v)
}

// Evaluate runs the full Monte-Carlo batch for one board seed.
func Evaluate(seed int64, opts Options) (Outcome, error) {
	state, cfg, err := prepare(seed, opts)
	if err != nil {
		return Outcome{}, err
	}
	results := sim.RunMonteCarlo(state, opts.Runs, seed, cfg)
	out := Outcome{
		Seed:       seed,
		BoardCosts: state.PlayerCosts(),
		Runs:       opts.Runs,
	}
	var sumDemand, sumDamage ZoneTotals
	for _, res := range results {
		if res.AllDemandsMet {
			out.Wins++
		}
		if res.StepLimitExceeded {
			out.Loops++
		}
		if res.CriticalP1 {
			out.CriticalP1++
		}
		if res.CriticalP2 {
			out.CriticalP2++
		}
		for z := board.ZoneIndustry; z <= board.ZonePlant; z++ {
			sumDemand[z] += float64(res.EndDemands[z])
			sumDamage[z] += float64(res.EndDamage[z])
		}
	}
	if opts.Runs > 0 {
		for z := board.ZoneIndustry; z <= board.ZonePlant; z++ {
			out.AvgEndDemand[z] = sumDemand[z] / float64(opts.Runs)
			out.AvgEndDamage[z] = sumDamage[z] / float64(opts.Runs)
		}
	}
	return out, nil
}

// ProgressFunc is called after each seed is evaluated.
type ProgressFunc func(done, total int64)

// Scan evaluates every seed in [from, to] inclusive.
func Scan(from, to int64, opts Options, progress ProgressFunc) ([]Outcome, error) {
	if from > to {
		return nil, fmt.Errorf("from (%d) > to (%d)", from, to)
	}
	total := to - from + 1
	out := make([]Outcome, 0, total)
	for seed := from; seed <= to; seed++ {
		o, err := Evaluate(seed, opts)
		if err != nil {
			return nil, fmt.Errorf("seed %d: %w", seed, err)
		}
		out = append(out, o)
		if progress != nil {
			progress(seed-from+1, total)
		}
	}
	return out, nil
}

// TopWins returns up to n outcomes with the highest win counts.
func TopWins(outcomes []Outcome, n int) []Outcome {
	return topN(outcomes, n, func(a, b Outcome) bool {
		if a.Wins != b.Wins {
			return a.Wins > b.Wins
		}
		return a.Seed < b.Seed
	})
}

// TopLoops returns up to n outcomes with the highest loop counts.
func TopLoops(outcomes []Outcome, n int) []Outcome {
	return topN(outcomes, n, func(a, b Outcome) bool {
		if a.Loops != b.Loops {
			return a.Loops > b.Loops
		}
		return a.Seed < b.Seed
	})
}

func topN(outcomes []Outcome, n int, less func(a, b Outcome) bool) []Outcome {
	if n <= 0 || len(outcomes) == 0 {
		return nil
	}
	cp := append([]Outcome(nil), outcomes...)
	sort.Slice(cp, func(i, j int) bool { return less(cp[i], cp[j]) })
	if n > len(cp) {
		n = len(cp)
	}
	return cp[:n]
}

func prepare(seed int64, opts Options) (*board.State, sim.Config, error) {
	card, ok := energy.ByID(opts.EnergyCardID)
	if !ok {
		return nil, sim.Config{}, fmt.Errorf("unbekannte Energiekarte %q", opts.EnergyCardID)
	}
	if opts.Shift < 0 || opts.Shift > 5 {
		return nil, sim.Config{}, fmt.Errorf("shift muss 0-5 sein, got %d", opts.Shift)
	}

	rng := rand.New(rand.NewSource(seed))
	var state *board.State
	var err error
	if opts.CostP1 > 0 || opts.CostP2 > 0 {
		state, err = board.RandomWithPlayerCosts(rng, opts.CostP1, opts.CostP2)
	} else {
		state = board.Random(rng)
	}
	if err != nil {
		return nil, sim.Config{}, err
	}

	cfg := sim.DefaultConfig()
	cfg.InitialHeat = opts.InitialHeat
	cfg.InitialNeutron = opts.InitialNeutron
	cfg.MixedEmitterTrigger = opts.MixedEmitterTrigger
	cfg.EnergyCard = card
	cfg.Shift = opts.Shift
	cfg.RandomShift = opts.Shift == 0
	if cfg.RandomShift {
		cfg.ShiftDemands = card.ShiftDemands(1)
	} else {
		cfg.ShiftDemands = card.ShiftDemands(opts.Shift)
	}
	return state, cfg, nil
}
