package render

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

var (
	colReactor      = color.RGBA{R: 255, G: 235, B: 210, A: 255}
	colGrid         = color.RGBA{R: 210, G: 230, B: 255, A: 255}
	colNeutral      = color.RGBA{R: 235, G: 235, B: 235, A: 255}
	colBorder       = color.RGBA{R: 60, G: 60, B: 60, A: 255}
	colSpecial      = color.RGBA{R: 255, G: 220, B: 120, A: 255}
	colBurned       = color.RGBA{R: 180, G: 180, B: 180, A: 255}
	colDemand       = color.RGBA{R: 200, G: 255, B: 200, A: 255}
	colText         = color.RGBA{R: 20, G: 20, B: 20, A: 255}
	colEmptyOutline = color.RGBA{R: 220, G: 220, B: 225, A: 255}
)

// WriteBoardPNG saves a simple symbolic board rendering.
func WriteBoardPNG(state *board.State, path string, view ChipView) error {
	ly := newLayout()
	img := image.NewRGBA(image.Rect(0, 0, ly.width, ly.height))
	for y := 0; y < ly.height; y++ {
		for x := 0; x < ly.width; x++ {
			img.Set(x, y, color.White)
		}
	}

	for _, c := range hex.AllBoardCoords {
		tile := state.Tiles[c.Q][c.R]
		center := ly.center(c)
		corners := hexCorners(center.X, center.Y, hexRadius-1)
		if !CellHasContent(c, tile, view) {
			strokePolygon(img, corners, colEmptyOutline)
			continue
		}
		fill := cellFill(c, tile)
		fillPolygon(img, corners, fill)
		strokePolygon(img, corners, colBorder)
		drawCellLabels(img, center, cellLabelLines(state, c, tile, view), colText)
	}

	drawDemandOutside(img, state, ly, 0)

	drawLabelLeft(img, image.Pt(10, ly.gridHeight+10), "Reaktor (Spalten 1-5) | Turbine Tu | Stromnetz (Spalten 6-9)", colText)
	for i, line := range Legend() {
		drawLabelLeft(img, image.Pt(10, ly.gridHeight+28+i*legendLineHeight), line, colText)
	}
	zoneY := ly.gridHeight + 28 + len(Legend())*legendLineHeight
	drawLabelLeft(img, image.Pt(10, zoneY), "Rand-Bedarf ausserhalb: I oben  W rechts  b unten  R oben (Turbine)", colText)

	return writePNG(path, img)
}

func cellLabelLines(state *board.State, c hex.Coord, tile field.Tile, view ChipView) []string {
	lines := []string{Label(state, c)}
	if c.IsEmitter() && state.EmitterDamage > 0 {
		lines = append(lines, fmt.Sprintf("!%d", state.EmitterDamage))
	}
	if bound := BottomLabel(state, c, tile); bound != "" {
		lines = append(lines, bound)
	}
	if loose := LooseLabel(LooseCountsAt(view, c)); loose != "" {
		lines = append(lines, loose)
	}
	return lines
}

func cellFillFor(state *board.State, c hex.Coord, tile field.Tile) color.RGBA {
	return cellFill(c, tile)
}

func cellFill(c hex.Coord, tile field.Tile) color.RGBA {
	if c.IsEmitter() || c.IsTurbine() {
		return colSpecial
	}
	if tile.BurnedOut {
		return colBurned
	}
	if c.IsPlayer1() {
		return colReactor
	}
	if c.IsPlayer2() {
		return colGrid
	}
	return colNeutral
}

func drawLabel(img *image.RGBA, center image.Point, text string, col color.Color) {
	drawLabelStacked(img, center, text, "", col)
}

func drawLabelStacked(img *image.RGBA, center image.Point, top, bottom string, col color.Color) {
	drawCellLabels(img, center, []string{top, bottom}, col)
}

func drawCellLabels(img *image.RGBA, center image.Point, lines []string, col color.Color) {
	face := basicfont.Face7x13
	nonEmpty := make([]string, 0, len(lines))
	for _, line := range lines {
		if line != "" {
			nonEmpty = append(nonEmpty, line)
		}
	}
	if len(nonEmpty) == 0 {
		return
	}
	if len(nonEmpty) == 1 {
		width := len(ASCII(nonEmpty[0])) * 7
		d := &font.Drawer{
			Dst:  img,
			Src:  image.NewUniform(col),
			Face: face,
			Dot:  fixed.P(center.X-width/2, center.Y+4),
		}
		d.DrawString(ASCII(nonEmpty[0]))
		return
	}
	lineHeight := 11
	totalH := (len(nonEmpty) - 1) * lineHeight
	startY := center.Y - totalH/2 + 4
	for i, line := range nonEmpty {
		width := len(ASCII(line)) * 7
		d := &font.Drawer{
			Dst:  img,
			Src:  image.NewUniform(col),
			Face: face,
			Dot:  fixed.P(center.X-width/2, startY+i*lineHeight),
		}
		d.DrawString(ASCII(line))
	}
}

func drawLabelLeft(img *image.RGBA, topLeft image.Point, text string, col color.Color) {
	text = ASCII(text)
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  fixed.P(topLeft.X, topLeft.Y+13),
	}
	d.DrawString(text)
}

func captionLines(caption string) []string {
	return WrapCaption(caption, 800)
}

func captionBlockHeight(lines []string) int {
	if len(lines) == 0 {
		return 0
	}
	return len(lines)*14 + 4
}

func drawCaptionLines(img *image.RGBA, topLeft image.Point, lines []string, col color.Color) {
	for i, line := range lines {
		drawLabelLeft(img, image.Pt(topLeft.X, topLeft.Y+i*14), line, col)
	}
}

func writePNG(path string, img image.Image) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
