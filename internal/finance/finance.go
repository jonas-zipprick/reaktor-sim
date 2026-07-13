// Package finance defines Ministerium-fuer-Finanzen budget cards.
package finance

import (
	"fmt"
	"strings"
)

// Card is one finance card granting a per-shift budget for the whole month.
// The same budget is available at the start of every shift.
type Card struct {
	ID            string
	Name          string
	ReactorBudget int // Spieler 1 (Reaktor) money per shift
	GridBudget    int // Spieler 2 (Stromnetz) money per shift
	SpecialRule   string
}

// Cards is the standard finance deck from gameRules.md.
var Cards = []Card{
	{
		ID: "schwerindustrie", Name: "Triumph der Schwerindustrie",
		ReactorBudget: 3, GridBudget: 4,
		SpecialRule: "Uran ist um 1 Geld guenstiger",
	},
	{
		ID: "sparmassnahmen", Name: "Nationale Sparmaßnahmen",
		ReactorBudget: 2, GridBudget: 2,
	},
	{
		ID: "wettruesten", Name: "Nukleares Wettrüsten",
		ReactorBudget: 2, GridBudget: 4,
		SpecialRule: "Reparaturen nicht bewilligt; ausgebrannte Felder duerfen ueberbaut werden",
	},
	{
		ID: "um-jeden-preis", Name: "Um jeden Preis",
		ReactorBudget: 2, GridBudget: 2,
		SpecialRule: "Kritische Masse liegt bei 10 statt 7 Chips",
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

// CriticalLimit returns the geloest-chip limit per player half for this month.
func (c Card) CriticalLimit() int {
	if c.ID == "um-jeden-preis" {
		return 10
	}
	return 7
}

// Describe formats card name and per-shift budget for logging.
func (c Card) Describe() string {
	return fmt.Sprintf("%s — Reaktor %d Geld | Stromnetz %d Geld je Schicht",
		c.Name, c.ReactorBudget, c.GridBudget)
}
