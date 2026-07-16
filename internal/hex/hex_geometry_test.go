package hex_test

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/hex"
)

func TestBoardGeometryMatchesGameRules(t *testing.T) {
	if hex.Rows != 5 || hex.Cols != 9 {
		t.Fatalf("board size = %dx%d, want 9x5", hex.Cols, hex.Rows)
	}
	if len(hex.AllBoardCoords) != 41 {
		t.Fatalf("valid cells = %d, want 41", len(hex.AllBoardCoords))
	}

	placeable := 0
	for _, c := range hex.AllBoardCoords {
		if c.Kind() != hex.CellSlot {
			continue
		}
		placeable++
	}
	if placeable != 39 {
		t.Fatalf("placeable slots = %d, want 39", placeable)
	}

	// Slots beside the emitter, including column-2 top/bottom extension rows.
	for _, c := range []hex.Coord{
		{Q: 0, R: 1},
		{Q: 0, R: 3},
		{Q: 1, R: 0},
		{Q: 1, R: 1},
		{Q: 1, R: 3},
		{Q: 1, R: 4},
	} {
		if !c.Valid() || c.Kind() != hex.CellSlot {
			t.Fatalf("(%d,%d) should be a slot", c.Q, c.R)
		}
	}

	// Column 1 extension rows stay out of bounds.
	for _, c := range []hex.Coord{
		{Q: 0, R: 0},
		{Q: 0, R: 4},
	} {
		if c.Valid() {
			t.Fatalf("(%d,%d) should be out of bounds", c.Q, c.R)
		}
	}

	// Column 9 stays playable on the top and bottom extension rows.
	for _, r := range []int{0, hex.Rows - 1} {
		c := hex.Coord{Q: 8, R: r}
		if !c.Valid() || c.Kind() != hex.CellSlot {
			t.Fatalf("(%d,%d) should be a slot", c.Q, c.R)
		}
	}

	// Column 6 (Q=5) top/bottom extension rows are slots.
	for _, r := range []int{0, hex.Rows - 1} {
		c := hex.Coord{Q: 5, R: r}
		if !c.Valid() || c.Kind() != hex.CellSlot {
			t.Fatalf("(%d,%d) should be a slot", c.Q, c.R)
		}
	}

	// Reactor wall column on rows 1 and 3.
	for _, r := range []int{1, 3} {
		c := hex.Coord{Q: hex.ReactorWallCol, R: r}
		if !c.HasWallRight() {
			t.Fatalf("(%d,%d) should have reactor wall to the right", c.Q, c.R)
		}
	}
}

func TestEmitterAndTurbineOnMiddleRow(t *testing.T) {
	emitter := hex.Coord{Q: hex.EmitterCol, R: hex.EmitterRow}
	if emitter.Kind() != hex.CellEmitter {
		t.Fatal("emitter should sit on middle row")
	}
	turbine := hex.Coord{Q: hex.TurbineCol, R: hex.TurbineRow}
	if turbine.Kind() != hex.CellTurbine {
		t.Fatal("turbine should sit on middle row")
	}
}

func TestPlayerHalves(t *testing.T) {
	for _, c := range hex.AllBoardCoords {
		if c.IsEmitter() || c.IsTurbine() {
			continue
		}
		if c.Q <= hex.Player1MaxCol && !c.IsPlayer1() {
			t.Fatalf("(%d,%d) should be player 1", c.Q, c.R)
		}
		if c.Q >= hex.Player2MinCol && !c.IsPlayer2() {
			t.Fatalf("(%d,%d) should be player 2", c.Q, c.R)
		}
	}
}
