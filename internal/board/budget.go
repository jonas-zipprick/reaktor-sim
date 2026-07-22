package board

import (
	"fmt"
	"math"
	"math/rand"
	"sort"

	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/rules"
)

// PlayerLeftover holds unspent shift budget per player half after board purchases.
type PlayerLeftover struct {
	Player1 int
	Player2 int
}

// ShiftSpendResult holds leftover money and repair spending for one shift.
type ShiftSpendResult struct {
	Leftover          PlayerLeftover
	PreRepairSpentP1  int
	PreRepairSpentP2  int
	PostRepairSpentP1 int
	PostRepairSpentP2 int
}

// TotalRepairP1 returns total reactor repair spending (pre + post purchase).
func (r ShiftSpendResult) TotalRepairP1() int { return r.PreRepairSpentP1 + r.PostRepairSpentP1 }

// TotalRepairP2 returns total grid repair spending (pre + post purchase).
func (r ShiftSpendResult) TotalRepairP2() int { return r.PreRepairSpentP2 + r.PostRepairSpentP2 }

// MinFirstShiftFieldSpend is the minimum Geld each player spends on field
// purchases in shift 1 (strategy heuristic).
const MinFirstShiftFieldSpend = 2

// minBudgetSpendFraction is the lower bound for the field-spend target as a
// fraction of available budget (biases spending upward).
const minBudgetSpendFraction = 0.6

// RandomWithPlayerCosts builds a board spending a random achievable amount up to
// each player's budget. A budget of 0 leaves that half empty.
// minFieldSpend, when positive and affordable, forces at least that much field spend.
func RandomWithPlayerCosts(rng *rand.Rand, budgetP1, budgetP2, monthFilter, minFieldSpend int, month rules.Month) (*State, PlayerLeftover, error) {
	if budgetP1 < 0 || budgetP2 < 0 {
		return nil, PlayerLeftover{}, fmt.Errorf("Brett-Budget darf nicht negativ sein (Spieler 1: %d, Spieler 2: %d)", budgetP1, budgetP2)
	}
	if budgetP1 == 0 && budgetP2 == 0 {
		return NewEmpty(), PlayerLeftover{}, nil
	}

	s := NewEmpty()
	ApplyRandomShiftRotations(rng, s)
	left1, err := spendUpToOnSlots(rng, s, slotsForPlayer(true), budgetP1, "Spieler 1 (Reaktor)", monthFilter, minFieldSpend, month)
	if err != nil {
		return nil, PlayerLeftover{}, err
	}
	left2, err := spendUpToOnSlots(rng, s, slotsForPlayer(false), budgetP2, "Spieler 2 (Stromnetz)", monthFilter, minFieldSpend, month)
	if err != nil {
		return nil, PlayerLeftover{}, err
	}
	return s, PlayerLeftover{Player1: left1, Player2: left2}, nil
}

// SpendShiftBudget spends up to the given per-player budgets on an existing board.
// It handles repair (pre- and post-purchase) and field purchases.
// Repair is attempted with probability-based decisions before and after buying fields.
func SpendShiftBudget(rng *rand.Rand, s *State, budgetP1, budgetP2, monthFilter, minFieldSpend int, month rules.Month) (ShiftSpendResult, error) {
	if budgetP1 < 0 || budgetP2 < 0 {
		return ShiftSpendResult{}, fmt.Errorf("Schicht-Budget darf nicht negativ sein (Spieler 1: %d, Spieler 2: %d)", budgetP1, budgetP2)
	}
	repairsAllowed := month.RepairsAllowed()
	ApplyRandomShiftRotations(rng, s)

	var res ShiftSpendResult

	if repairsAllowed {
		res.PreRepairSpentP1 = s.MaybeRepair(rng, true, budgetP1, true)
		budgetP1 -= res.PreRepairSpentP1
		res.PreRepairSpentP2 = s.MaybeRepair(rng, false, budgetP2, true)
		budgetP2 -= res.PreRepairSpentP2
	}

	left1, err := spendHalfBudget(rng, s, true, budgetP1, monthFilter, minFieldSpend, month)
	if err != nil {
		return ShiftSpendResult{}, err
	}
	left2, err := spendHalfBudget(rng, s, false, budgetP2, monthFilter, minFieldSpend, month)
	if err != nil {
		return ShiftSpendResult{}, err
	}

	if repairsAllowed {
		res.PostRepairSpentP1 = s.MaybeRepair(rng, true, left1, false)
		left1 -= res.PostRepairSpentP1
		res.PostRepairSpentP2 = s.MaybeRepair(rng, false, left2, false)
		left2 -= res.PostRepairSpentP2
	}

	res.Leftover = PlayerLeftover{Player1: left1, Player2: left2}
	return res, nil
}

func effectiveMinFieldSpend(minFieldSpend, budget int) int {
	if minFieldSpend <= 0 || budget <= 0 || budget < minFieldSpend {
		return 0
	}
	return minFieldSpend
}

