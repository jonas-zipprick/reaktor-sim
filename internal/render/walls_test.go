package render

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/hex"
)

func hasWall(edges []wallEdge, c hex.Coord, dir int, kind wallKind) bool {
	for _, e := range edges {
		if e.coord == c && e.dir == dir && e.kind == kind {
			return true
		}
	}
	return false
}

func TestBoardWallEdgesPlantWallBesideTurbine(t *testing.T) {
	edges := boardWallEdges()
	for _, hit := range board.PlantWallHits() {
		if !hasWall(edges, hit.From, hit.Dir, wallDemand) {
			t.Fatalf("(%d,%d) dir %d missing plant wall", hit.From.Q, hit.From.R, hit.Dir)
		}
	}
}

func TestBoardWallEdgesReactorWallNotchHeatReflect(t *testing.T) {
	edges := boardWallEdges()
	for _, tc := range []struct {
		c   hex.Coord
		dir hex.Rotation
	}{
		{hex.Coord{Q: 3, R: 1}, hex.RotNE},
		{hex.Coord{Q: 3, R: 3}, hex.RotSE},
	} {
		if !hasWall(edges, tc.c, tc.dir.TravelDir(), wallP1HeatReflect) {
			t.Fatalf("(%d,%d) %s missing heat reflect wall", tc.c.Q, tc.c.R, tc.dir)
		}
	}
}

func TestBoardWallEdgesReactorWall(t *testing.T) {
	edges := boardWallEdges()
	for _, r := range []int{1, 3} {
		c := hex.Coord{Q: hex.ReactorWallCol, R: r}
		if !hasWall(edges, c, hex.RotE.TravelDir(), wallInternal) {
			t.Fatalf("(%d,%d) missing reactor wall on east edge", c.Q, c.R)
		}
	}
}

func TestBoardWallEdgesVoltageReflect(t *testing.T) {
	edges := boardWallEdges()
	cases := []struct {
		c   hex.Coord
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
		if !hasWall(edges, tc.c, tc.dir.TravelDir(), wallReflectVoltage) {
			t.Fatalf("(%d,%d) dir %s missing voltage reflect wall", tc.c.Q, tc.c.R, tc.dir)
		}
	}
}

func TestBoardWallEdgesDemandZones(t *testing.T) {
	edges := boardWallEdges()
	if !hasWall(edges, hex.Coord{Q: 8, R: 2}, hex.RotE.TravelDir(), wallDemand) {
		t.Fatal("missing residential demand on east edge of column 9")
	}
	if !hasWall(edges, hex.Coord{Q: hex.TurbineCol, R: hex.TurbineRow}, hex.RotW.TravelDir(), wallDemand) {
		t.Fatal("missing plant demand marker on turbine west edge")
	}
}

func TestBoardWallEdgesPlantZoneLetter(t *testing.T) {
	for _, e := range boardWallEdges() {
		if e.kind == wallDemand && e.zone == board.ZonePlant && e.coord.IsTurbine() {
			return
		}
	}
	t.Fatal("turbine plant wall edge not found")
}

func TestTravelDirEdgeIndexMatchesRotations(t *testing.T) {
	cases := []struct {
		rot  hex.Rotation
		edge int
	}{
		{hex.RotE, 1},
		{hex.RotNE, 2},
		{hex.RotNW, 3},
		{hex.RotW, 4},
		{hex.RotSW, 5},
		{hex.RotSE, 0},
	}
	for _, tc := range cases {
		if got := travelDirEdgeIndex(tc.rot.TravelDir()); got != tc.edge {
			t.Fatalf("%s travel dir edge = %d, want %d", tc.rot, got, tc.edge)
		}
	}
}
