package field_test

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
)

func TestEmergencyGeneratorCostAndRotation(t *testing.T) {
	info := field.Catalog[field.EmergencyGenerator]
	if info.Cost != 2 {
		t.Fatalf("cost = %d, want 2", info.Cost)
	}
	if info.InitialCharge != 1 || info.MaxCharge != 1 {
		t.Fatalf("charge = %d/%d, want 1/1", info.InitialCharge, info.MaxCharge)
	}
	if !field.HasRotation(field.EmergencyGenerator) {
		t.Fatal("notgenerator orientation should affect simulation")
	}
}

func TestAllowedOnCellReactorFuelNotOnTurbineColumn(t *testing.T) {
	wallSlot := hex.Coord{Q: hex.TurbineCol, R: 3}
	if field.AllowedOnCell(field.CoalChamber, wallSlot) {
		t.Fatal("coal must not be placeable on turbine-column wall slots")
	}
	if !field.AllowedOnCell(field.Mirror, wallSlot) {
		t.Fatal("mirror should remain placeable on turbine-column wall slots")
	}
}

func TestAllowedOnCellReactorFuelOnCoreColumns(t *testing.T) {
	core := hex.Coord{Q: 2, R: 2}
	if !field.AllowedOnCell(field.CoalChamber, core) {
		t.Fatal("coal should be placeable on reactor core columns")
	}
}

func TestAllowedOnCellGridOnlyOnPlayer2(t *testing.T) {
	grid := hex.Coord{Q: 6, R: 2}
	reactorCol := hex.Coord{Q: 3, R: 2}
	if !field.AllowedOnCell(field.Relay, grid) {
		t.Fatal("grid fields belong on player 2 cells")
	}
	if field.AllowedOnCell(field.Relay, reactorCol) {
		t.Fatal("grid fields must not be placeable on player 1 cells")
	}
	if field.AllowedOnCell(field.CoalChamber, grid) {
		t.Fatal("reactor fuel must not be placeable on player 2 cells")
	}
}
