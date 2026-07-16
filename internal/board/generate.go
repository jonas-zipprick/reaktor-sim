package board

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/rules"
)

// ErrCostNotAchievable means no placement sums to the requested cost.
var ErrCostNotAchievable = errors.New("board cost not achievable")

// PlaceableSlots returns all hex cells where fields can be placed.
func PlaceableSlots() []hex.Coord {
	out := make([]hex.Coord, 0, len(hex.AllBoardCoords))
	for _, c := range hex.AllBoardCoords {
		if c.Kind() == hex.CellSlot {
			out = append(out, c)
		}
	}
	return out
}

// RandomWithCost builds a random board whose total placement cost equals target.
func RandomWithCost(rng *rand.Rand, target int) (*State, error) {
	if target < 0 {
		return nil, fmt.Errorf("Brettkosten duerfen nicht negativ sein (erhalten: %d)", target)
	}
	return randomWithCostOnSlots(rng, target)
}

// RandomWithPlayerCosts builds a board with exact costs per player half.
// Deprecated: use RandomWithPlayerCosts in budget.go which spends up to budget.
// Kept as exact-fill helper for tests.
func randomWithPlayerCostsExact(rng *rand.Rand, player1, player2, monthFilter int) (*State, error) {
	if player1 < 0 || player2 < 0 {
		return nil, fmt.Errorf("Brettkosten duerfen nicht negativ sein (Spieler 1: %d, Spieler 2: %d)", player1, player2)
	}
	if player1 == 0 && player2 == 0 {
		return NewEmpty(), nil
	}

	s := NewEmpty()
	if err := fillSlotCosts(rng, s, slotsForPlayer(true), player1, "Spieler 1 (Reaktor)", monthFilter, rules.Month{}); err != nil {
		return nil, err
	}
	if err := fillSlotCosts(rng, s, slotsForPlayer(false), player2, "Spieler 2 (Stromnetz)", monthFilter, rules.Month{}); err != nil {
		return nil, err
	}
	return s, nil
}

const removeFieldBudgetCost = 1

// burnedRefreshWeight is how much more likely a same-type rebuild on a burned-out
// slot is compared to other affordable shift actions.
const burnedRefreshWeight = 3

type shiftAction struct {
	kind  string
	coord hex.Coord
	tile  field.Type
	cost  int
}

func validShiftActions(s *State, slots []hex.Coord, market []field.Type, budget int, month rules.Month) []shiftAction {
	var actions []shiftAction
	if budget >= removeFieldBudgetCost {
		for _, c := range slots {
			t := s.tileAt(c)
			if t != nil && t.Type != field.Empty {
				actions = append(actions, shiftAction{
					kind:  "remove",
					coord: c,
					cost:  removeFieldBudgetCost,
				})
			}
		}
	}
	for _, c := range slots {
		vacant := slotIsVacant(c, s)
		existing := s.tileAt(c)
		for _, tileType := range market {
			cost := month.FieldCost(tileType)
			if cost > budget {
				continue
			}
			if !vacant && existing != nil && existing.Type == tileType {
				continue // same-type refresh on live fields: use free shift rotation
			}
			// Any slot can be built on: vacant slots are "place", occupied ones are
			// "overbuild" (full new-field cost, fresh charge/orientation).
			kind := "overbuild"
			if vacant {
				kind = "place"
			}
			actions = append(actions, shiftAction{
				kind:  kind,
				coord: c,
				tile:  tileType,
				cost:  cost,
			})
		}
	}
	return actions
}

func slotIsVacant(c hex.Coord, s *State) bool {
	t := s.tileAt(c)
	if t == nil {
		return true
	}
	return t.Type == field.Empty || t.BurnedOut
}

func applyShiftAction(s *State, act shiftAction, rng *rand.Rand, month rules.Month) {
	switch act.kind {
	case "remove":
		s.Tiles[act.coord.Q][act.coord.R] = field.Tile{Type: field.Empty}
	default:
		placeTile(s, act.coord, act.tile, rng, month)
	}
}

func pickShiftAction(rng *rand.Rand, actions []shiftAction, s *State) shiftAction {
	if len(actions) == 1 {
		return actions[0]
	}
	total := 0
	weights := make([]int, len(actions))
	for i, act := range actions {
		w := 1
		if isBurnedSameTypeRefresh(s, act) {
			w = burnedRefreshWeight
		}
		weights[i] = w
		total += w
	}
	r := rng.Intn(total)
	for i, w := range weights {
		r -= w
		if r < 0 {
			return actions[i]
		}
	}
	return actions[len(actions)-1]
}

func isBurnedSameTypeRefresh(s *State, act shiftAction) bool {
	if act.kind == "remove" {
		return false
	}
	t := s.tileAt(act.coord)
	if t == nil || !t.BurnedOut {
		return false
	}
	return act.tile == t.Type
}

func slotsForPlayer(player1 bool) []hex.Coord {
	out := make([]hex.Coord, 0, len(PlaceableSlots()))
	for _, c := range PlaceableSlots() {
		if c.IsPlayer1() == player1 {
			out = append(out, c)
		}
	}
	return out
}

