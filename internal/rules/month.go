// Package rules combines energy and finance ministry cards for one campaign month.
package rules

import (
	"github.com/jonas/reaktor-sim/internal/field"
)

// Month identifies the active energy + finance card pair by slug.
type Month struct {
	EnergyID  string
	FinanceID string
}

// CriticalLimit returns the geloest-chip limit per player half (default 7).
func (m Month) CriticalLimit() int {
	if m.FinanceID == "um-jeden-preis" {
		return 10
	}
	return 7
}

// RepairsAllowed reports whether damage repair is funded this month.
func (m Month) RepairsAllowed() bool {
	return m.FinanceID != "wettruesten"
}

// FieldCost returns placement cost for t under this month's card rules.
func (m Month) FieldCost(t field.Type) int {
	if t == field.Empty {
		return 0
	}
	base, ok := field.Catalog[t]
	if !ok {
		return 0
	}
	cost := base.Cost
	switch t {
	case field.Transformer:
		if m.EnergyID == "netzoptimierung" {
			cost = 1
		}
	case field.UraniumPlate:
		if m.EnergyID == "technologische-transformation" {
			cost = 4
		}
		if m.FinanceID == "schwerindustrie" {
			cost--
		}
		if cost < 1 {
			cost = 1
		}
	}
	return cost
}

// InitialCharge returns the starting bound charge when placing t.
func (m Month) InitialCharge(t field.Type) int {
	info, ok := field.Catalog[t]
	if !ok || info.InitialCharge < 0 {
		return info.InitialCharge
	}
	charge := info.InitialCharge
	if m.EnergyID == "gossnab" && (t == field.CoalChamber || t == field.Transformer) {
		charge--
	}
	if charge < 0 {
		return 0
	}
	return charge
}

// AbsorberCoolingDisabled is true during "Testlauf unter Volllast".
func (m Month) AbsorberCoolingDisabled() bool {
	return m.EnergyID == "testlauf-volllast"
}

// DoubleIgniterHeat is true during "Schturmowschtschina".
func (m Month) DoubleIgniterHeat() bool {
	return m.EnergyID == "schturmowschtschina"
}
