package board

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/rules"
)

func TestValidShiftActionsAllowsOverbuildOnOccupiedSlot(t *testing.T) {
	pos := hex.Coord{Q: 1, R: 1}
	s := NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.NewTile(field.CoalChamber, 0, 0)

	slots := slotsForPlayer(true)
	market := marketFor(pos, 0)
	actions := validShiftActions(s, slots, market, 3, rules.Month{})

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

func TestValidShiftActionsSkipsSameTypeOverbuildOnLiveField(t *testing.T) {
	pos := hex.Coord{Q: 1, R: 1}
	s := NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.NewTile(field.Mirror, hex.RotE, 0)

	actions := validShiftActions(s, slotsForPlayer(true), marketFor(pos, 0), 3, rules.Month{})
	for _, a := range actions {
		if a.coord == pos && a.tile == field.Mirror {
			t.Fatalf("same-type overbuild on live mirror should be unavailable, got %+v", a)
		}
	}
}

func TestValidShiftActionsAllowsBuildOnBurnedOutSlot(t *testing.T) {
	pos := hex.Coord{Q: 1, R: 1}
	s := NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.Tile{Type: field.CoalChamber, BurnedOut: true}

	slots := slotsForPlayer(true)
	market := marketFor(pos, 0)
	actions := validShiftActions(s, slots, market, 2, rules.Month{})

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
	}, nil, rules.Month{})

	tile := s.Tiles[pos.Q][pos.R]
	if tile.Type != field.GasBoiler {
		t.Fatalf("tile type = %v, want gas boiler", tile.Type)
	}
	if tile.Charge != field.Catalog[field.GasBoiler].InitialCharge {
		t.Fatalf("charge = %d, want full initial charge", tile.Charge)
	}
}

func TestPickShiftActionPrefersBurnedSameTypeRefresh(t *testing.T) {
	pos := hex.Coord{Q: 1, R: 1}
	s := NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.Tile{Type: field.CoalChamber, BurnedOut: true}

	refresh := shiftAction{kind: "place", coord: pos, tile: field.CoalChamber, cost: 2}
	other := shiftAction{kind: "place", coord: pos, tile: field.Mirror, cost: 1}
	actions := []shiftAction{other, refresh}

	refreshCount := 0
	const runs = 1000
	for i := 0; i < runs; i++ {
		act := pickShiftAction(rand.New(rand.NewSource(int64(i))), actions, s)
		if act.tile == field.CoalChamber {
			refreshCount++
		}
	}
	if refreshCount < 650 {
		t.Fatalf("refresh rate %d/%d, want ~75%% with 3x weight", refreshCount, runs)
	}
}