func fillSlotCosts(rng *rand.Rand, s *State, slots []hex.Coord, target int, label string, monthFilter int, month rules.Month) error {
	if target == 0 {
		return nil
	}
	// Shuffle so low budgets don't always spend on early row-major edge slots.
	slots = append([]hex.Coord(nil), slots...)
	rng.Shuffle(len(slots), func(i, j int) {
		slots[i], slots[j] = slots[j], slots[i]
	})
	planner, err := newCostPlanner(slots, monthFilter, month)
	if err != nil {
		return err
	}
	if !planner.achievable[target] {
		return fmt.Errorf("%w: %s %d Geld (erreichbar: %d-%d)", ErrCostNotAchievable, label, target, planner.minCost, planner.maxCost)
	}

	remaining := target
	for slotIdx, c := range slots {
		choices := planner.validChoices(slotIdx, remaining)
		if len(choices) == 0 {
			return fmt.Errorf("%w: %s %d Geld (erreichbar: %d-%d)", ErrCostNotAchievable, label, target, planner.minCost, planner.maxCost)
		}
		rng.Shuffle(len(choices), func(i, j int) {
			choices[i], choices[j] = choices[j], choices[i]
		})
		t := choices[0]
		placeTile(s, c, t, rng, month)
		remaining -= month.FieldCost(t)
	}
	if remaining != 0 {
		return fmt.Errorf("%w: %s %d Geld (erreichbar: %d-%d)", ErrCostNotAchievable, label, target, planner.minCost, planner.maxCost)
	}
	return nil
}

// randomWithCostOnSlots builds a random board on all slots with exact total cost.
func randomWithCostOnSlots(rng *rand.Rand, target int) (*State, error) {
	if target == 0 {
		return NewEmpty(), nil
	}
	s := NewEmpty()
	if err := fillSlotCosts(rng, s, PlaceableSlots(), target, "gesamt", 0, rules.Month{}); err != nil {
		return nil, err
	}
	return s, nil
}

type costPlanner struct {
	slots      []hex.Coord
	slotTypes  [][]field.Type
	suffix     [][]bool
	achievable map[int]bool
	minCost    int
	maxCost    int
	month      rules.Month
}

func newCostPlanner(slots []hex.Coord, monthFilter int, month rules.Month) (*costPlanner, error) {
	n := len(slots)
	p := &costPlanner{
		slots:      slots,
		slotTypes:  make([][]field.Type, n),
		achievable: map[int]bool{0: true},
		minCost:    0,
		month:      month,
	}
	for i, c := range slots {
		types := make([]field.Type, 0, len(marketFor(c, monthFilter))+1)
		types = append(types, field.Empty)
		types = append(types, marketFor(c, monthFilter)...)
		p.slotTypes[i] = types
		slotMax := 0
		for _, t := range types {
			if cost := month.FieldCost(t); cost > slotMax {
				slotMax = cost
			}
		}
		p.maxCost += slotMax
	}

	reach := map[int]bool{0: true}
	for i := 0; i < n; i++ {
		costs := costsForTypes(p.slotTypes[i], p.month)
		next := make(map[int]bool, len(reach)*len(costs))
		for sum := range reach {
			for _, cost := range costs {
				next[sum+cost] = true
			}
		}
		reach = next
	}
	p.achievable = reach

	p.suffix = make([][]bool, n+1)
	p.suffix[n] = make([]bool, p.maxCost+1)
	p.suffix[n][0] = true
	for i := n - 1; i >= 0; i-- {
		p.suffix[i] = make([]bool, p.maxCost+1)
		costs := costsForTypes(p.slotTypes[i], p.month)
		for sum := 0; sum <= p.maxCost; sum++ {
			for _, cost := range costs {
				if sum >= cost && p.suffix[i+1][sum-cost] {
					p.suffix[i][sum] = true
					break
				}
			}
		}
	}
	return p, nil
}

func (p *costPlanner) validChoices(slotIdx, remaining int) []field.Type {
	if remaining < 0 || remaining >= len(p.suffix[slotIdx]) || !p.suffix[slotIdx][remaining] {
		return nil
	}
	out := make([]field.Type, 0, len(p.slotTypes[slotIdx]))
	for _, t := range p.slotTypes[slotIdx] {
		cost := p.month.FieldCost(t)
		if cost > remaining {
			continue
		}
		if p.suffix[slotIdx+1][remaining-cost] {
			out = append(out, t)
		}
	}
	return out
}

func marketFor(c hex.Coord, monthFilter int) []field.Type {
	types := field.FilterMarket(marketForPlayer(c.IsPlayer1()), monthFilter)
	return field.FilterForCell(types, c)
}

func marketForPlayer(player1 bool) []field.Type {
	if player1 {
		return field.ReactorMarket
	}
	return field.GridMarket
}

func fieldCost(t field.Type) int {
	if t == field.Empty {
		return 0
	}
	return field.Catalog[t].Cost
}

func costsForTypes(types []field.Type, month rules.Month) []int {
	costs := make([]int, len(types))
	for i, t := range types {
		costs[i] = month.FieldCost(t)
	}
	return costs
}

func placeTile(s *State, c hex.Coord, t field.Type, rng *rand.Rand, month rules.Month) {
	if t == field.Empty {
		s.Tiles[c.Q][c.R] = field.Tile{Type: field.Empty}
		return
	}
	if !field.AllowedOnCell(t, c) {
		return
	}
	superTarget := hex.Rotation(0)
	orient := hex.Rotation(0)
	if t == field.Superconductor {
		superTarget = hex.RandomRotation(rng)
	}
	if t == field.Relay || t == field.Mirror || t == field.CoolingTower ||
		t == field.Ground || t == field.EmergencyGenerator {
		orient = hex.RandomRotation(rng)
	}
	s.Tiles[c.Q][c.R] = field.NewTile(t, orient, superTarget)
	if info, ok := field.Catalog[t]; ok && info.InitialCharge >= 0 {
		s.Tiles[c.Q][c.R].Charge = month.InitialCharge(t)
	}
}
