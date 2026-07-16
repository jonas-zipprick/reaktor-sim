package board

import (
	"fmt"
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

// MinFirstShiftFieldSpend is the minimum Geld each player spends on field
// purchases in shift 1 (strategy heuristic).
const MinFirstShiftFieldSpend = 2

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

const (
	damageRepairThreshold   = 3
	damageRepairLikelihood  = 0.9
)

// SpendShiftBudget spends up to the given per-player budgets on an existing board.
// It returns the unspent remainder per half.
// minFieldSpend, when positive and affordable, forces at least that much field spend.
// When s carries more than damageRepairThreshold total damage chips and repairs
// are allowed, most draws reserve money for repairs on the affected half.
func SpendShiftBudget(rng *rand.Rand, s *State, budgetP1, budgetP2, monthFilter, minFieldSpend int, month rules.Month) (PlayerLeftover, error) {
	if budgetP1 < 0 || budgetP2 < 0 {
		return PlayerLeftover{}, fmt.Errorf("Schicht-Budget darf nicht negativ sein (Spieler 1: %d, Spieler 2: %d)", budgetP1, budgetP2)
	}
	ApplyRandomShiftRotations(rng, s)
	left1, err := spendHalfBudget(rng, s, true, budgetP1, monthFilter, minFieldSpend, month)
	if err != nil {
		return PlayerLeftover{}, err
	}
	left2, err := spendHalfBudget(rng, s, false, budgetP2, monthFilter, minFieldSpend, month)
	if err != nil {
		return PlayerLeftover{}, err
	}
	return PlayerLeftover{Player1: left1, Player2: left2}, nil
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
	if min := effectiveMinFieldSpend(minFieldSpend, maxBudget); min > 0 {
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

func pickFieldSpendTarget(rng *rand.Rand, budget int, s *State, player1 bool, month rules.Month, minFieldSpend int) int {
	if budget <= 0 {
		return 0
	}
	maxSpend := budget
	if month.RepairsAllowed() && s.TotalBoardDamage() > damageRepairThreshold {
		damage := s.EmitterDamage
		if !player1 {
			damage = s.TotalPlayer2Damage()
		}
		if damage > 0 && rng.Float64() < damageRepairLikelihood {
			reserve := damage
			if reserve > budget {
				reserve = budget
			}
			maxSpend = budget - reserve
		}
	}
	if minFieldSpend > 0 && minFieldSpend > maxSpend {
		maxSpend = budget
	}
	if minFieldSpend > maxSpend {
		minFieldSpend = maxSpend
	}
	if maxSpend < minFieldSpend {
		maxSpend = minFieldSpend
	}
	return minFieldSpend + rng.Intn(maxSpend-minFieldSpend+1)
}
