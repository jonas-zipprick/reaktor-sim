package hex_test

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/hex"
)

func TestBlockedBoundaryDiagonalPastReactorWallIsOuter(t *testing.T) {
	cases := []struct {
		c   hex.Coord
		dir hex.Rotation
	}{
		{hex.Coord{Q: 3, R: 1}, hex.RotNE},
		{hex.Coord{Q: 3, R: 3}, hex.RotSE},
	}
	for _, tc := range cases {
		next := tc.c.Neighbor(tc.dir.TravelDir())
		if got := hex.BlockedBoundary(tc.c, next, tc.dir.TravelDir()); got != hex.BoundaryOuter {
			t.Fatalf("(%d,%d) %s boundary = %v, want outer", tc.c.Q, tc.c.R, tc.dir, got)
		}
		if !hex.HeatReflectsAtOuterWall(tc.c, tc.dir.TravelDir()) {
			t.Fatalf("(%d,%d) %s should reflect heat", tc.c.Q, tc.c.R, tc.dir)
		}
	}
}

func TestBlockedBoundaryStraightReactorWallStaysInternal(t *testing.T) {
	from := hex.Coord{Q: 3, R: 1}
	to := hex.Coord{Q: 4, R: 1}
	if got := hex.BlockedBoundary(from, to, hex.RotE.TravelDir()); got != hex.BoundaryInternalWall {
		t.Fatalf("east into reactor wall = %v, want internal", got)
	}
}

func TestReflectOffOuterWallReversesTravelDir(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"NW", "SE"},
		{"SE", "NW"},
		{"NE", "SW"},
		{"SW", "NE"},
		{"E", "W"},
		{"W", "E"},
	}
	for _, tc := range cases {
		var inDir int
		for d := 0; d < 6; d++ {
			if hex.DisplayDirName(d) == tc.in {
				inDir = d
				break
			}
		}
		got := hex.DisplayDirName(hex.ReflectOffOuterWall(inDir))
		if got != tc.want {
			t.Fatalf("reflect %s: got %s, want %s", tc.in, got, tc.want)
		}
	}
}
