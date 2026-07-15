package sim

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/graph"
	"github.com/jonas/reaktor-sim/internal/hex"
)

func TestCoolingTowerOrientationBlocksCrossAxisHeat(t *testing.T) {
	pos := hex.Coord{Q: 2, R: 1}
	tile := field.NewTile(field.CoolingTower, hex.RotNW, 0) // NW-SE through-axis
	chip := Chip{Type: ChipHeat, Pos: pos, Dir: hex.RotE.TravelDir()}
	incoming := (chip.Dir + 3) % 6

	out, _ := react(board.NewEmpty(), graph.BuildFlow(board.NewEmpty(), nil), pos, &tile, chip, incoming, "", rand.New(rand.NewSource(1)))
	if len(out) != 0 {
		t.Fatalf("cross-axis heat should be absorbed, got %d chips", len(out))
	}
}

func TestCoolingTowerOrientationPassesAxisHeat(t *testing.T) {
	pos := hex.Coord{Q: 2, R: 1}
	tile := field.NewTile(field.CoolingTower, hex.RotNW, 0) // NW-SE through-axis
	chip := Chip{Type: ChipHeat, Pos: pos, Dir: hex.RotNW.TravelDir()}
	incoming := (chip.Dir + 3) % 6

	out, _ := react(board.NewEmpty(), graph.BuildFlow(board.NewEmpty(), nil), pos, &tile, chip, incoming, "", rand.New(rand.NewSource(1)))
	if len(out) != 1 {
		t.Fatalf("axis-aligned heat should pass through, got %d chips", len(out))
	}
	if out[0].Dir != hex.RotNW.TravelDir() {
		t.Fatalf("pass-through dir = %s, want NW", hex.DisplayDirName(out[0].Dir))
	}
}

func TestCoolingTowerOrientationChangesPassDirection(t *testing.T) {
	pos := hex.Coord{Q: 2, R: 1}
	tile := field.NewTile(field.CoolingTower, hex.RotW, 0) // E-W through-axis
	chip := Chip{Type: ChipHeat, Pos: pos, Dir: hex.RotNW.TravelDir()}
	incoming := (chip.Dir + 3) % 6

	out, _ := react(board.NewEmpty(), graph.BuildFlow(board.NewEmpty(), nil), pos, &tile, chip, incoming, "", rand.New(rand.NewSource(1)))
	if len(out) != 0 {
		t.Fatalf("NW approach should be blocked with W orientation, got %d chips", len(out))
	}

	chip = Chip{Type: ChipHeat, Pos: pos, Dir: hex.RotE.TravelDir()}
	incoming = (chip.Dir + 3) % 6
	out, _ = react(board.NewEmpty(), graph.BuildFlow(board.NewEmpty(), nil), pos, &tile, chip, incoming, "", rand.New(rand.NewSource(1)))
	if len(out) != 1 {
		t.Fatalf("E approach should pass with W orientation, got %d chips", len(out))
	}
}

func TestCoolingTowerNeutronsAlwaysPass(t *testing.T) {
	pos := hex.Coord{Q: 2, R: 1}
	tile := field.NewTile(field.CoolingTower, hex.RotNW, 0)
	chip := Chip{Type: ChipNeutron, Pos: pos, Dir: hex.RotE.TravelDir()}
	incoming := (chip.Dir + 3) % 6

	out, _ := react(board.NewEmpty(), graph.BuildFlow(board.NewEmpty(), nil), pos, &tile, chip, incoming, "", rand.New(rand.NewSource(1)))
	if len(out) != 1 || out[0].Type != ChipNeutron {
		t.Fatalf("neutrons should pass cooling tower unchanged, got %+v", out)
	}
}
