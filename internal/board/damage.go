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
	zones := ZonesForOuterWallHit(from, travelDir)
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

// TotalPlayer2Damage sums damage chips across all player-2 zones.
func (s *State) TotalPlayer2Damage() int {
	total := 0
	for z := ZoneIndustry; z <= ZonePlant; z++ {
		total += s.TotalDamage(z)
	}
	return total
}

const repairCostPerChip = 1

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

func zonesWithDamage(s *State) []Zone {
	out := make([]Zone, 0)
	for z := ZoneIndustry; z <= ZonePlant; z++ {
		for i := 0; i < s.TotalDamage(z); i++ {
			out = append(out, z)
		}
	}
	return out
}
