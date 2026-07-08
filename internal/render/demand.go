package render

import (
	"fmt"
	"image"
	"image/color"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/hex"
)

func demandOffset(f float64) int {
	return int(hexRadius * f)
}

type demandSide struct {
	zone   board.Zone
	anchor func(layout) image.Point
}

var demandSides = []demandSide{
	{board.ZoneIndustry, industryDemandAnchor},
	{board.ZoneResidential, residentialDemandAnchor},
	{board.ZoneRail, railDemandAnchor},
	{board.ZonePlant, plantDemandAnchor},
}

// drawDemandOutside renders remaining shift demand totals beside the hex grid.
func drawDemandOutside(img *image.RGBA, state *board.State, ly layout, yOffset int) {
	for _, side := range demandSides {
		pos := side.anchor(ly)
		pos.Y += yOffset
		drawDemandBadge(img, pos, board.ZoneLetter(side.zone), state.TotalDemand(side.zone))
	}
}

func drawDemandBadge(img *image.RGBA, center image.Point, letter string, count int) {
	label := fmt.Sprintf("%s%d", letter, count)
	w := len(ASCII(label))*7 + 8
	h := 18
	x0 := center.X - w/2
	y0 := center.Y - h/2
	for y := y0; y < y0+h; y++ {
		for x := x0; x < x0+w; x++ {
			if image.Pt(x, y).In(img.Bounds()) {
				img.Set(x, y, colDemand)
			}
		}
	}
	strokeRect(img, image.Rect(x0, y0, x0+w, y0+h), colBorder)
	drawCellLabels(img, center, []string{label}, colText)
}

func strokeRect(img *image.RGBA, r image.Rectangle, c color.Color) {
	drawLine(img, image.Pt(r.Min.X, r.Min.Y), image.Pt(r.Max.X-1, r.Min.Y), c)
	drawLine(img, image.Pt(r.Max.X-1, r.Min.Y), image.Pt(r.Max.X-1, r.Max.Y-1), c)
	drawLine(img, image.Pt(r.Max.X-1, r.Max.Y-1), image.Pt(r.Min.X, r.Max.Y-1), c)
	drawLine(img, image.Pt(r.Min.X, r.Max.Y-1), image.Pt(r.Min.X, r.Min.Y), c)
}

func industryDemandAnchor(ly layout) image.Point {
	return anchorAbove(ly, []hex.Coord{{Q: 5, R: 0}, {Q: 6, R: 0}, {Q: 7, R: 0}})
}

func residentialDemandAnchor(ly layout) image.Point {
	cells := []hex.Coord{{Q: 8, R: 0}, {Q: 8, R: 1}, {Q: 8, R: 2}}
	var sumY, maxX int
	n := 0
	for _, c := range cells {
		if !c.Valid() {
			continue
		}
		p := ly.center(c)
		sumY += p.Y
		if p.X > maxX {
			maxX = p.X
		}
		n++
	}
	if n == 0 {
		return image.Point{}
	}
	return image.Pt(maxX+demandOffset(1.25), sumY/n)
}

func railDemandAnchor(ly layout) image.Point {
	return anchorBelow(ly, []hex.Coord{{Q: 5, R: 2}, {Q: 6, R: 2}, {Q: 7, R: 2}})
}

func plantDemandAnchor(ly layout) image.Point {
	return anchorAbove(ly, []hex.Coord{{Q: hex.TurbineCol, R: hex.TurbineRow}})
}

func anchorAbove(ly layout, cells []hex.Coord) image.Point {
	var sumX, minY int
	n := 0
	for _, c := range cells {
		if !c.Valid() {
			continue
		}
		p := ly.center(c)
		sumX += p.X
		if n == 0 || p.Y < minY {
			minY = p.Y
		}
		n++
	}
	if n == 0 {
		return image.Point{}
	}
	return image.Pt(sumX/n, minY-demandOffset(1.15))
}

func anchorBelow(ly layout, cells []hex.Coord) image.Point {
	var sumX, maxY int
	n := 0
	for _, c := range cells {
		if !c.Valid() {
			continue
		}
		p := ly.center(c)
		sumX += p.X
		if n == 0 || p.Y > maxY {
			maxY = p.Y
		}
		n++
	}
	if n == 0 {
		return image.Point{}
	}
	return image.Pt(sumX/n, maxY+demandOffset(1.15))
}

// DemandSummaryLines returns remaining demand per zone for text output.
func DemandSummaryLines(state *board.State) []string {
	lines := make([]string, 0, 4)
	for _, z := range []board.Zone{
		board.ZoneIndustry,
		board.ZoneResidential,
		board.ZoneRail,
		board.ZonePlant,
	} {
		lines = append(lines, fmt.Sprintf("%s %s: %d", board.ZoneLetter(z), z.String(), state.TotalDemand(z)))
	}
	return lines
}
