package sim

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/graph"
	"github.com/jonas/reaktor-sim/internal/hex"
)

func TestGroundOrientationBlocksCrossAxisVoltage(t *testing.T) {
	pos := hex.Coord{Q: 6, R: 1}
	tile := field.NewTile(field.Ground, hex.RotNW, 0) // NW-SE through-axis
	chip := Chip{Type: ChipVoltage, Pos: pos, Dir: hex.RotE.TravelDir()}
	incoming := (chip.Dir + 3) % 6

	out, _ := react(board.NewEmpty(), graph.BuildFlow(board.NewEmpty(), nil), pos, &tile, chip, incoming, "", rand.New(rand.NewSource(1)))
	if len(out) != 0 {
		t.Fatalf("cross-axis voltage should be grounded, got %d chips", len(out))
	}
}

func TestGroundOrientationPassesAxisVoltage(t *testing.T) {
	pos := hex.Coord{Q: 6, R: 1}
	tile := field.NewTile(field.Ground, hex.RotW, 0) // E-W through-axis
	chip := Chip{Type: ChipVoltage, Pos: pos, Dir: hex.RotE.TravelDir()}
	incoming := (chip.Dir + 3) % 6

	out, _ := react(board.NewEmpty(), graph.BuildFlow(board.NewEmpty(), nil), pos, &tile, chip, incoming, "", rand.New(rand.NewSource(1)))
	if len(out) != 1 {
		t.Fatalf("axis-aligned voltage should pass through, got %d chips", len(out))
	}
	if out[0].Dir != hex.RotE.TravelDir() {
		t.Fatalf("pass-through dir = %s, want E", hex.DisplayDirName(out[0].Dir))
	}
}

func TestGroundHeatAlwaysPasses(t *testing.T) {
	pos := hex.Coord{Q: 6, R: 1}
	tile := field.NewTile(field.Ground, hex.RotNW, 0)
	chip := Chip{Type: ChipHeat, Pos: pos, Dir: hex.RotE.TravelDir()}
	incoming := (chip.Dir + 3) % 6

	out, _ := react(board.NewEmpty(), graph.BuildFlow(board.NewEmpty(), nil), pos, &tile, chip, incoming, "", rand.New(rand.NewSource(1)))
	if len(out) != 1 || out[0].Type != ChipHeat {
		t.Fatalf("heat should pass ground unchanged, got %+v", out)
	}
}
