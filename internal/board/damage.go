package board

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/jonas/reaktor-sim/internal/hex"
)

// TotalDamage returns accumulated damage chips for zone z.
func (s *State) TotalDamage(z Zone) int {
	return s.Damage[int(z)]
}

// AddZoneDamage adds one damage chip to a zone's Schadensbereich.
func (s *State) AddZoneDamage(z Zone) {
	s.Damage[int(z)]++
}

// AddWallDamage places one damage chip on a zone for an outer-wall voltage hit
// when no demand chip was available. The zone is chosen from the wall crossing.
func (s *State) AddWallDamage(from hex.Coord, travelDir int, rng *rand.Rand) (Zone, bool) {
	zones := ZonesForWallDemandHit(from, travelDir)
	if len(zones) == 0 {
		return 0, false
	}
	z := zones[rng.Intn(len(zones))]
	s.AddZoneDamage(z)
	return z, true
}

// BorderDamageEvent formats a trace event for zone damage from border overload.
func BorderDamageEvent(z Zone) string {
	return fmt.Sprintf("Rand-Schaden %s", z.String())
}

// ZoneFromBorderDamageEvent parses the zone from a BorderDamageEvent string.
func ZoneFromBorderDamageEvent(event string) (Zone, bool) {
	const prefix = "Rand-Schaden "
	if !strings.HasPrefix(event, prefix) {
		return 0, false
	}
	name := strings.TrimPrefix(event, prefix)
	for z := ZoneIndustry; z <= ZonePlant; z++ {
		if z.String() == name {
			return z, true
		}
	}
	return 0, false
}

// TotalBoardDamage returns all damage chips on the board (zones + igniter).
func (s *State) TotalBoardDamage() int {
	return s.TotalPlayer2Damage() + s.EmitterDamage
}

// TotalPlayer2Damage sums damage chips across all player-2 zones.
func (s *State) TotalPlayer2Damage() int {
	total := 0
	for z := ZoneIndustry; z <= ZonePlant; z++ {
		total += s.TotalDamage(z)
	}
	return total
}

const repairCostPerChip = 1

// AddEmitterDamage adds one damage chip to the igniter.
func (s *State) AddEmitterDamage() {
	s.EmitterDamage++
}

// RepairEmitterDamage removes igniter damage at 1 money per chip until budget
// or damage runs out. Returns money spent.
func (s *State) RepairEmitterDamage(budget int) int {
	if budget <= 0 || s.EmitterDamage <= 0 {
		return 0
	}
	if budget > s.EmitterDamage {
		budget = s.EmitterDamage
	}
	s.EmitterDamage -= budget
	return budget
}

// RepairRandomDamage removes damage chips at 1 money each, choosing zones
// uniformly at random until budget is spent or no damage remains.
// Returns money spent on repairs.
func (s *State) RepairRandomDamage(rng *rand.Rand, budget int) int {
	if budget < repairCostPerChip {
		return 0
	}
	spent := 0
	for budget >= repairCostPerChip {
		candidates := zonesWithDamage(s)
		if len(candidates) == 0 {
			break
		}
		z := candidates[rng.Intn(len(candidates))]
		s.Damage[int(z)]--
		budget -= repairCostPerChip
		spent += repairCostPerChip
	}
	return spent
}

const damageHighThreshold = 3

// repairChance returns the probability that a player half decides to repair,
// depending on the number of damage chips on that half.
func repairChance(damage int, high bool) float64 {
	if damage <= 0 {
		return 0
	}
	if damage > damageHighThreshold {
		if high {
			return 0.80
		}
		return 0.50
	}
	if high {
		return 0.40
	}
	return 0.20
}

// HalfDamage returns the damage count relevant for one player half.
func (s *State) HalfDamage(player1 bool) int {
	if player1 {
		return s.EmitterDamage
	}
	return s.TotalPlayer2Damage()
}

// RepairHalf repairs damage on one player half, spending a random amount
// (1 to damage, capped by budget) at 1 money per chip.
// Returns money spent.
func (s *State) RepairHalf(rng *rand.Rand, player1 bool, budget int) int {
	damage := s.HalfDamage(player1)
	if budget <= 0 || damage <= 0 {
		return 0
	}
	maxRepair := damage
	if maxRepair > budget {
		maxRepair = budget
	}
	amount := 1 + rng.Intn(maxRepair)
	if player1 {
		return s.RepairEmitterDamage(amount)
	}
	return s.RepairRandomDamage(rng, amount)
}

// MaybeRepair rolls against the repair chance for one player half.
// If the roll succeeds, it repairs a random amount and returns the money spent.
// preFieldPurchase distinguishes between the pre- and post-purchase repair phases.
func (s *State) MaybeRepair(rng *rand.Rand, player1 bool, budget int, preFieldPurchase bool) int {
	damage := s.HalfDamage(player1)
	chance := repairChance(damage, !preFieldPurchase)
	if chance <= 0 || rng.Float64() >= chance {
		return 0
	}
	return s.RepairHalf(rng, player1, budget)
}

func zonesWithDamage(s *State) []Zone {
	out := make([]Zone, 0)
	for z := ZoneIndustry; z <= ZonePlant; z++ {
		for i := 0; i < s.TotalDamage(z); i++ {
			out = append(out, z)
		}
	}
	return out
}
