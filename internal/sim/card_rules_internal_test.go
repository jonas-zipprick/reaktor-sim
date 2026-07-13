package sim

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/graph"
	"github.com/jonas/reaktor-sim/internal/hex"
)

func TestReactTestlaufDisablesCoolingTower(t *testing.T) {
	pos := hex.Coord{Q: 2, R: 1}
	tile := field.NewTile(field.CoolingTower, 0, 0)
	chip := Chip{Type: ChipHeat, Pos: pos, Dir: 0}
	b := board.NewEmpty()
	g := graph.BuildFlow(b, nil)

	normal, _ := react(b, g, pos, &tile, chip, 0, "", rand.New(rand.NewSource(1)))
	if len(normal) != 0 {
		t.Fatalf("normal cooling should destroy heat, got %d chips", len(normal))
	}

	testlauf, _ := react(b, g, pos, &tile, chip, 0, "testlauf-volllast", rand.New(rand.NewSource(1)))
	if len(testlauf) != 1 {
		t.Fatalf("testlauf cooling disabled, want pass-through, got %d chips", len(testlauf))
	}
}
