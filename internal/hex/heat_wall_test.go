package hex

import "testing"

func TestHeatReflectsAtEveryPlayer1OuterEdge(t *testing.T) {
	for _, c := range AllBoardCoords {
		if !c.IsPlayer1() {
			continue
		}
		for dir := 0; dir < 6; dir++ {
			next := c.Neighbor(dir)
			if BlockedBoundary(c, next, dir) != BoundaryOuter {
				continue
			}
			if c.Q == TurbineCol && (c.R == TurbineRow-1 || c.R == TurbineRow+1) {
				continue
			}
			if VoltageReflectsAtOuterWall(c, dir) {
				continue
			}
			if !HeatReflectsAtOuterWall(c, dir) {
				t.Fatalf("(%d,%d) dir %s should reflect heat on P1 outer wall",
					c.Q, c.R, DisplayDirName(dir))
			}
		}
	}
}

func TestHeatReflectsAtIgniterIndent(t *testing.T) {
	cases := []struct {
		c   Coord
		dir Rotation
	}{
		{Coord{Q: 1, R: 1}, RotNW},
		{Coord{Q: 1, R: 3}, RotSW},
		{Coord{Q: 0, R: 1}, RotNW},
		{Coord{Q: 0, R: 3}, RotSW},
	}
	for _, tc := range cases {
		if !HeatReflectsAtOuterWall(tc.c, tc.dir.TravelDir()) {
			t.Fatalf("(%d,%d) %s should reflect heat on P1 field border", tc.c.Q, tc.c.R, tc.dir)
		}
	}
}

func TestHeatDoesNotReflectAtTurbineInterfaceSlots(t *testing.T) {
	for _, tc := range []struct {
		c   Coord
		dir Rotation
	}{
		{Coord{Q: 4, R: 1}, RotNE},
		{Coord{Q: 4, R: 1}, RotNW},
		{Coord{Q: 4, R: 3}, RotSE},
		{Coord{Q: 4, R: 3}, RotSW},
	} {
		if HeatReflectsAtOuterWall(tc.c, tc.dir.TravelDir()) {
			t.Fatalf("(%d,%d) %s is beside turbine, heat must not reflect", tc.c.Q, tc.c.R, tc.dir)
		}
	}
}

func TestHeatDoesNotReflectOnPlayer2OuterEdge(t *testing.T) {
	c := Coord{Q: 6, R: 0}
	if HeatReflectsAtOuterWall(c, RotW.TravelDir()) {
		t.Fatal("P2 outer wall must not reflect heat")
	}
}
