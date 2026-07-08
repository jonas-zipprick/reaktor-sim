package render

import (
	"image"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/hex"
)

func TestPlantDemandRendersAboveTurbine(t *testing.T) {
	ly := newLayout()
	turbine := ly.center(hex.Coord{Q: hex.TurbineCol, R: hex.TurbineRow})
	plant := plantDemandAnchor(ly)

	if plant.Y >= turbine.Y {
		t.Fatalf("plant demand Y=%d should be above turbine Y=%d", plant.Y, turbine.Y)
	}
	if abs(plant.X-turbine.X) > 20 {
		t.Fatalf("plant demand X=%d should align with turbine X=%d", plant.X, turbine.X)
	}
}

func TestRailDemandAnchorInsideCanvas(t *testing.T) {
	ly := newLayout()
	rail := railDemandAnchor(ly)
	if rail.X <= 0 || rail.Y <= 0 || rail.X >= ly.width || rail.Y >= ly.gridHeight {
		t.Fatalf("rail demand anchor %v outside grid %dx%d", rail, ly.width, ly.gridHeight)
	}
}

func TestZeroRailDemandRendersBadge(t *testing.T) {
	state := board.NewEmpty()
	state.ApplyDemands(board.ShiftDemands{})
	ly := newLayout()
	img := image.NewRGBA(image.Rect(0, 0, ly.width, ly.gridHeight))
	drawDemandOutside(img, state, ly, 0)

	rail := railDemandAnchor(ly)
	found := false
	for dy := -12; dy <= 12; dy++ {
		for dx := -20; dx <= 20; dx++ {
			p := image.Pt(rail.X+dx, rail.Y+dy)
			if !p.In(img.Bounds()) {
				continue
			}
			r, g, b, _ := img.At(p.X, p.Y).RGBA()
			if r>>8 == 200 && g>>8 == 255 && b>>8 == 200 {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatal("expected b0 demand badge near rail anchor")
	}
}
