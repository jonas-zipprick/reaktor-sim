// Package board manages the game state and random board generation.
package board

import (
	"math/rand"

	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/rules"
)

// Zone is a demand region on the player-2 border.
type Zone int

const (
	ZoneIndustry Zone = iota
	ZoneResidential
	ZoneRail
	ZonePlant
)

func (z Zone) String() string {
	switch z {
	case ZoneIndustry:
		return "Industrie"
	case ZoneResidential:
		return "Wohnviertel"
	case ZoneRail:
		return "Bahn"
	case ZonePlant:
		return "Reaktoreigenbedarf"
	default:
		return "unbekannt"
	}
}

// State is a full board placement with tile data.
type State struct {
	Tiles         [hex.Cols][hex.Rows]field.Tile
	Demands       map[hex.Coord][4]int
	Damage        [4]int // per Zone: accumulated Schadens-Chips
	EmitterDamage int    // Schadens-Chips on the Zuender (player 1)
}

func NewEmpty() *State {
	return &State{}
}

func (s *State) tileAt(c hex.Coord) *field.Tile {
	if !c.Valid() {
		return nil
	}
	return &s.Tiles[c.Q][c.R]
}

// ZonesOf returns demand zones wired to a cell per gameRules.md layout.
func ZonesOf(c hex.Coord) []Zone {
	if !c.Valid() {
		return nil
	}
	if c.IsTurbine() {
		return []Zone{ZonePlant}
	}
	if !c.IsPlayer2() {
		return nil
	}
	switch {
	case c.Q == 8:
		return []Zone{ZoneResidential}
	case c.R == 0 && c.Q == 5:
		return []Zone{ZoneIndustry, ZonePlant}
	case c.R == 0 && (c.Q == 6 || c.Q == 7):
		return []Zone{ZoneIndustry}
	case c.R == 2 && c.Q == 5:
		return []Zone{ZoneRail, ZonePlant}
	case c.R == 2 && (c.Q == 6 || c.Q == 7):
		return []Zone{ZoneRail}
	default:
		return nil
	}
}

// ZoneOf returns the primary demand zone for a wired border cell.
func ZoneOf(c hex.Coord) (Zone, bool) {
	zones := ZonesOf(c)
	if len(zones) == 0 {
		return 0, false
	}
	return zones[0], true
}

// Random generates a random legal board with fields on placeable slots.
// monthFilter limits purchasable fields (0 = no filter).
func Random(rng *rand.Rand, monthFilter int) *State {
	s := NewEmpty()
	for _, c := range PlaceableSlots() {
		if rng.Float64() >= 0.75 {
			continue
		}
		market := marketFor(c, monthFilter)
		if len(market) == 0 {
			continue
		}
		t := market[rng.Intn(len(market))]
		placeTile(s, c, t, rng, rules.Month{})
	}
	return s
}

// Clone deep-copies tile state.
func (s *State) Clone() *State {
	c := &State{}
	for q := 0; q < hex.Cols; q++ {
		for r := 0; r < hex.Rows; r++ {
			c.Tiles[q][r] = s.Tiles[q][r]
		}
	}
	if s.Demands != nil {
		c.Demands = make(map[hex.Coord][4]int, len(s.Demands))
		for coord, zones := range s.Demands {
			c.Demands[coord] = zones
		}
	}
	c.Damage = s.Damage
	c.EmitterDamage = s.EmitterDamage
	return c
}
