package hex_test

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/hex"
)

func TestVoltageReflectsAtMarkedOuterWalls(t *testing.T) {
	cases := []struct {
		pos hex.Coord
		dir hex.Rotation
	}{
		{hex.Coord{Q: 6, R: 0}, hex.RotW},
		{hex.Coord{Q: 6, R: 4}, hex.RotW},
		{hex.Coord{Q: 4, R: 1}, hex.RotNE},
		{hex.Coord{Q: 5, R: 1}, hex.RotNW},
		{hex.Coord{Q: 4, R: 3}, hex.RotSE},
		{hex.Coord{Q: 5, R: 3}, hex.RotSW},
	}
	for _, tc := range cases {
		if !hex.VoltageReflectsAtOuterWall(tc.pos, tc.dir.TravelDir()) {
			t.Fatalf("(%d,%d) dir %s should reflect", tc.pos.Q, tc.pos.R, tc.dir)
		}
	}
}

func TestVoltageDoesNotReflectDemandWalls(t *testing.T) {
	cases := []struct {
		pos hex.Coord
		dir hex.Rotation
	}{
		{hex.Coord{Q: 6, R: 0}, hex.RotNW},
		{hex.Coord{Q: 6, R: 4}, hex.RotSE},
		{hex.Coord{Q: 8, R: 2}, hex.RotE},
	}
	for _, tc := range cases {
		if hex.VoltageReflectsAtOuterWall(tc.pos, tc.dir.TravelDir()) {
			t.Fatalf("(%d,%d) dir %s should consume demand, not reflect", tc.pos.Q, tc.pos.R, tc.dir)
		}
	}
}

func TestReactorWallBlocksBetweenColumns4And5(t *testing.T) {
	wall := hex.Coord{Q: hex.ReactorWallCol, R: 1}
	target := hex.Coord{Q: hex.ReactorWallCol + 1, R: 1}
	if hex.CanEnter(wall, target) {
		t.Fatal("E/W crossing should be blocked at reactor wall")
	}
	open := hex.Coord{Q: hex.ReactorWallCol + 1, R: 2}
	if !hex.CanEnter(wall, open) {
		t.Fatal("diagonal crossing at turbine row should stay open")
	}
}
