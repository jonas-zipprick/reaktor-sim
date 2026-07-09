// Package seedsearch evaluates Monte-Carlo outcomes across many board seeds.
package seedsearch

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/energy"
	"github.com/jonas/reaktor-sim/internal/finance"
	"github.com/jonas/reaktor-sim/internal/sim"
)

// Options configure a full-month (multi-shift) Monte-Carlo scan.
type Options struct {
	Runs       int
	EnergyCard energy.Card  // provides the per-shift demand plan (Schichtplan)
	Finance    finance.Card // provides the per-shift budget
	Shifts     int          // number of consecutive shifts to simulate (1-5)
	ShiftKeep   int          // boards kept per ranking table to branch into the next shift
	MonthFilter int          // campaign month for field availability (0 = all fields)
}

// shiftCount clamps the configured shift count to the valid 1-5 range.
func (o Options) shiftCount() int {
	if o.Shifts < 1 {
		return 1
	}
	if o.Shifts > 5 {
		return 5
	}
	return o.Shifts
}

// shiftKeep clamps the number of boards kept per ranking table to at least 1.
func (o Options) shiftKeep() int {
	if o.ShiftKeep < 1 {
		return 1
	}
	return o.ShiftKeep
}

// ZoneTotals holds per-zone averages across Monte-Carlo runs (index = board.Zone).
type ZoneTotals [4]float64

