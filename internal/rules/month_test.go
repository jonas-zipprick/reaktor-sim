package rules_test

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/rules"
)

func TestFieldCostDiscounts(t *testing.T) {
	if got := (rules.Month{}).FieldCost(field.Transformer); got != 3 {
		t.Fatalf("transformer base = %d, want 3", got)
	}

	netz := rules.Month{EnergyID: "netzoptimierung"}
	if got := netz.FieldCost(field.Transformer); got != 1 {
		t.Fatalf("transformer = %d, want 1", got)
	}

	tech := rules.Month{EnergyID: "technologische-transformation"}
	if got := tech.FieldCost(field.UraniumPlate); got != 4 {
		t.Fatalf("uran tech = %d, want 4", got)
	}

	ind := rules.Month{FinanceID: "schwerindustrie"}
	if got := ind.FieldCost(field.UraniumPlate); got != 4 {
		t.Fatalf("uran schwerindustrie = %d, want 4", got)
	}

	stacked := rules.Month{EnergyID: "technologische-transformation", FinanceID: "schwerindustrie"}
	if got := stacked.FieldCost(field.UraniumPlate); got != 3 {
		t.Fatalf("uran stacked = %d, want 3", got)
	}
}

func TestInitialChargeModifiers(t *testing.T) {
	eroeff := rules.Month{EnergyID: "eroeffnungsfeier"}
	if got := eroeff.InitialCharge(field.EmergencyGenerator); got != 2 {
		t.Fatalf("generator = %d, want 2", got)
	}

	gossnab := rules.Month{EnergyID: "gossnab"}
	if got := gossnab.InitialCharge(field.CoalChamber); got != 3 {
		t.Fatalf("coal = %d, want 3", got)
	}
	if got := gossnab.InitialCharge(field.Transformer); got != 5 {
		t.Fatalf("transformer = %d, want 5", got)
	}
}

func TestCriticalLimitUmJedenPreis(t *testing.T) {
	m := rules.Month{FinanceID: "um-jeden-preis"}
	if got := m.CriticalLimit(); got != 10 {
		t.Fatalf("limit = %d, want 10", got)
	}
}
