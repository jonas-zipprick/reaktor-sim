package board

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
)

func TestValidShiftActionsAllowsOverbuildOnOccupiedSlot(t *testing.T) {
	pos := hex.Coord{Q: 1, R: 1}
	s := NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.NewTile(field.CoalChamber, 0, 0)

	slots := slotsForPlayer(true)
	market := marketFor(pos, 0)
	actions := validShiftActions(s, slots, market, 3)

	var found bool
	for _, a := range actions {
		if a.coord == pos && a.kind == "overbuild" && a.tile == field.GasBoiler {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected overbuild action on occupied coal, got %+v", actions)
	}
}

func TestValidShiftActionsAllowsBuildOnBurnedOutSlot(t *testing.T) {
	pos := hex.Coord{Q: 1, R: 1}
	s := NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.Tile{Type: field.CoalChamber, BurnedOut: true}

	slots := slotsForPlayer(true)
	market := marketFor(pos, 0)
	actions := validShiftActions(s, slots, market, 2)

	var found bool
	for _, a := range actions {
		if a.coord == pos && a.tile == field.CoalChamber && a.cost == 2 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected place action on burned coal slot, got %+v", actions)
	}
}

func TestApplyShiftActionOverbuildRefreshesTile(t *testing.T) {
	pos := hex.Coord{Q: 1, R: 1}
	s := NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.NewTile(field.CoalChamber, 0, 0)
	s.Tiles[pos.Q][pos.R].Charge = 1

	applyShiftAction(s, shiftAction{
		kind:  "overbuild",
		coord: pos,
		tile:  field.GasBoiler,
		cost:  3,
	}, nil)

	tile := s.Tiles[pos.Q][pos.R]
	if tile.Type != field.GasBoiler {
		t.Fatalf("tile type = %v, want gas boiler", tile.Type)
	}
	if tile.Charge != field.Catalog[field.GasBoiler].InitialCharge {
		t.Fatalf("charge = %d, want full initial charge", tile.Charge)
	}
}
