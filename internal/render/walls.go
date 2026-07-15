package render

import (
	"image"
	"image/color"
	"math"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/hex"
)

var (
	colWallInternal = color.RGBA{R: 160, G: 75, B: 30, A: 255}
	colWallReflect  = color.RGBA{R: 35, G: 110, B: 220, A: 255}
	colWallP1Outer  = color.RGBA{R: 230, G: 150, B: 60, A: 255}
)

func demandWallColor(z board.Zone) color.RGBA {
	switch z {
	case board.ZoneIndustry:
		return color.RGBA{R: 70, G: 170, B: 70, A: 255}
	case board.ZoneResidential:
		return color.RGBA{R: 50, G: 150, B: 180, A: 255}
	case board.ZoneRail:
		return color.RGBA{R: 130, G: 90, B: 190, A: 255}
	case board.ZonePlant:
		return color.RGBA{R: 210, G: 120, B: 40, A: 255}
	default:
		return colDemand
	}
}

type wallKind int

const (
	wallInternal wallKind = iota
	wallReflectVoltage
	wallDemand
	wallP1HeatReflect
)

type wallEdge struct {
	coord hex.Coord
	dir   int
	kind  wallKind
	zone  board.Zone // for wallDemand
}

// travelDirEdgeIndex maps a chip travel direction to the hex edge between corners i and i+1.
// Derived from pointy-top odd-r layout in geo.hexCorners (y grows downward).
func travelDirEdgeIndex(dir int) int {
	switch dir % 6 {
	case hex.RotE.TravelDir():
		return 1
	case hex.RotNE.TravelDir():
		return 2
	case hex.RotNW.TravelDir():
		return 3
	case hex.RotW.TravelDir():
		return 4
	case hex.RotSW.TravelDir():
		return 5
	case hex.RotSE.TravelDir():
		return 0
	default:
		return 0
	}
}

func boardWallEdges() []wallEdge {
	var edges []wallEdge
	seen := make(map[[3]int]bool, 32)

	add := func(c hex.Coord, dir int, kind wallKind, zone board.Zone) {
		key := [3]int{c.Q, c.R, dir}
		if seen[key] {
			return
		}
		seen[key] = true
		edges = append(edges, wallEdge{coord: c, dir: dir, kind: kind, zone: zone})
	}

	for _, c := range hex.AllBoardCoords {
		if c.HasWallRight() {
			add(c, hex.RotE.TravelDir(), wallInternal, 0)
		}
		if c.WallBlocksWest() {
			add(c, hex.RotW.TravelDir(), wallInternal, 0)
		}

		for dir := 0; dir < 6; dir++ {
			if hex.VoltageReflectsAtOuterWall(c, dir) {
				add(c, dir, wallReflectVoltage, 0)
			}
		}

		for dir := 0; dir < 6; dir++ {
			if hex.HeatReflectsAtOuterWall(c, dir) {
				add(c, dir, wallP1HeatReflect, 0)
			}
		}

		for _, hit := range board.PlantWallHits() {
			add(hit.From, hit.Dir, wallDemand, board.ZonePlant)
		}

		if c.IsPlayer2() {
			for dir := 0; dir < 6; dir++ {
				if hex.VoltageReflectsAtOuterWall(c, dir) {
					continue
				}
				next := c.Neighbor(dir)
				if hex.BlockedBoundary(c, next, dir) != hex.BoundaryOuter {
					continue
				}
				for _, z := range board.ZonesForOuterWallHit(c, dir) {
					add(c, dir, wallDemand, z)
				}
			}
		}
	}

	return edges
}

func drawBoardWalls(img *image.RGBA, ly layout, yOffset int) {
	for _, w := range boardWallEdges() {
		center := ly.center(w.coord)
		center.Y += yOffset
		corners := hexCorners(center.X, center.Y, hexRadius-1)
		drawWallEdge(img, center, corners, w)
	}
}

func drawWallEdge(img *image.RGBA, center image.Point, corners [6]image.Point, w wallEdge) {
	ei := travelDirEdgeIndex(w.dir)
	a := corners[ei]
	b := corners[(ei+1)%6]

	switch w.kind {
	case wallInternal:
		drawThickLine(img, a, b, colWallInternal, 5)
	case wallReflectVoltage:
		drawThickLine(img, a, b, colWallReflect, 4)
		drawInsetParallel(img, center, a, b, colWallReflect, 2, 6)
	case wallP1HeatReflect:
		drawThickLine(img, a, b, colWallP1Outer, 2)
	case wallDemand:
		col := demandWallColor(w.zone)
		drawThickLine(img, a, b, col, 3)
		mid := image.Pt((a.X+b.X)/2, (a.Y+b.Y)/2)
		in := edgeInsetPoint(center, mid, 10)
		drawEdgeLetter(img, in, board.ZoneLetter(w.zone), colText)
	}
}

func edgeInsetPoint(center, edgeMid image.Point, inset int) image.Point {
	dx := float64(center.X - edgeMid.X)
	dy := float64(center.Y - edgeMid.Y)
	length := math.Hypot(dx, dy)
	if length == 0 {
		return edgeMid
	}
	scale := float64(inset) / length
	return image.Pt(
		edgeMid.X+int(dx*scale),
		edgeMid.Y+int(dy*scale),
	)
}

func drawInsetParallel(img *image.RGBA, center, a, b image.Point, col color.Color, width, inset int) {
	mid := image.Pt((a.X+b.X)/2, (a.Y+b.Y)/2)
	dx := float64(mid.X - center.X)
	dy := float64(mid.Y - center.Y)
	length := math.Hypot(dx, dy)
	if length == 0 {
		return
	}
	n := float64(inset) / length
	ox, oy := int(dx*n), int(dy*n)
	drawThickLine(img, image.Pt(a.X+ox, a.Y+oy), image.Pt(b.X+ox, b.Y+oy), col, width)
}

func drawEdgeLetter(img *image.RGBA, at image.Point, letter string, col color.Color) {
	drawCellLabels(img, at, []string{letter}, col)
}

// WallLegendLines documents the edge colors on board PNGs.
func WallLegendLines() []string {
	return []string{
		"Wand-Kanten: braun = Reaktorwand (Spalte 4)  blau = Spannungs-Reflektion",
		"gruen/blau/violett/orange + Buchstabe = Rand-Bedarf (I/W/b/R)  orange duenn = P1-Waermereflektion",
	}
}
