package board

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/jonas/reaktor-sim/internal/hex"
)

// ShiftDemands is the per-shift demand quota from the energy card.
type ShiftDemands struct {
	Industry    int
	Residential int
	Rail        int
	Plant       int
}

// DefaultShiftDemands matches Eröffnungsfeier Schicht 1 (see internal/energy).
func DefaultShiftDemands() ShiftDemands {
	return ShiftDemands{Industry: 1, Residential: 0, Rail: 0, Plant: 1}
}

var industryCells = []hex.Coord{{Q: 6, R: 0}, {Q: 7, R: 0}, {Q: 8, R: 0}}
var railCells = []hex.Coord{{Q: 6, R: 4}, {Q: 7, R: 4}, {Q: 8, R: 4}}
var residentialCells = []hex.Coord{{Q: 8, R: 1}, {Q: 8, R: 2}, {Q: 8, R: 3}}
var plantCells = []hex.Coord{{Q: hex.TurbineCol, R: hex.TurbineRow}}

// ApplyDemands adds demand chips on wired border cells for a new shift.
func (s *State) ApplyDemands(d ShiftDemands) {
	if s.Demands == nil {
		s.Demands = make(map[hex.Coord][4]int)
	}
	distributeZone(s, ZoneIndustry, d.Industry, industryCells)
	distributeZone(s, ZoneRail, d.Rail, railCells)
	distributeZone(s, ZoneResidential, d.Residential, residentialCells)
	distributeZone(s, ZonePlant, d.Plant, plantCells)
}

func distributeZone(s *State, z Zone, total int, cells []hex.Coord) {
	if total <= 0 {
		return
	}
	valid := make([]hex.Coord, 0, len(cells))
	for _, c := range cells {
		if !c.Valid() {
			continue
		}
		hasZone := false
		for _, wired := range ZonesOf(c) {
			if wired == z {
				hasZone = true
				break
			}
		}
		if hasZone {
			valid = append(valid, c)
		}
	}
	if len(valid) == 0 {
		return
	}
	per := total / len(valid)
	rem := total % len(valid)
	for i, c := range valid {
		n := per
		if i < rem {
			n++
		}
		if n == 0 {
			continue
		}
		cell := s.Demands[c]
		cell[int(z)] += n
		s.Demands[c] = cell
	}
}

// DemandAt returns remaining demand chips for zone z on cell c.
func (s *State) DemandAt(c hex.Coord, z Zone) int {
	if s.Demands == nil {
		return 0
	}
	cell, ok := s.Demands[c]
	if !ok {
		return 0
	}
	return cell[int(z)]
}

// TotalDemand returns remaining demand chips for a zone across all cells.
func (s *State) TotalDemand(z Zone) int {
	total := 0
	for c := range s.Demands {
		total += s.DemandAt(c, z)
	}
	return total
}

// ZonesForOuterWallHit returns demand zones for the outer wall crossed when
// moving from in travelDir. Which wall is hit depends on position and direction.
func ZonesForOuterWallHit(from hex.Coord, travelDir int) []Zone {
	if !from.IsPlayer2() {
		return nil
	}
	next := from.Neighbor(travelDir)
	var zones []Zone

	if next.R < 0 {
		zones = append(zones, ZoneIndustry)
	}
	if next.R >= hex.Rows {
		zones = append(zones, ZoneRail)
	}
	if next.Q >= hex.Cols {
		switch from.R {
		case 0:
			zones = append(zones, ZoneIndustry)
		case hex.Rows - 1:
			zones = append(zones, ZoneRail)
		default:
			zones = append(zones, ZoneResidential)
		}
	}
	if next.Q < 0 {
		zones = append(zones, ZonePlant)
	}

	return zones
}

// PlantWallZone reports whether travelDir from is a fixed Reaktoreigenbedarf wall.
func PlantWallZone(from hex.Coord, travelDir int) (Zone, bool) {
	return plantWallZone(from, travelDir)
}

// PlantWallHits lists fixed Reaktoreigenbedarf walls beside the turbine column.
func PlantWallHits() []struct {
	From hex.Coord
	Dir  int
} {
	t := hex.Coord{Q: hex.TurbineCol, R: hex.TurbineRow}
	return []struct {
		From hex.Coord
		Dir  int
	}{
		{hex.Coord{Q: hex.TurbineCol, R: 1}, hex.RotNW.TravelDir()},
		{hex.Coord{Q: hex.TurbineCol, R: 1}, hex.RotW.TravelDir()},
		{hex.Coord{Q: hex.TurbineCol, R: 3}, hex.RotSW.TravelDir()},
		{hex.Coord{Q: hex.TurbineCol, R: 3}, hex.RotW.TravelDir()},
		{t, hex.RotNW.TravelDir()},
		{t, hex.RotSW.TravelDir()},
		{t, hex.RotW.TravelDir()},
	}
}

