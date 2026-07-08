package render

import (
	"image"
	"image/color"
	"math"

	"github.com/jonas/reaktor-sim/internal/hex"
)

const (
	hexRadius        = 34.0
	demandPadLeft    = 52.0
	demandPadRight   = 56.0
	demandPadTop     = 36.0
	demandPadBottom  = 52.0
)

// layout places all hex centers on a canvas.
type layout struct {
	width, height  int
	gridHeight     int
	margin         float64
	centers        map[hex.Coord]image.Point
}

func newLayout() layout {
	ly := layout{
		margin:  50,
		centers: make(map[hex.Coord]image.Point, len(hex.AllBoardCoords)),
	}
	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64

	raw := make(map[hex.Coord][2]float64)
	for _, c := range hex.AllBoardCoords {
		x, y := c.Pixel(hexRadius)
		raw[c] = [2]float64{x, y}
		minX = math.Min(minX, x)
		minY = math.Min(minY, y)
		maxX = math.Max(maxX, x)
		maxY = math.Max(maxY, y)
	}

	shiftX := ly.margin + demandPadLeft - minX
	shiftY := ly.margin + demandPadTop - minY
	for c, xy := range raw {
		ly.centers[c] = image.Point{
			X: int(xy[0] + shiftX),
			Y: int(xy[1] + shiftY),
		}
	}
	ly.width = int(maxX-minX + 2*ly.margin + demandPadLeft + demandPadRight)
	ly.gridHeight = int(maxY-minY + 2*ly.margin + demandPadTop + demandPadBottom)
	ly.height = ly.gridHeight + legendBlockHeight()
	return ly
}

func (l layout) center(c hex.Coord) image.Point {
	return l.centers[c]
}

func hexCorners(cx, cy int, radius float64) [6]image.Point {
	var pts [6]image.Point
	for i := 0; i < 6; i++ {
		angle := -math.Pi/2 + float64(i)*math.Pi/3
		pts[i] = image.Point{
			X: cx + int(math.Cos(angle)*radius),
			Y: cy + int(-math.Sin(angle)*radius),
		}
	}
	return pts
}

func fillPolygon(img *image.RGBA, pts [6]image.Point, c color.Color) {
	bounds := img.Bounds()
	minY, maxY := bounds.Max.Y, bounds.Min.Y
	for _, p := range pts {
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	for y := minY; y <= maxY; y++ {
		if y < bounds.Min.Y || y >= bounds.Max.Y {
			continue
		}
		var crossings []int
		for i := 0; i < 6; i++ {
			a := pts[i]
			b := pts[(i+1)%6]
			if (a.Y <= y && b.Y > y) || (b.Y <= y && a.Y > y) {
				x := a.X + (y-a.Y)*(b.X-a.X)/(b.Y-a.Y)
				crossings = append(crossings, x)
			}
		}
		for i := 0; i < len(crossings); i++ {
			for j := i + 1; j < len(crossings); j++ {
				if crossings[i] > crossings[j] {
					crossings[i], crossings[j] = crossings[j], crossings[i]
				}
			}
		}
		for i := 0; i+1 < len(crossings); i += 2 {
			x0 := crossings[i]
			x1 := crossings[i+1]
			if x0 < bounds.Min.X {
				x0 = bounds.Min.X
			}
			if x1 >= bounds.Max.X {
				x1 = bounds.Max.X - 1
			}
			for x := x0; x <= x1; x++ {
				img.Set(x, y, c)
			}
		}
	}
}

func strokePolygon(img *image.RGBA, pts [6]image.Point, c color.Color) {
	for i := 0; i < 6; i++ {
		drawLine(img, pts[i], pts[(i+1)%6], c)
	}
}

func drawLine(img *image.RGBA, a, b image.Point, c color.Color) {
	dx := abs(b.X - a.X)
	dy := -abs(b.Y - a.Y)
	sx := 1
	if a.X > b.X {
		sx = -1
	}
	sy := 1
	if a.Y > b.Y {
		sy = -1
	}
	err := dx + dy
	x, y := a.X, a.Y
	for {
		if image.Pt(x, y).In(img.Bounds()) {
			img.Set(x, y, c)
		}
		if x == b.X && y == b.Y {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x += sx
		}
		if e2 <= dx {
			err += dx
			y += sy
		}
	}
}

func drawThickLine(img *image.RGBA, a, b image.Point, c color.Color, width int) {
	if width < 1 {
		width = 1
	}
	for oy := -width / 2; oy <= width/2; oy++ {
		for ox := -width / 2; ox <= width/2; ox++ {
			drawLine(img, image.Pt(a.X+ox, a.Y+oy), image.Pt(b.X+ox, b.Y+oy), c)
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

const legendLineHeight = 15

func legendBlockHeight() int {
	return 18 + len(Legend())*legendLineHeight + 2*legendLineHeight + 16
}
