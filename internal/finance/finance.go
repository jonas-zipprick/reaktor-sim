// Package finance defines Ministerium-fuer-Finanzen budget cards.
package finance

import (
	"fmt"
	"strings"
)

// Card is one finance card granting a per-shift budget for the whole month.
// The same budget is available at the start of every shift.
type Card struct {
	ID                 string
	Name               string
	ReactorBudget      int // Spieler 1 (Reaktor) money per shift
	GridBudget         int // Spieler 2 (Stromnetz) money per shift
	SpecialRule        string
	AvailableFromMonth int // earliest campaign month (1 = from start; 0 = from start)
}

// Cards is the standard finance deck from gameRules.md.
var Cards = []Card{
	{
		ID: "schwerindustrie", Name: "Triumph der Schwerindustrie",
		ReactorBudget: 4, GridBudget: 4,
		SpecialRule: "Kohle ist um 1 Geld guenstiger",
	},
	{
		ID: "sparmassnahmen", Name: "Nationale Sparmaßnahmen",
		ReactorBudget: 3, GridBudget: 3,
		AvailableFromMonth: 2,
	},
	{
		ID: "wettruesten", Name: "Nukleares Wettrüsten",
		ReactorBudget: 5, GridBudget: 3,
		SpecialRule: "Reparaturen nicht bewilligt; ausgebrannte Felder duerfen ueberbaut werden",
	},
}

// DefaultCard returns the first finance card.
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

// RepairsAllowed reports whether leftover money may be spent on damage repair.
func (c Card) RepairsAllowed() bool {
	return c.ID != "wettruesten"
}

// AvailableInMonth reports whether this finance card may appear in campaign month m.
// monthFilter 0 means no filter (all cards allowed).
func (c Card) AvailableInMonth(monthFilter int) bool {
	if monthFilter <= 0 {
		return true
	}
	from := c.AvailableFromMonth
	if from <= 0 {
		from = 1
	}
	return monthFilter >= from
}

// Describe formats card name and per-shift budget for logging.
func (c Card) Describe() string {
	return fmt.Sprintf("%s — Reaktor %d Geld | Stromnetz %d Geld je Schicht",
		c.Name, c.ReactorBudget, c.GridBudget)
}