// Outcome aggregates one seed's Monte-Carlo batch for a single shift.
type Outcome struct {
	Seed                 int64
	Shift                int
	BoardFingerprint     string
	PrevBoardFingerprint string             // board carried in from the previous shift ("" for shift 1)
	StartDemands         board.ShiftDemands // card + carry at shift start
	StartDamage          [4]int             // damage carry at shift start
	StartLeftover        board.PlayerLeftover // money carry at shift start
	EndLeftover          board.PlayerLeftover // unspent money after board purchases
	CarryBoardFingerprint string              // board after shift simulation carry (burnout cleanup)
	RepairBudget         int                // max repair spend per MC run
	BoardCosts           board.PlayerCosts
	Wins                 int
	AllDemandsNoDamage   int // all demands met, zero damage
	Max1DemandNoDamage   int // at most one unmet demand, zero damage
	Max1DemandMax1Damage int // at most one unmet demand, at most one damage
	Loops                int
	CriticalP1           int
	CriticalP2           int
	AvgEndDemand         ZoneTotals
	AvgEndDamage         ZoneTotals
	MedianEndDemand      [4]int // per-zone median remaining demand (carried to next shift)
	MedianEndDamage      [4]int // per-zone median remaining damage (carried to next shift)
	AvgSavedP1           float64 // avg unspent reactor money after repair per run
	AvgSavedP2           float64 // avg unspent grid money after repair per run
	AvgRepairSpent       float64 // avg money spent on damage repair per run
	AvgSteps             float64
	Runs                 int
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

// TotalMedianEndDemand returns the sum of per-zone median remaining demand across MC runs.
func (o Outcome) TotalMedianEndDemand() int {
	total := 0
	for z := board.ZoneIndustry; z <= board.ZonePlant; z++ {
		total += o.MedianEndDemand[z]
	}
	return total
}

// TotalMedianEndDamage returns the sum of per-zone median remaining damage across MC runs.
func (o Outcome) TotalMedianEndDamage() int {
	total := 0
	for z := board.ZoneIndustry; z <= board.ZonePlant; z++ {
		total += o.MedianEndDamage[z]
	}
	return total
}

// AvgStepsSummary formats the average number of simulation steps per run.
func (o Outcome) AvgStepsSummary() string {
	return formatAvg(o.AvgSteps)
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

// AvgSavedSummary formats average unspent money after repair as "P1/P2".
func (o Outcome) AvgSavedSummary() string {
	return fmt.Sprintf("%s/%s", formatAvg(o.AvgSavedP1), formatAvg(o.AvgSavedP2))
}

// AvgRepairSummary formats average repair spend per run.
func (o Outcome) AvgRepairSummary() string {
	return formatAvg(o.AvgRepairSpent)
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

// EvaluateChain simulates all configured shifts for one seed and returns one
// Outcome per shift. The board, remaining demand and damage carry forward.
func EvaluateChain(seed int64, opts Options) ([]Outcome, error) {
	rng := rand.New(rand.NewSource(seed))
	state, endLeft, err := buildInitialBoard(rng, opts)
	if err != nil {
		return nil, err
	}
	return evaluateChain(seed, rng, state, opts, board.PlayerLeftover{}, endLeft), nil
}

// buildInitialBoard creates the shift-1 board using the finance budget.
func buildInitialBoard(rng *rand.Rand, opts Options) (*board.State, board.PlayerLeftover, error) {
	p1 := opts.Finance.ReactorBudget
	p2 := opts.Finance.GridBudget
	if p1 <= 0 && p2 <= 0 {
		return board.Random(rng, opts.MonthFilter), board.PlayerLeftover{}, nil
	}
	state, left, err := board.RandomWithPlayerCosts(rng, p1, p2, opts.MonthFilter)
	return state, left, err
}

func evaluateChain(seed int64, rng *rand.Rand, state *board.State, opts Options, carryLeft, firstEndLeft board.PlayerLeftover) []Outcome {
	shifts := opts.shiftCount()
	outcomes := make([]Outcome, 0, shifts)
	var carryDemand, carryDamage [4]int
	var prevFP string
	for k := 1; k <= shifts; k++ {
		startLeft := carryLeft
		var endLeft board.PlayerLeftover
		if k == 1 {
			endLeft = firstEndLeft
		} else {
			budgetP1 := opts.Finance.ReactorBudget + carryLeft.Player1
			budgetP2 := opts.Finance.GridBudget + carryLeft.Player2
			var err error
			endLeft, err = board.SpendShiftBudget(rng, state, budgetP1, budgetP2, opts.MonthFilter)
			if err != nil {
				return outcomes
			}
		}
		out := evaluateShift(seed, state, opts, k, prevFP, carryDemand, carryDamage, startLeft, endLeft)
		outcomes = append(outcomes, out)
		prevFP = out.BoardFingerprint
		carryDemand = out.MedianEndDemand
		carryDamage = out.MedianEndDamage
		carryLeft = out.EndLeftover
	}
	return outcomes
}

// evaluateShift simulates one shift on state (already reflecting this shift's
// purchases) with the given carried demand/damage, returning the outcome.
func evaluateShift(seed int64, state *board.State, opts Options, shift int, prevFP string, carryDemand, carryDamage [4]int, startLeft, endLeft board.PlayerLeftover) Outcome {
	cardDemand := opts.EnergyCard.ShiftDemands(shift)
	combined := board.ShiftDemands{
		Industry:    cardDemand.Industry + carryDemand[board.ZoneIndustry],
		Residential: cardDemand.Residential + carryDemand[board.ZoneResidential],
		Rail:        cardDemand.Rail + carryDemand[board.ZoneRail],
		Plant:       cardDemand.Plant + carryDemand[board.ZonePlant],
	}
	state.Damage = carryDamage

	cfg := sim.DefaultConfig()
	cfg.EnergyCard = energy.Card{}
	cfg.Shift = shift
	cfg.RandomShift = false
	cfg.ShiftDemands = combined
	cfg.RepairBudget = gridRepairBudget(endLeft.Player2, state, opts.Finance.RepairsAllowed())

	out := evaluatePrepared(seed, state, opts.Runs, cfg, endLeft)
	out.Shift = shift
	out.PrevBoardFingerprint = prevFP
	out.StartDemands = combined
	out.StartDamage = carryDamage
	out.StartLeftover = startLeft
	out.EndLeftover = endLeft
	return out
}

func evaluatePrepared(seed int64, state *board.State, runs int, cfg sim.Config, endLeft board.PlayerLeftover) Outcome {
	results := sim.RunMonteCarlo(state, runs, seed, cfg)
	out := aggregateOutcome(seed, state, runs, cfg.RepairBudget, endLeft, results)
	if runs > 0 {
		idx := sim.MedianRunIndex(results)
		sim.ApplyShiftCarry(state, seed, idx+1, cfg)
		out.CarryBoardFingerprint = board.Fingerprint(state)
	}
	return out
}

func aggregateOutcome(seed int64, state *board.State, runs int, repairBudget int, endLeft board.PlayerLeftover, results []sim.Result) Outcome {
	out := Outcome{
		Seed:             seed,
		BoardFingerprint: board.Fingerprint(state),
		BoardCosts:       state.PlayerCosts(),
		Runs:             runs,
		RepairBudget:     repairBudget,
	}
	var sumDemand, sumDamage ZoneTotals
	var sumSteps, sumRepair, sumSavedP2 float64
	var endDemand, endDamage [4][]int
	for _, res := range results {
		if res.AllDemandsMet {
			out.Wins++
		}
		remainingDemand := totalRemainingDemand(res)
		remainingDamage := totalRemainingDamage(res)
		if remainingDemand == 0 && remainingDamage == 0 {
			out.AllDemandsNoDamage++
		}
		if remainingDemand <= 1 && remainingDamage == 0 {
			out.Max1DemandNoDamage++
		}
		if remainingDemand <= 1 && remainingDamage <= 1 {
			out.Max1DemandMax1Damage++
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
		sumRepair += float64(res.RepairSpent)
		savedP2 := endLeft.Player2 - res.RepairSpent
		if savedP2 < 0 {
			savedP2 = 0
		}
		sumSavedP2 += float64(savedP2)
		for z := board.ZoneIndustry; z <= board.ZonePlant; z++ {
			sumDemand[z] += float64(res.EndDemands[z])
			sumDamage[z] += float64(res.EndDamage[z])
			endDemand[z] = append(endDemand[z], res.EndDemands[z])
			endDamage[z] = append(endDamage[z], res.EndDamage[z])
		}
		sumSteps += float64(res.Steps)
	}
	if runs > 0 {
		for z := board.ZoneIndustry; z <= board.ZonePlant; z++ {
			out.AvgEndDemand[z] = sumDemand[z] / float64(runs)
			out.AvgEndDamage[z] = sumDamage[z] / float64(runs)
			out.MedianEndDemand[z] = medianInt(endDemand[z])
			out.MedianEndDamage[z] = medianInt(endDamage[z])
		}
		out.AvgSteps = sumSteps / float64(runs)
		out.AvgRepairSpent = sumRepair / float64(runs)
		out.AvgSavedP1 = float64(endLeft.Player1)
		out.AvgSavedP2 = sumSavedP2 / float64(runs)
	}
	return out
}

// medianInt returns the median of vals, rounding a two-element midpoint up.
func medianInt(vals []int) int {
	if len(vals) == 0 {
		return 0
	}
	s := append([]int(nil), vals...)
	sort.Ints(s)
	n := len(s)
	if n%2 == 1 {
		return s[n/2]
	}
	return int(math.Round(float64(s[n/2-1]+s[n/2]) / 2))
}

// ProgressFunc is called after each seed evaluation. shift is the active shift (1-based).
type ProgressFunc func(done, total int64, shift int)

// ShiftResult holds all seed outcomes for one shift.
type ShiftResult struct {
	Shift    int
	Outcomes []Outcome
}

// ScanResult holds per-shift outcomes from a seed range scan.
type ScanResult struct {
	Shifts            []ShiftResult
	SkippedDuplicates int64
}

// parentBoard is a board kept from one shift to branch into the next: its
// post-simulation carry fingerprint plus demand/damage/leftover carried forward.
type parentBoard struct {
	carryFP  string
	prevFP   string // purchased board fingerprint (for PrevBoardFingerprint linkage)
	demand   [4]int
	damage   [4]int
	leftover board.PlayerLeftover
}

// KeepTableCount is the number of ranking tables from which boards are kept to
// branch into the next shift (the success tables; loops are excluded).
const KeepTableCount = 4

var keepTables = []func([]Outcome, int) []Outcome{
	TopWins,
	TopAllDemandsNoDamage,
	TopMax1DemandNoDamage,
	TopMax1DemandMax1Damage,
}

// EstimateScanWork returns an upper bound on Scan evaluations for progress bars.
func EstimateScanWork(from, to int64, opts Options) int64 {
	seedCount := to - from + 1
	if seedCount < 1 {
		return 1
	}
	shifts := opts.shiftCount()
	if shifts <= 1 {
		return seedCount
	}
	maxParents := int64(opts.shiftKeep()) * int64(KeepTableCount)
	return seedCount + (int64(shifts)-1)*maxParents*seedCount
}

// Scan simulates the month shift by shift. Shift 1 builds a board for every seed
// (duplicates removed). For each following shift, the top boards of the previous
// shift (opts.ShiftKeep per ranking table) are re-developed with every seed's
// purchase, keeping the branching bounded.
func Scan(from, to int64, opts Options, progress ProgressFunc) (ScanResult, error) {
	if from > to {
		return ScanResult{}, fmt.Errorf("from (%d) > to (%d)", from, to)
	}
	shifts := opts.shiftCount()
	var result ScanResult
	tracker := newScanTracker(progress, EstimateScanWork(from, to, opts))
	tracker.setShift(1)

	shift1, err := scanShift1(from, to, opts, tracker)
	if err != nil {
		return ScanResult{}, err
	}
	result.Shifts = append(result.Shifts, ShiftResult{Shift: 1, Outcomes: shift1})
	parents := selectParents(shift1, opts.shiftKeep())

	for k := 2; k <= shifts; k++ {
		tracker.setShift(k)
		outcomes, err := scanShiftBranch(k, from, to, parents, opts, tracker)
		if err != nil {
			return ScanResult{}, err
		}
		result.Shifts = append(result.Shifts, ShiftResult{Shift: k, Outcomes: outcomes})
		parents = selectParents(outcomes, opts.shiftKeep())
	}
	result.SkippedDuplicates = tracker.skipped.Load()
	return result, nil
}

// selectParents gathers the top keep boards from each success ranking table,
// de-duplicated by board fingerprint and carried state. Entries that appear in
// the Loops table are skipped so the next-ranked success outcomes are used.
func selectParents(outcomes []Outcome, keep int) []parentBoard {
	if keep < 1 {
		keep = 1
	}
	exclude := loopTableKeys(outcomes, keep)
	seen := make(map[string]struct{})
	var parents []parentBoard
	for _, pick := range keepTables {
		ranked := pick(outcomes, len(outcomes))
		picked := 0
		for _, o := range ranked {
			if picked >= keep {
				break
			}
			key := parentBoardKey(o)
			if _, ok := seen[key]; ok {
				continue
			}
			if _, skip := exclude[key]; skip {
				continue
			}
			seen[key] = struct{}{}
			parents = append(parents, parentBoard{
				carryFP:  o.CarryBoardFingerprint,
				prevFP:   o.BoardFingerprint,
				demand:   o.MedianEndDemand,
				damage:   o.MedianEndDamage,
				leftover: o.EndLeftover,
			})
			picked++
		}
	}
	return parents
}

func parentBoardKey(o Outcome) string {
	return o.BoardFingerprint + carryKey(o.MedianEndDemand, o.MedianEndDamage, o.EndLeftover)
}

func loopTableKeys(outcomes []Outcome, keep int) map[string]struct{} {
	exclude := make(map[string]struct{})
	for _, o := range TopLoops(outcomes, keep) {
		exclude[parentBoardKey(o)] = struct{}{}
	}
	return exclude
}

func carryKey(demand, damage [4]int, leftover board.PlayerLeftover) string {
	return fmt.Sprintf("|d%v|s%v|g%v", demand, damage, leftover)
}

func gridRepairBudget(leftoverP2 int, state *board.State, repairsAllowed bool) int {
	if !repairsAllowed {
		return 0
	}
	return board.GridRepairBudget(leftoverP2, state)
}

// WinningOnly returns outcomes with at least one run where all demands were met.
func WinningOnly(outcomes []Outcome) []Outcome {
	out := make([]Outcome, 0, len(outcomes))
	for _, o := range outcomes {
		if o.Wins > 0 {
			out = append(out, o)
		}
	}
	return out
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

// TopLoops returns up to n outcomes with the highest loop counts. Outcomes
// without any loops are excluded, so the result is empty when no loops occurred.
func TopLoops(outcomes []Outcome, n int) []Outcome {
	withLoops := make([]Outcome, 0, len(outcomes))
	for _, o := range outcomes {
		if o.Loops > 0 {
			withLoops = append(withLoops, o)
		}
	}
	return topN(withLoops, n, func(a, b Outcome) bool {
		if a.Loops != b.Loops {
			return a.Loops > b.Loops
		}
		return a.Seed < b.Seed
	})
}

// TopAllDemandsNoDamage returns up to n outcomes with the most runs where all
// demands were met and no damage remained.
func TopAllDemandsNoDamage(outcomes []Outcome, n int) []Outcome {
	return topN(outcomes, n, func(a, b Outcome) bool {
		if a.AllDemandsNoDamage != b.AllDemandsNoDamage {
			return a.AllDemandsNoDamage > b.AllDemandsNoDamage
		}
		return a.Seed < b.Seed
	})
}

// TopMax1DemandNoDamage returns up to n outcomes with the most runs where at
// most one demand was unmet and no damage remained.
func TopMax1DemandNoDamage(outcomes []Outcome, n int) []Outcome {
	return topN(outcomes, n, func(a, b Outcome) bool {
		if a.Max1DemandNoDamage != b.Max1DemandNoDamage {
			return a.Max1DemandNoDamage > b.Max1DemandNoDamage
		}
		return a.Seed < b.Seed
	})
}

// TopMax1DemandMax1Damage returns up to n outcomes with the most runs where at
// most one demand was unmet and at most one damage chip remained.
func TopMax1DemandMax1Damage(outcomes []Outcome, n int) []Outcome {
	return topN(outcomes, n, func(a, b Outcome) bool {
		if a.Max1DemandMax1Damage != b.Max1DemandMax1Damage {
			return a.Max1DemandMax1Damage > b.Max1DemandMax1Damage
		}
		return a.Seed < b.Seed
	})
}

func totalRemainingDemand(res sim.Result) int {
	total := 0
	for z := board.ZoneIndustry; z <= board.ZonePlant; z++ {
		total += res.EndDemands[z]
	}
	return total
}

func totalRemainingDamage(res sim.Result) int {
	total := 0
	for z := board.ZoneIndustry; z <= board.ZonePlant; z++ {
		total += res.EndDamage[z]
	}
	return total
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

