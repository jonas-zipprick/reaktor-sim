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
	"github.com/jonas/reaktor-sim/internal/rules"
	"github.com/jonas/reaktor-sim/internal/sim"
)

// Options configure a full-month (multi-shift) Monte-Carlo scan.
type Options struct {
	Runs                  int
	EnergyCard            energy.Card  // provides the per-shift demand plan (Schichtplan)
	Finance               finance.Card // provides the per-shift budget
	Shifts                int          // number of consecutive shifts to simulate (1-5)
	ShiftKeep             int          // boards kept per ranking table to branch into the next shift
	MonthFilter           int          // campaign month for field availability (0 = all fields)
	Workers               int          // parallel scan workers (0 = GOMAXPROCS)
	StartBoardFingerprint string       // if set, shift 1 loads this board and spends budget per seed
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

func (o Options) workerCount() int {
	if o.Workers > 0 {
		return o.Workers
	}
	return scanWorkers()
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
	StartDamage          [4]int             // zone damage carry at shift start
	StartEmitterDamage   int                // igniter damage carry at shift start
	StartLeftover        board.PlayerLeftover // money carry at shift start
	EndLeftover          board.PlayerLeftover // unspent money after board purchases
	CarryBoardFingerprint string              // board after shift simulation carry (burnout cleanup)
	ReactorRepairBudget  int                // max igniter repair spend per MC run
	RepairBudget         int                // max grid repair spend per MC run
	BoardCosts           board.PlayerCosts
	Wins                 int
	AllDemandsNoDamage   int // all demands met, zero damage
	AllDemandsMax1Damage int // all demands met, at most one damage
	Max1DemandNoDamage   int // at most one unmet demand, zero damage
	Max1DemandMax1Damage int // at most one unmet demand, at most one damage
	Loops                int
	CriticalP1           int
	CriticalP2           int
	AvgEndDemand         ZoneTotals
	AvgEndDamage         ZoneTotals
	MedianEndDemand      [4]int // per-zone median remaining demand (carried to next shift)
	MedianEndDamage      [4]int // per-zone median remaining damage (carried to next shift)
	MedianEndEmitterDamage int    // median remaining igniter damage (carried to next shift)
	AvgSavedP1           float64 // avg unspent reactor money after repair per run
	AvgSavedP2           float64 // avg unspent grid money after repair per run
	AvgReactorRepairSpent float64 // avg money spent on igniter repair per run
	AvgRepairSpent       float64 // avg money spent on grid damage repair per run
	AvgSteps             float64
	Runs                 int
	TraceLoopRun         int // 1-based MC run for loop trace export, 0 = none
	TraceWinRun          int // 1-based MC run for win trace export, 0 = none
}

// WinRate returns the fraction of runs with all demands fulfilled.
func (o Outcome) WinRate() float64 {
	if o.Runs == 0 {
		return 0
	}
	return float64(o.Wins) / float64(o.Runs)
}

// AllDemandsMax1DamageRate returns the fraction of runs with no remaining demand
// and at most one damage chip.
func (o Outcome) AllDemandsMax1DamageRate() float64 {
	if o.Runs == 0 {
		return 0
	}
	return float64(o.AllDemandsMax1Damage) / float64(o.Runs)
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
	state, endLeft, shift1Prev, err := prepareShift1Board(rng, opts)
	if err != nil {
		return nil, err
	}
	return evaluateChain(seed, rng, state, opts, board.PlayerLeftover{}, endLeft, shift1Prev), nil
}

// prepareShift1Board builds or loads the shift-1 board and applies this shift's purchases.
func prepareShift1Board(rng *rand.Rand, opts Options) (*board.State, board.PlayerLeftover, string, error) {
	month := rules.Month{EnergyID: opts.EnergyCard.ID, FinanceID: opts.Finance.ID}
	if opts.StartBoardFingerprint != "" {
		state, err := board.FromFingerprint(opts.StartBoardFingerprint)
		if err != nil {
			return nil, board.PlayerLeftover{}, "", err
		}
		left, err := spendShift1Budget(rng, state, opts, month)
		return state, left, opts.StartBoardFingerprint, err
	}
	state, left, err := buildInitialBoard(rng, opts)
	return state, left, "", err
}

func spendShift1Budget(rng *rand.Rand, state *board.State, opts Options, month rules.Month) (board.PlayerLeftover, error) {
	p1 := opts.Finance.ReactorBudget
	p2 := opts.Finance.GridBudget
	return board.SpendShiftBudget(rng, state, p1, p2, opts.MonthFilter, month)
}

// buildInitialBoard creates the shift-1 board using the finance budget.
func buildInitialBoard(rng *rand.Rand, opts Options) (*board.State, board.PlayerLeftover, error) {
	p1 := opts.Finance.ReactorBudget
	p2 := opts.Finance.GridBudget
	if p1 <= 0 && p2 <= 0 {
		return board.Random(rng, opts.MonthFilter), board.PlayerLeftover{}, nil
	}
	state, left, err := board.RandomWithPlayerCosts(rng, p1, p2, opts.MonthFilter, rules.Month{
		EnergyID:  opts.EnergyCard.ID,
		FinanceID: opts.Finance.ID,
	})
	return state, left, err
}

func evaluateChain(seed int64, rng *rand.Rand, state *board.State, opts Options, carryLeft, firstEndLeft board.PlayerLeftover, shift1PrevFP string) []Outcome {
	shifts := opts.shiftCount()
	outcomes := make([]Outcome, 0, shifts)
	var carryDemand, carryDamage [4]int
	var carryEmitterDamage int
	prevFP := shift1PrevFP
	for k := 1; k <= shifts; k++ {
		startLeft := carryLeft
		var endLeft board.PlayerLeftover
		if k == 1 {
			endLeft = firstEndLeft
		} else {
			state.Damage = carryDamage
			state.EmitterDamage = carryEmitterDamage
			budgetP1 := opts.Finance.ReactorBudget + carryLeft.Player1
			budgetP2 := opts.Finance.GridBudget + carryLeft.Player2
			var err error
			endLeft, err = board.SpendShiftBudget(rng, state, budgetP1, budgetP2, opts.MonthFilter, rules.Month{
				EnergyID:  opts.EnergyCard.ID,
				FinanceID: opts.Finance.ID,
			})
			if err != nil {
				return outcomes
			}
		}
		out := evaluateShift(seed, state, opts, k, prevFP, carryDemand, carryDamage, carryEmitterDamage, startLeft, endLeft)
		outcomes = append(outcomes, out)
		prevFP = out.BoardFingerprint
		carryDemand = out.MedianEndDemand
		carryDamage = out.MedianEndDamage
		carryEmitterDamage = out.MedianEndEmitterDamage
		carryLeft = out.EndLeftover
	}
	return outcomes
}

// evaluateShift simulates one shift on state (already reflecting this shift's
// purchases) with the given carried demand/damage, returning the outcome.
func evaluateShift(seed int64, state *board.State, opts Options, shift int, prevFP string, carryDemand, carryDamage [4]int, carryEmitterDamage int, startLeft, endLeft board.PlayerLeftover) Outcome {
	cardDemand := opts.EnergyCard.ShiftDemands(shift)
	combined := board.ShiftDemands{
		Industry:    cardDemand.Industry + carryDemand[board.ZoneIndustry],
		Residential: cardDemand.Residential + carryDemand[board.ZoneResidential],
		Rail:        cardDemand.Rail + carryDemand[board.ZoneRail],
		Plant:       cardDemand.Plant + carryDemand[board.ZonePlant],
	}
	state.Damage = carryDamage
	state.EmitterDamage = carryEmitterDamage

	cfg := sim.DefaultConfig()
	cfg.EnergyCard = opts.EnergyCard
	cfg.FinanceCard = opts.Finance
	cfg.CriticalLimit = opts.Finance.CriticalLimit()
	cfg.Shift = shift
	cfg.RandomShift = false
	cfg.ShiftDemands = combined
	repairsAllowed := opts.Finance.RepairsAllowed()
	cfg.ReactorRepairBudget = reactorRepairBudget(endLeft.Player1, state, repairsAllowed)
	cfg.RepairBudget = gridRepairBudget(endLeft.Player2, state, repairsAllowed)

	out := evaluatePrepared(seed, state, opts.Runs, cfg, endLeft)
	out.Shift = shift
	out.PrevBoardFingerprint = prevFP
	out.StartDemands = combined
	out.StartDamage = carryDamage
	out.StartEmitterDamage = carryEmitterDamage
	out.StartLeftover = startLeft
	out.EndLeftover = endLeft
	out.ReactorRepairBudget = cfg.ReactorRepairBudget
	out.RepairBudget = cfg.RepairBudget
	return out
}

func evaluatePrepared(seed int64, state *board.State, runs int, cfg sim.Config, endLeft board.PlayerLeftover) Outcome {
	results := sim.RunMonteCarlo(state, runs, seed, cfg)
	out := aggregateOutcome(seed, state, runs, cfg.ReactorRepairBudget, cfg.RepairBudget, endLeft, results)
	if runs > 0 {
		if loops := sim.LoopTraceRunIndices(results, 1); len(loops) > 0 {
			out.TraceLoopRun = loops[0]
		}
		if wins := sim.WinTraceRunIndices(results, 1); len(wins) > 0 {
			out.TraceWinRun = wins[0]
		}
		idx := sim.MedianRunIndex(results)
		sim.ApplyShiftCarry(state, seed, idx+1, cfg)
		out.CarryBoardFingerprint = board.Fingerprint(state)
	}
	return out
}

func aggregateOutcome(seed int64, state *board.State, runs, reactorRepairBudget, gridRepairBudget int, endLeft board.PlayerLeftover, results []sim.Result) Outcome {
	out := Outcome{
		Seed:                seed,
		BoardFingerprint:    board.Fingerprint(state),
		BoardCosts:          state.PlayerCosts(),
		Runs:                runs,
		ReactorRepairBudget: reactorRepairBudget,
		RepairBudget:        gridRepairBudget,
	}
	var sumDemand, sumDamage ZoneTotals
	var sumSteps, sumGridRepair, sumReactorRepair, sumSavedP2 float64
	endDemand := make([][]int, 4)
	endDamage := make([][]int, 4)
	for z := range endDemand {
		endDemand[z] = make([]int, runs)
		endDamage[z] = make([]int, runs)
	}
	endEmitter := make([]int, runs)
	for i, res := range results {
		if res.AllDemandsMet {
			out.Wins++
		}
		remainingDemand := totalRemainingDemand(res)
		remainingDamage := totalRemainingDamage(res)
		if remainingDemand == 0 && remainingDamage == 0 {
			out.AllDemandsNoDamage++
		}
		if remainingDemand == 0 && remainingDamage <= 1 {
			out.AllDemandsMax1Damage++
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
		sumGridRepair += float64(res.RepairSpent)
		sumReactorRepair += float64(res.ReactorRepairSpent)
		savedP2 := endLeft.Player2 - res.RepairSpent
		if savedP2 < 0 {
			savedP2 = 0
		}
		sumSavedP2 += float64(savedP2)
		for z := board.ZoneIndustry; z <= board.ZonePlant; z++ {
			sumDemand[z] += float64(res.EndDemands[z])
			sumDamage[z] += float64(res.EndDamage[z])
			endDemand[z][i] = res.EndDemands[z]
			endDamage[z][i] = res.EndDamage[z]
		}
		endEmitter[i] = res.EndEmitterDamage
		sumSteps += float64(res.Steps)
	}
	if runs > 0 {
		for z := board.ZoneIndustry; z <= board.ZonePlant; z++ {
			out.AvgEndDemand[z] = sumDemand[z] / float64(runs)
			out.AvgEndDamage[z] = sumDamage[z] / float64(runs)
			out.MedianEndDemand[z] = medianInt(endDemand[z])
			out.MedianEndDamage[z] = medianInt(endDamage[z])
		}
		out.MedianEndEmitterDamage = medianInt(endEmitter)
		out.AvgSteps = sumSteps / float64(runs)
		out.AvgReactorRepairSpent = sumReactorRepair / float64(runs)
		out.AvgRepairSpent = sumGridRepair / float64(runs)
		savedP1 := float64(endLeft.Player1) - sumReactorRepair/float64(runs)
		if savedP1 < 0 {
			savedP1 = 0
		}
		out.AvgSavedP1 = savedP1
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
	carryFP       string
	prevFP        string // purchased board fingerprint (for PrevBoardFingerprint linkage)
	demand        [4]int
	damage        [4]int
	emitterDamage int
	leftover      board.PlayerLeftover
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
		if k >= 2 {
			result.Shifts[k-2].Outcomes = pruneShiftOutcomes(result.Shifts[k-2].Outcomes, outcomes)
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
	candidateN := keep + len(exclude)
	if candidateN < keep*4 {
		candidateN = keep * 4
	}
	if candidateN > len(outcomes) {
		candidateN = len(outcomes)
	}
	for _, pick := range keepTables {
		ranked := pick(outcomes, candidateN)
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
				carryFP:       o.CarryBoardFingerprint,
				prevFP:        o.BoardFingerprint,
				demand:        o.MedianEndDemand,
				damage:        o.MedianEndDamage,
				emitterDamage: o.MedianEndEmitterDamage,
				leftover:      o.EndLeftover,
			})
			picked++
		}
	}
	return parents
}

func parentBoardKey(o Outcome) string {
	return o.BoardFingerprint + carryKey(o.MedianEndDemand, o.MedianEndDamage, o.MedianEndEmitterDamage, o.EndLeftover)
}

func loopTableKeys(outcomes []Outcome, keep int) map[string]struct{} {
	exclude := make(map[string]struct{})
	for _, o := range TopLoops(outcomes, keep) {
		exclude[parentBoardKey(o)] = struct{}{}
	}
	return exclude
}

func carryKey(demand, damage [4]int, emitterDamage int, leftover board.PlayerLeftover) string {
	return fmt.Sprintf("|d%v|s%v|z%d|g%v", demand, damage, emitterDamage, leftover)
}

func gridRepairBudget(leftoverP2 int, state *board.State, repairsAllowed bool) int {
	if !repairsAllowed {
		return 0
	}
	return board.GridRepairBudget(leftoverP2, state)
}

func reactorRepairBudget(leftoverP1 int, state *board.State, repairsAllowed bool) int {
	if !repairsAllowed {
		return 0
	}
	return board.ReactorRepairBudget(leftoverP1, state)
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
	return total + res.EndEmitterDamage
}

func pruneShiftOutcomes(outcomes []Outcome, childOutcomes []Outcome) []Outcome {
	if len(outcomes) == 0 || len(childOutcomes) == 0 {
		return outcomes
	}
	needed := make(map[string]struct{})
	for _, child := range childOutcomes {
		if child.PrevBoardFingerprint != "" {
			needed[child.PrevBoardFingerprint] = struct{}{}
		}
	}
	if len(needed) == 0 {
		return outcomes
	}
	pruned := make([]Outcome, 0, len(needed))
	for _, o := range outcomes {
		if _, ok := needed[o.BoardFingerprint]; ok {
			pruned = append(pruned, o)
		}
	}
	return pruned
}

