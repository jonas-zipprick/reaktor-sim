package rules_test

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/rules"
)

func TestFieldCostDiscounts(t *testing.T) {
	if got := (rules.Month{}).FieldCost(field.Transformer); got != 2 {
		t.Fatalf("transformer base = %d, want 2", got)
	}
	if got := (rules.Month{}).FieldCost(field.DistributionStation); got != 2 {
		t.Fatalf("distribution station base = %d, want 2", got)
	}

	netz := rules.Month{EnergyID: "netzoptimierung"}
	if got := netz.FieldCost(field.DistributionStation); got != 1 {
		t.Fatalf("distribution station netz = %d, want 1", got)
	}
	if got := netz.FieldCost(field.Transformer); got != 2 {
		t.Fatalf("transformer under netz = %d, want 2 (no discount)", got)
	}

	tech := rules.Month{EnergyID: "technologische-transformation"}
	if got := tech.FieldCost(field.UraniumPlate); got != 4 {
		t.Fatalf("uran tech = %d, want 4", got)
	}

	ind := rules.Month{FinanceID: "schwerindustrie"}
	if got := ind.FieldCost(field.CoalChamber); got != 2 {
		t.Fatalf("coal schwerindustrie = %d, want 2", got)
	}
	if got := ind.FieldCost(field.UraniumPlate); got != 5 {
		t.Fatalf("uran schwerindustrie = %d, want 5 (no discount)", got)
	}
}

func TestInitialChargeModifiers(t *testing.T) {
	gossnab := rules.Month{EnergyID: "gossnab"}
	if got := gossnab.InitialCharge(field.CoalChamber); got != 7 {
		t.Fatalf("coal = %d, want 7", got)
	}
	if got := gossnab.InitialCharge(field.Transformer); got != 5 {
		t.Fatalf("transformer = %d, want 5", got)
	}
}

func TestCriticalLimitEroeffnungsfeier(t *testing.T) {
	m := rules.Month{EnergyID: "eroeffnungsfeier"}
	if got := m.CriticalLimit(); got != 8 {
		t.Fatalf("limit = %d, want 8", got)
	}
}

func TestCriticalLimitUmJedenPreis(t *testing.T) {
	m := rules.Month{EnergyID: "um-jeden-preis"}
	if got := m.CriticalLimit(); got != 10 {
		t.Fatalf("limit = %d, want 10", got)
	}
}
