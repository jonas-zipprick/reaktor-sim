package board

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
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
// A target of 0 leaves that half empty (no fields placed).
func RandomWithPlayerCosts(rng *rand.Rand, player1, player2 int) (*State, error) {
	if player1 < 0 || player2 < 0 {
		return nil, fmt.Errorf("Brettkosten duerfen nicht negativ sein (Spieler 1: %d, Spieler 2: %d)", player1, player2)
	}
	if player1 == 0 && player2 == 0 {
		return NewEmpty(), nil
	}

	s := NewEmpty()
	if err := fillSlotCosts(rng, s, slotsForPlayer(true), player1, "Spieler 1 (Reaktor)"); err != nil {
		return nil, err
	}
	if err := fillSlotCosts(rng, s, slotsForPlayer(false), player2, "Spieler 2 (Stromnetz)"); err != nil {
		return nil, err
	}
	return s, nil
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

func fillSlotCosts(rng *rand.Rand, s *State, slots []hex.Coord, target int, label string) error {
	if target == 0 {
		return nil
	}
	planner, err := newCostPlanner(slots)
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
		placeTile(s, c, t, rng)
		remaining -= fieldCost(t)
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
	if err := fillSlotCosts(rng, s, PlaceableSlots(), target, "gesamt"); err != nil {
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
}

func newCostPlanner(slots []hex.Coord) (*costPlanner, error) {
	n := len(slots)
	p := &costPlanner{
		slots:      slots,
		slotTypes:  make([][]field.Type, n),
		achievable: map[int]bool{0: true},
		minCost:    0,
	}
	for i, c := range slots {
		types := make([]field.Type, 0, len(marketFor(c))+1)
		types = append(types, field.Empty)
		types = append(types, marketFor(c)...)
		p.slotTypes[i] = types
		slotMax := 0
		for _, t := range types {
			if cost := fieldCost(t); cost > slotMax {
				slotMax = cost
			}
		}
		p.maxCost += slotMax
	}

	reach := map[int]bool{0: true}
	for i := 0; i < n; i++ {
		costs := costsForTypes(p.slotTypes[i])
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
		costs := costsForTypes(p.slotTypes[i])
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
		cost := fieldCost(t)
		if cost > remaining {
			continue
		}
		if p.suffix[slotIdx+1][remaining-cost] {
			out = append(out, t)
		}
	}
	return out
}

func marketFor(c hex.Coord) []field.Type {
	if c.IsPlayer2() {
		return field.GridMarket
	}
	return field.ReactorMarket
}

func fieldCost(t field.Type) int {
	if t == field.Empty {
		return 0
	}
	return field.Catalog[t].Cost
}

func costsForTypes(types []field.Type) []int {
	costs := make([]int, len(types))
	for i, t := range types {
		costs[i] = fieldCost(t)
	}
	return costs
}

func placeTile(s *State, c hex.Coord, t field.Type, rng *rand.Rand) {
	if t == field.Empty {
		s.Tiles[c.Q][c.R] = field.Tile{Type: field.Empty}
		return
	}
	superTarget := hex.Rotation(0)
	orient := hex.Rotation(0)
	if t == field.Superconductor {
		superTarget = hex.RandomRotation(rng)
	}
	if t == field.Relay || t == field.Mirror {
		orient = hex.RandomRotation(rng)
	}
	s.Tiles[c.Q][c.R] = field.NewTile(t, orient, superTarget)
}
