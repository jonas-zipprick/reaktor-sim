// Package hex provides coordinates for the 9-column, 5-row hex board.
package hex

// Board dimensions from gameRules.md "Das Spielfeld im Detail".
const (
	Cols = 9
	Rows = 5

	// Column indices (0-based). Columns 1–5 are player 1, 6–9 are player 2.
	Player1MaxCol = 4
	Player2MinCol = 5

	EmitterCol = 0
	EmitterRow = 2
	TurbineCol = 4
	TurbineRow = 2

	// ReactorWallCol is the player-1 column with a fixed wall to player 2 on rows 1 and 3.
	ReactorWallCol = 3
)

// Coord is an odd-r offset hex (pointy-top). Q = column, R = row.
type Coord struct {
	Q int
	R int
}

// CellKind describes a board position.
type CellKind int

const (
	CellOutOfBounds CellKind = iota
	CellSlot
	CellEmitter
	CellTurbine
)

// AllBoardCoords lists every in-bounds hex in row-major order.
var AllBoardCoords = func() []Coord {
	out := make([]Coord, 0, 40)
	for r := 0; r < Rows; r++ {
		for q := 0; q < Cols; q++ {
			c := Coord{Q: q, R: r}
			if c.Valid() {
				out = append(out, c)
			}
		}
	}
	return out
}()

func (c Coord) Valid() bool {
	if c.R < 0 || c.R >= Rows || c.Q < 0 || c.Q >= Cols {
		return false
	}
	// Column 1 (Q=0): slots above and below the emitter on the middle row.
	if c.Q == 0 {
		return c.R >= 1 && c.R <= 3
	}
	// Column 2 (Q=1): out of bounds on the top and bottom extension rows.
	if c.Q == 1 && (c.R == 0 || c.R == Rows-1) {
		return false
	}
	// Center column (Q=4): out of bounds on the top and bottom extension rows.
	if c.Q == TurbineCol && (c.R == 0 || c.R == Rows-1) {
		return false
	}
	// Column 6 (Q=5): out of bounds on the top and bottom extension rows.
	if c.Q == 5 && (c.R == 0 || c.R == Rows-1) {
		return false
	}
	return true
}

func (c Coord) Kind() CellKind {
	if !c.Valid() {
		return CellOutOfBounds
	}
	if c.IsEmitter() {
		return CellEmitter
	}
	if c.IsTurbine() {
		return CellTurbine
	}
	return CellSlot
}

func (c Coord) IsEmitter() bool {
	return c.Q == EmitterCol && c.R == EmitterRow
}

func (c Coord) IsTurbine() bool {
	return c.Q == TurbineCol && c.R == TurbineRow
}

// IsIgniter is an alias for the emitter (Zuender).
func (c Coord) IsIgniter() bool {
	return c.IsEmitter()
}

func (c Coord) IsPlayer1() bool {
	return c.Valid() && c.Q <= Player1MaxCol
}

func (c Coord) IsPlayer2() bool {
	return c.Valid() && c.Q >= Player2MinCol
}

// wallRow reports whether the fixed player-1/player-2 wall exists in row r.
// Row 2 holds the turbine interface, so it is open.
func wallRow(r int) bool {
	return r == 1 || r == 3
}

// HasWallRight is true for player-1 cells with a fixed wall to player 2 (rows 1 and 3).
func (c Coord) HasWallRight() bool {
	return c.Q == ReactorWallCol && wallRow(c.R)
}

// WallBlocksEast returns true if a chip cannot move east from this cell into player 2.
func (c Coord) WallBlocksEast() bool {
	return c.HasWallRight()
}

// WallBlocksWest returns true if a chip cannot move west from this cell into
// player 1 (the fixed reactor wall, seen from the player-2 side).
func (c Coord) WallBlocksWest() bool {
	return c.Q == ReactorWallCol+1 && wallRow(c.R)
}

// oddRNeighborDeltas are pointy-top odd-r offsets (E, NE, NW, W, SW, SE).
var oddRNeighborDeltas = [2][6][2]int{
	{{1, 0}, {0, -1}, {-1, -1}, {-1, 0}, {-1, 1}, {0, 1}}, // even row
	{{1, 0}, {1, -1}, {0, -1}, {-1, 0}, {0, 1}, {1, 1}},   // odd row
}

func (c Coord) Neighbor(dir int) Coord {
	d := oddRNeighborDeltas[c.R&1][dir%6]
	return Coord{Q: c.Q + d[0], R: c.R + d[1]}
}

// EmitterShotTarget returns the first hex entered when the igniter fires in dir.
// On the 5-row board the Zünder sits on an even row where NE/SE hex neighbors are
// out-of-bounds; gameRules still allow shots into the three reactor slots beside it.
func EmitterShotTarget(dir int) Coord {
	switch Rotation(DisplayDir(dir % 6)) {
	case RotNE:
		return Coord{Q: 1, R: EmitterRow - 1}
	case RotE:
		return Coord{Q: 1, R: EmitterRow}
	case RotSE:
		return Coord{Q: 1, R: EmitterRow + 1}
	default:
		return Coord{Q: -1, R: -1}
	}
}

// StepTarget is the next cell for a chip at c moving in dir.
func (c Coord) StepTarget(dir int) Coord {
	if c.IsEmitter() {
		if t := EmitterShotTarget(dir); t.Valid() {
			return t
		}
	}
	return c.Neighbor(dir)
}

// ReflectDirection mirrors a chip off a wall.
func ReflectDirection(dir int) int {
	return (dir + 3) % 6
}

// DeflectDirection applies a fixed mirror offset to the incoming direction.
func DeflectDirection(incoming, offset int) int {
	return (incoming + offset) % 6
}

// CanEnter reports whether a chip may move from one valid cell to another.
func CanEnter(from, to Coord) bool {
	if !to.Valid() {
		return false
	}
	// The fixed reactor wall between columns 4 and 5 (Q=3|Q=4, rows 1 and 3) blocks
	// the straight E/W crossing in both directions; the diagonal turbine edges stay open.
	if from.WallBlocksEast() && to.Q > from.Q && to.R == from.R {
		return false
	}
	if from.WallBlocksWest() && to.Q < from.Q && to.R == from.R {
		return false
	}
	return true
}

// Pixel returns odd-r layout coordinates for rendering (pointy-top).
func (c Coord) Pixel(radius float64) (float64, float64) {
	x := radius * 1.732050808 * (float64(c.Q) + 0.5*float64(c.R&1))
	y := radius * 1.5 * float64(c.R)
	return x, y
}
