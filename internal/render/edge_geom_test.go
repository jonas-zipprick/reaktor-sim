package render

import (
	"math"
	"testing"

	"github.com/jonas/reaktor-sim/internal/hex"
)

func edgeMidpoint(cx, cy int, radius float64, edge int) (float64, float64) {
	corners := hexCorners(cx, cy, radius)
	a := corners[edge]
	b := corners[(edge+1)%6]
	return float64(a.X+b.X) / 2, float64(a.Y+b.Y) / 2
}

func bestEdgeForNeighbor(t *testing.T, c hex.Coord, dir int) int {
	t.Helper()
	const radius = hexRadius
	cx, cy := c.Pixel(radius)
	cxI, cyI := int(cx), int(cy)
	n := c.Neighbor(dir)
	nx, ny := n.Pixel(radius)
	dx, dy := nx-cx, ny-cy
	dLen := math.Hypot(dx, dy)
	if dLen == 0 {
		t.Fatal("zero neighbor offset")
	}

	best, bestDot := -1, -1e9
	for i := 0; i < 6; i++ {
		mx, my := edgeMidpoint(cxI, cyI, radius, i)
		ex, ey := mx-cx, my-cy
		eLen := math.Hypot(ex, ey)
		if eLen == 0 {
			continue
		}
		dot := (ex/eLen)*(dx/dLen) + (ey/eLen)*(dy/dLen)
		if dot > bestDot {
			bestDot = dot
			best = i
		}
	}
	return best
}

func TestTravelDirEdgeIndexMatchesGeometry(t *testing.T) {
	c := hex.Coord{Q: 5, R: 1}
	for _, rot := range []hex.Rotation{hex.RotE, hex.RotNE, hex.RotNW, hex.RotW, hex.RotSW, hex.RotSE} {
		dir := rot.TravelDir()
		want := bestEdgeForNeighbor(t, c, dir)
		got := travelDirEdgeIndex(dir)
		if got != want {
			t.Fatalf("%s from (%d,%d): mapped edge %d, geometry edge %d", rot, c.Q, c.R, got, want)
		}
	}
}
