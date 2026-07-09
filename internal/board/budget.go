package board

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
)

// PlayerLeftover holds unspent shift budget per player half after board purchases.
type PlayerLeftover struct {
	Player1 int
	Player2 int
}

// RandomWithPlayerCosts builds a board spending a random achievable amount up to
// each player's budget. A budget of 0 leaves that half empty.
func RandomWithPlayerCosts(rng *rand.Rand, budgetP1, budgetP2, monthFilter int) (*State, PlayerLeftover, error) {
	if budgetP1 < 0 || budgetP2 < 0 {
		return nil, PlayerLeftover{}, fmt.Errorf("Brett-Budget darf nicht negativ sein (Spieler 1: %d, Spieler 2: %d)", budgetP1, budgetP2)
	}
	if budgetP1 == 0 && budgetP2 == 0 {
		return NewEmpty(), PlayerLeftover{}, nil
	}

	s := NewEmpty()
	left1, err := spendUpToOnSlots(rng, s, slotsForPlayer(true), budgetP1, "Spieler 1 (Reaktor)", monthFilter)
	if err != nil {
		return nil, PlayerLeftover{}, err
	}
	left2, err := spendUpToOnSlots(rng, s, slotsForPlayer(false), budgetP2, "Spieler 2 (Stromnetz)", monthFilter)
	if err != nil {
		return nil, PlayerLeftover{}, err
	}
	return s, PlayerLeftover{Player1: left1, Player2: left2}, nil
}

// SpendShiftBudget spends up to the given per-player budgets on an existing board.
// It returns the unspent remainder per half.
func SpendShiftBudget(rng *rand.Rand, s *State, budgetP1, budgetP2, monthFilter int) (PlayerLeftover, error) {
	if budgetP1 < 0 || budgetP2 < 0 {
		return PlayerLeftover{}, fmt.Errorf("Schicht-Budget darf nicht negativ sein (Spieler 1: %d, Spieler 2: %d)", budgetP1, budgetP2)
	}
	left1, err := spendHalfBudget(rng, s, true, budgetP1, monthFilter)
	if err != nil {
		return PlayerLeftover{}, err
	}
	left2, err := spendHalfBudget(rng, s, false, budgetP2, monthFilter)
	if err != nil {
		return PlayerLeftover{}, err
	}
	return PlayerLeftover{Player1: left1, Player2: left2}, nil
}

func spendUpToOnSlots(rng *rand.Rand, s *State, slots []hex.Coord, maxBudget int, label string, monthFilter int) (int, error) {
	if maxBudget == 0 {
		return 0, nil
	}
	planner, err := newCostPlanner(slots, monthFilter)
	if err != nil {
		return 0, err
	}
	targets := achievableUpTo(planner, maxBudget)
	if len(targets) == 0 {
		return maxBudget, nil
	}
	target := targets[rng.Intn(len(targets))]
	if target == 0 {
		return maxBudget, nil
	}
	if err := fillSlotCosts(rng, s, slots, target, label, monthFilter); err != nil {
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

func spendHalfBudget(rng *rand.Rand, s *State, player1 bool, budget, monthFilter int) (int, error) {
	if budget <= 0 {
		return 0, nil
	}
	targetSpend := rng.Intn(budget + 1)
	spent := 0
	slots := slotsForPlayer(player1)
	market := field.FilterMarket(marketForPlayer(player1), monthFilter)
	for spent < targetSpend {
		actions := affordableShiftActions(s, slots, market, budget, spent, targetSpend)
		if len(actions) == 0 {
			break
		}
		act := actions[rng.Intn(len(actions))]
		applyShiftAction(s, act, rng)
		spent += act.cost
		budget -= act.cost
	}
	return budget, nil
}

func affordableShiftActions(s *State, slots []hex.Coord, market []field.Type, budget, spent, target int) []shiftAction {
	all := validShiftActions(s, slots, market, budget)
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