func spendUpToOnSlots(rng *rand.Rand, s *State, slots []hex.Coord, maxBudget int, label string, monthFilter, minFieldSpend int, month rules.Month) (int, error) {
	if maxBudget == 0 {
		return 0, nil
	}
	planner, err := newCostPlanner(slots, monthFilter, month)
	if err != nil {
		return 0, err
	}
	targets := achievableUpTo(planner, maxBudget)
	min := effectiveMinFieldSpend(minFieldSpend, maxBudget)
	if floor := budgetSpendFloor(maxBudget); floor > min {
		min = floor
	}
	if min > 0 {
		targets = filterAtLeast(targets, min)
	}
	if len(targets) == 0 {
		return maxBudget, nil
	}
	target := targets[rng.Intn(len(targets))]
	if target == 0 {
		return maxBudget, nil
	}
	if err := fillSlotCosts(rng, s, slots, target, label, monthFilter, month); err != nil {
		return 0, err
	}
	return maxBudget - target, nil
}

func achievableUpTo(p *costPlanner, max int) []int {
	var sums []int
	for sum, ok := range p.achievable {
		if ok && sum <= max {
			sums = append(sums, sum)
		}
	}
	sort.Ints(sums)
	return sums
}

func filterAtLeast(values []int, min int) []int {
	out := make([]int, 0, len(values))
	for _, v := range values {
		if v >= min {
			out = append(out, v)
		}
	}
	return out
}

func spendHalfBudget(rng *rand.Rand, s *State, player1 bool, budget, monthFilter, minFieldSpend int, month rules.Month) (int, error) {
	if budget <= 0 {
		return 0, nil
	}
	minFields := effectiveMinFieldSpend(minFieldSpend, budget)
	floor := budgetSpendFloor(budget)
	if floor > minFields {
		minFields = floor
	}
	targetSpend := pickFieldSpendTarget(rng, budget, s, player1, month, minFields)
	spent := 0
	fieldSpent := 0
	slots := slotsForPlayer(player1)
	market := field.FilterMarket(marketForPlayer(player1), monthFilter)
	for spent < targetSpend || fieldSpent < minFields {
		if spent >= budget && fieldSpent >= minFields {
			break
		}
		if spent >= targetSpend && fieldSpent < minFields {
			need := minFields - fieldSpent
			if spent+need > budget {
				break
			}
			targetSpend = spent + need
		}
		actions := affordableShiftActions(s, slots, market, budget, spent, targetSpend, month)
		if fieldSpent < minFields {
			actions = filterFieldPurchaseActions(actions)
		}
		if len(actions) == 0 {
			break
		}
		act := pickShiftAction(rng, actions, s)
		applyShiftAction(s, act, rng, month)
		spent += act.cost
		budget -= act.cost
		if act.kind != "remove" {
			fieldSpent += act.cost
		}
	}
	// If discrete field costs left us under the spend floor, keep buying
	// until the floor is reached or no affordable purchase remains.
	for fieldSpent < floor && budget > 0 {
		actions := filterFieldPurchaseActions(validShiftActions(s, slots, market, budget, month))
		if len(actions) == 0 {
			break
		}
		act := pickShiftAction(rng, actions, s)
		applyShiftAction(s, act, rng, month)
		spent += act.cost
		budget -= act.cost
		fieldSpent += act.cost
	}
	return budget, nil
}

func affordableShiftActions(s *State, slots []hex.Coord, market []field.Type, budget, spent, target int, month rules.Month) []shiftAction {
	all := validShiftActions(s, slots, market, budget, month)
	if len(all) == 0 {
		return nil
	}
	out := make([]shiftAction, 0, len(all))
	for _, act := range all {
		if spent+act.cost <= target {
			out = append(out, act)
		}
	}
	return out
}

func filterFieldPurchaseActions(actions []shiftAction) []shiftAction {
	out := make([]shiftAction, 0, len(actions))
	for _, act := range actions {
		if act.kind != "remove" {
			out = append(out, act)
		}
	}
	return out
}

func budgetSpendFloor(budget int) int {
	if budget <= 0 {
		return 0
	}
	return int(math.Ceil(float64(budget) * minBudgetSpendFraction))
}

func pickFieldSpendTarget(rng *rand.Rand, budget int, _ *State, _ bool, _ rules.Month, minFieldSpend int) int {
	if budget <= 0 {
		return 0
	}
	maxSpend := budget
	if floor := budgetSpendFloor(budget); floor > minFieldSpend {
		minFieldSpend = floor
	}
	if minFieldSpend > maxSpend {
		minFieldSpend = maxSpend
	}
	if maxSpend < minFieldSpend {
		maxSpend = minFieldSpend
	}
	return minFieldSpend + rng.Intn(maxSpend-minFieldSpend+1)
}
