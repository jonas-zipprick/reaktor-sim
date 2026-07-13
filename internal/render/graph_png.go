package render

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/graph"
	"github.com/jonas/reaktor-sim/internal/hex"
)

var (
	colHeat    = color.RGBA{R: 220, G: 60, B: 40, A: 220}
	colNeutron = color.RGBA{R: 40, G: 160, B: 60, A: 220}
	colVoltage = color.RGBA{R: 40, G: 90, B: 220, A: 220}
)

const minEdgeProb = 0.02

// WriteGraphPNG renders nodes and colored edges for heat/neutron/voltage flow.
func WriteGraphPNG(state *board.State, g *graph.Graph, path string, caption string, view ChipView) error {
	ly := newLayout()
	const legend1 = "Kanten: rot=Waerme  gruen=Neutron  blau=Spannung"
	const legend2 = "(Linienstaerke ~ Wahrscheinlichkeit)"
	const legend3 = "Ladung gebunden: n/max  *=unendlich  (+nW/+nN/+nS ungebunden  >nW/>nN/>nS einkommend)"
	const legend4 = "Reaktive Felder mit Flug-Chips: alle 6 Wuerfelkanten (je 1/6)"

	minWidth := ly.width
	legendLines := []string{
		legend1, legend2, legend3, legend4,
		"Rand-Bedarf ausserhalb des Feldes: I/W/b/R + Zahl (!n = Schaden)",
		"Zuender-Schaden links am Raster (!n)",
		fmt.Sprintf("%d Knoten, Kanten mit P >= %.0f%%", len(g.Nodes), minEdgeProb*100),
	}
	for _, leg := range legendLines {
		if w := len(leg)*labelCharWidth + 20; w > minWidth {
			minWidth = w
		}
	}
	captionLines := WrapCaption(caption, minWidth-20)
	if w := captionTextWidth(captionLines) + 20; w > minWidth {
		minWidth = w
		captionLines = WrapCaption(caption, minWidth-20)
	}
	width := minWidth
	captionHeight := captionBlockHeight(captionLines)
	height := ly.height + 2*legendLineHeight + captionHeight

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			img.Set(x, y, color.RGBA{R: 250, G: 250, B: 252, A: 255})
		}
	}

	captionOffset := 0
	if len(captionLines) > 0 {
		drawCaptionLines(img, image.Pt(10, 8), captionLines, colText)
		captionOffset = captionHeight
	}

	// Faint hex grid for orientation (including empty slots).
	for _, c := range hex.AllBoardCoords {
		center := ly.center(c)
		center.Y += captionOffset
		corners := hexCorners(center.X, center.Y, hexRadius-2)
		strokePolygon(img, corners, colEmptyOutline)
	}

	// Edges first, then nodes on top.
	for _, node := range g.Nodes {
		from := ly.center(node.Coord)
		from.Y += captionOffset
		for _, e := range node.Edges {
			to := ly.center(e.To)
			to.Y += captionOffset
			if e.Heat >= minEdgeProb {
				drawThickLine(img, offsetPoint(from, to, 0), offsetPoint(to, from, 0), colHeat, edgeWidth(e.Heat))
			}
			if e.Neutron >= minEdgeProb {
				drawThickLine(img, offsetPoint(from, to, 1), offsetPoint(to, from, 1), colNeutron, edgeWidth(e.Neutron))
			}
			if e.Voltage >= minEdgeProb {
				drawThickLine(img, offsetPoint(from, to, 2), offsetPoint(to, from, 2), colVoltage, edgeWidth(e.Voltage))
			}
		}
	}

	for _, node := range g.Nodes {
		tile := state.Tiles[node.Coord.Q][node.Coord.R]
		if !CellHasContent(node.Coord, tile, view) {
			continue
		}
		center := ly.center(node.Coord)
		center.Y += captionOffset
		fill := cellFillFor(state, node.Coord, tile)
		corners := hexCorners(center.X, center.Y, hexRadius-4)
		fillPolygon(img, corners, fill)
		strokePolygon(img, corners, colBorder)
		drawCellLabels(img, center, cellLabelLines(state, node.Coord, tile, view), colText)
	}

	drawDemandOutside(img, state, ly, captionOffset)
	drawEmitterDamageOutside(img, state, ly, captionOffset)

	legendY := ly.gridHeight + 10 + captionOffset
	drawLabelLeft(img, image.Pt(10, legendY), legend1, colText)
	drawLabelLeft(img, image.Pt(10, legendY+legendLineHeight), legend2, colText)
	drawLabelLeft(img, image.Pt(10, legendY+2*legendLineHeight), legend3, colText)
	drawLabelLeft(img, image.Pt(10, legendY+3*legendLineHeight), legend4, colText)
	drawLabelLeft(img, image.Pt(10, legendY+4*legendLineHeight), "Rand-Bedarf ausserhalb des Feldes: I/W/b/R + Zahl", colText)
	drawLabelLeft(img, image.Pt(10, legendY+5*legendLineHeight), fmt.Sprintf("%d Knoten, Kanten mit P >= %.0f%%", len(g.Nodes), minEdgeProb*100), colText)

	return writePNG(path, img)
}

func edgeWidth(p float64) int {
	w := int(1 + p*5)
	if w < 1 {
		return 1
	}
	if w > 6 {
		return 6
	}
	return w
}

// offsetPoint shifts line endpoints perpendicular to the connection for parallel edge lanes.
func offsetPoint(from, to image.Point, lane int) image.Point {
	dx := float64(to.X - from.X)
	dy := float64(to.Y - from.Y)
	length := math.Hypot(dx, dy)
	if length == 0 {
		return from
	}
	nx := -dy / length
	ny := dx / length
	shift := float64(lane-1) * 2.5
	return image.Pt(
		from.X+int(nx*shift),
		from.Y+int(ny*shift),
	)
}