// TurbinePlantWallLeave reports whether a voltage chip leaving the turbine in
// travelDir should satisfy Reaktoreigenbedarf before entering a neighbor cell.
func TurbinePlantWallLeave(from hex.Coord, travelDir int) bool {
	if !from.IsTurbine() {
		return false
	}
	switch travelDir {
	case hex.RotNW.TravelDir(), hex.RotSW.TravelDir():
		_, ok := PlantWallZone(from, travelDir)
		return ok
	default:
		return false
	}
}

func plantWallZone(from hex.Coord, travelDir int) (Zone, bool) {
	for _, hit := range PlantWallHits() {
		if hit.From == from && hit.Dir == travelDir {
			return ZonePlant, true
		}
	}
	return 0, false
}

// ZonesForWallDemandHit returns zones whose border demand may be satisfied when
// a voltage chip leaves from in travelDir and hits a wired wall.
func ZonesForWallDemandHit(from hex.Coord, travelDir int) []Zone {
	zones := ZonesForOuterWallHit(from, travelDir)
	if z, ok := plantWallZone(from, travelDir); ok {
		zones = appendZoneIfMissing(zones, z)
	}
	return zones
}

func appendZoneIfMissing(zones []Zone, z Zone) []Zone {
	for _, existing := range zones {
		if existing == z {
			return zones
		}
	}
	return append(zones, z)
}

// TryConsumeWallDemand removes one demand chip for a zone allowed by the wall
// hit when leaving from in travelDir. The field under the chip does not matter.
func (s *State) TryConsumeWallDemand(from hex.Coord, travelDir int, rng *rand.Rand) (Zone, bool) {
	zones := ZonesForWallDemandHit(from, travelDir)
	eligible := make([]Zone, 0, len(zones))
	for _, z := range zones {
		if s.TotalDemand(z) > 0 {
			eligible = append(eligible, z)
		}
	}
	if len(eligible) == 0 {
		return 0, false
	}
	z := pickWallDemandZone(eligible)
	if !s.TryConsumeZone(z, rng) {
		return 0, false
	}
	return z, true
}

func pickWallDemandZone(eligible []Zone) Zone {
	if len(eligible) == 1 {
		return eligible[0]
	}
	for _, pref := range []Zone{ZoneIndustry, ZonePlant, ZoneResidential, ZoneRail} {
		for _, cand := range eligible {
			if cand == pref {
				return cand
			}
		}
	}
	return eligible[0]
}

// TryConsumeZone removes one demand chip for zone z from any wired cell.
func (s *State) TryConsumeZone(z Zone, rng *rand.Rand) bool {
	candidates := make([]hex.Coord, 0)
	for c, zones := range s.Demands {
		if zones[int(z)] > 0 {
			candidates = append(candidates, c)
		}
	}
	if len(candidates) == 0 {
		return false
	}
	c := candidates[rng.Intn(len(candidates))]
	cell := s.Demands[c]
	cell[int(z)]--
	s.Demands[c] = cell
	return true
}

// TryConsumeDemand removes one chip for a random zone with demand at c.
func (s *State) TryConsumeDemand(c hex.Coord, rng *rand.Rand) (Zone, bool) {
	zones := ZonesOf(c)
	available := make([]Zone, 0, len(zones))
	for _, z := range zones {
		if s.DemandAt(c, z) > 0 {
			available = append(available, z)
		}
	}
	if len(available) == 0 {
		return 0, false
	}
	z := available[rng.Intn(len(available))]
	cell := s.Demands[c]
	cell[int(z)]--
	s.Demands[c] = cell
	return z, true
}

// BorderDemandEvent formats a trace event for satisfying border demand in zone z.
func BorderDemandEvent(z Zone) string {
	return fmt.Sprintf("Rand-Bedarf %s erfuellt", z.String())
}

// ZoneFromBorderDemandEvent parses the zone from a BorderDemandEvent string.
func ZoneFromBorderDemandEvent(event string) (Zone, bool) {
	const prefix = "Rand-Bedarf "
	const suffix = " erfuellt"
	if !strings.HasPrefix(event, prefix) || !strings.HasSuffix(event, suffix) {
		return 0, false
	}
	name := strings.TrimSuffix(strings.TrimPrefix(event, prefix), suffix)
	for z := ZoneIndustry; z <= ZonePlant; z++ {
		if z.String() == name {
			return z, true
		}
	}
	return 0, false
}

// DemandLabel formats remaining demand chips on a wired border cell.
func (s *State) DemandLabel(c hex.Coord) string {
	zones := ZonesOf(c)
	if len(zones) == 0 {
		return ""
	}
	parts := make([]string, 0, len(zones))
	for _, z := range zones {
		n := s.DemandAt(c, z)
		if n > 0 {
			parts = append(parts, fmt.Sprintf("%s%d", zoneLetter(z), n))
		}
	}
	return strings.Join(parts, "")
}

// ZoneLetter returns the single-letter demand zone code (I, W, b, R).
func ZoneLetter(z Zone) string {
	return zoneLetter(z)
}

func zoneLetter(z Zone) string {
	switch z {
	case ZoneIndustry:
		return "I"
	case ZoneResidential:
		return "W"
	case ZoneRail:
		return "b"
	case ZonePlant:
		return "R"
	default:
		return "?"
	}
}
