// Package energy defines Ministerium-fuer-Energiewirtschaft cards and shift plans.
package energy

import (
	"fmt"
	"strings"

	"github.com/jonas/reaktor-sim/internal/board"
)

// Card is one energy quota card with a five-shift demand plan.
type Card struct {
	ID          string
	Name        string
	Level       int
	SpecialRule string
	Shifts      [5]board.ShiftDemands
}

// Cards is the standard deck order from gameRules.md (2x L1, 2x L2, 2x L3).
var Cards = []Card{
	{
		ID: "eroeffnungsfeier", Name: "Eröffnungsfeier", Level: 1,
		SpecialRule: "Kritische Chip-Grenze liegt bei 8 statt 7",
		Shifts: [5]board.ShiftDemands{
			{Industry: 1, Residential: 1, Rail: 0, Plant: 1},
			{Industry: 1, Residential: 1, Rail: 0, Plant: 1},
			{Industry: 0, Residential: 1, Rail: 1, Plant: 1},
			{Industry: 1, Residential: 1, Rail: 1, Plant: 1},
			{Industry: 2, Residential: 1, Rail: 1, Plant: 1},
		},
	},
	{
		ID: "netzoptimierung", Name: "Optimierung des lokalen Netzes", Level: 1,
		SpecialRule: "Transformatoren kosten 2 Geld statt 3 Geld",
		Shifts: [5]board.ShiftDemands{
			{Industry: 0, Residential: 1, Rail: 1, Plant: 1},
			{Industry: 0, Residential: 1, Rail: 1, Plant: 1},
			{Industry: 0, Residential: 2, Rail: 0, Plant: 1},
			{Industry: 1, Residential: 1, Rail: 1, Plant: 1},
			{Industry: 1, Residential: 2, Rail: 1, Plant: 1},
		},
	},
	{
		ID: "technologische-transformation", Name: "Die technologische Transformation", Level: 2,
		SpecialRule: "Uran-Platten kosten diesen Monat 4 Geld statt 5",
		Shifts: [5]board.ShiftDemands{
			{Industry: 1, Residential: 1, Rail: 0, Plant: 1},
			{Industry: 2, Residential: 1, Rail: 0, Plant: 1},
			{Industry: 1, Residential: 1, Rail: 1, Plant: 1},
			{Industry: 2, Residential: 1, Rail: 0, Plant: 1},
			{Industry: 2, Residential: 1, Rail: 1, Plant: 1},
		},
	},
	{
		ID: "gossnab", Name: "Der Gossnab-Liefervertrag", Level: 2,
		SpecialRule: "Minderwertige Lieferungen: neue Kohle-Brennkammern und Transformatoren starten mit 1 Ladung weniger",
		Shifts: [5]board.ShiftDemands{
			{Industry: 1, Residential: 1, Rail: 0, Plant: 1},
			{Industry: 2, Residential: 0, Rail: 1, Plant: 1},
			{Industry: 0, Residential: 3, Rail: 1, Plant: 1},
			{Industry: 1, Residential: 1, Rail: 0, Plant: 1},
			{Industry: 2, Residential: 1, Rail: 1, Plant: 1},
		},
	},
	{
		ID: "testlauf-volllast", Name: "Testlauf unter Volllast", Level: 3,
		SpecialRule: "Absorber-Staebe und Kuehltuerme verlieren fuer dieses Jahr ihre Funktion",
		Shifts: [5]board.ShiftDemands{
			{Industry: 1, Residential: 1, Rail: 1, Plant: 1},
			{Industry: 2, Residential: 1, Rail: 1, Plant: 1},
			{Industry: 2, Residential: 2, Rail: 1, Plant: 1},
			{Industry: 3, Residential: 1, Rail: 1, Plant: 1},
			{Industry: 2, Residential: 1, Rail: 2, Plant: 1},
		},
	},
	{
		ID: "schturmowschtschina", Name: "Schturmowschtschina (Die Sturmarbeit)", Level: 3,
		SpecialRule: "Spieler 1 muss 2 Waerme-Chips am Zuender feuern (statt 1)",
		Shifts: [5]board.ShiftDemands{
			{Industry: 1, Residential: 1, Rail: 1, Plant: 1},
			{Industry: 2, Residential: 2, Rail: 1, Plant: 1},
			{Industry: 3, Residential: 2, Rail: 2, Plant: 2},
			{Industry: 4, Residential: 1, Rail: 2, Plant: 2},
			{Industry: 4, Residential: 2, Rail: 1, Plant: 2},
		},
	},
}

// DefaultCard returns the first level-1 energy card.
func DefaultCard() Card {
	return Cards[0]
}

// ByID finds a card by slug (case-insensitive).
func ByID(id string) (Card, bool) {
	id = strings.ToLower(strings.TrimSpace(id))
	for _, c := range Cards {
		if c.ID == id {
			return c, true
		}
	}
	return Card{}, false
}

// CriticalLimit returns the geloest-chip limit per player half for this month.
func (c Card) CriticalLimit() int {
	switch c.ID {
	case "um-jeden-preis":
		return 10
	case "eroeffnungsfeier":
		return 8
	default:
		return 7
	}
}

// ShiftDemands returns the quota for shift 1..5.
func (c Card) ShiftDemands(shift int) board.ShiftDemands {
	if shift < 1 || shift > len(c.Shifts) {
		return board.ShiftDemands{}
	}
	return c.Shifts[shift-1]
}

// DescribeShift formats card, shift number and demand totals for logging.
func (c Card) DescribeShift(shift int) string {
	d := c.ShiftDemands(shift)
	return fmt.Sprintf(
		"%s (Stufe %d), Schicht %d — I=%d W=%d b=%d R=%d",
		c.Name, c.Level, shift, d.Industry, d.Residential, d.Rail, d.Plant,
	)
}
