package hex

// BoundaryKind classifies why movement between two cells is blocked.
type BoundaryKind int

const (
	BoundaryNone BoundaryKind = iota
	BoundaryOuter
	BoundaryInternalWall
)

// BoundaryKind reports why from cannot enter to when moving in dir.
func BlockedBoundary(from, to Coord, dir int) BoundaryKind {
	if CanEnter(from, to) {
		return BoundaryNone
	}
	if !to.Valid() {
		return BoundaryOuter
	}
	if from.WallBlocksEast() && to.Q > from.Q {
		return BoundaryInternalWall
	}
	return BoundaryInternalWall
}

// ReflectOffOuterWall returns the outbound direction after bouncing off an outer wall.
func ReflectOffOuterWall(dir int) int {
	return ReflectDirection(dir)
}

// VoltageReflectsAtOuterWall reports whether a voltage chip at from moving in dir
// should bounce off a reflective outer wall (gameRules.md field layout).
func VoltageReflectsAtOuterWall(from Coord, dir int) bool {
	next := from.Neighbor(dir)
	if BlockedBoundary(from, next, dir) != BoundaryOuter {
		return false
	}
	switch {
	case from.Q == 6 && from.R == 0 && dir == RotW.TravelDir():
		return true
	case from.Q == 6 && from.R == Rows-1 && dir == RotW.TravelDir():
		return true
	case from.Q == 4 && from.R == 1 && dir == RotNE.TravelDir():
		return true
	case from.Q == 5 && from.R == 1 && dir == RotNW.TravelDir():
		return true
	case from.Q == 4 && from.R == 3 && dir == RotSE.TravelDir():
		return true
	case from.Q == 5 && from.R == 3 && dir == RotSW.TravelDir():
		return true
	}
	return false
}

// HeatReflectsAtOuterWall reports whether P1 heat should bounce off the player-1
// outline at from moving in dir. Every P1 outer edge reflects heat except the
// slots beside the turbine (column 5, rows above/below) where only voltage
// reflects on selected diagonals (gameRules.md).
func HeatReflectsAtOuterWall(from Coord, dir int) bool {
	if !from.IsPlayer1() {
		return false
	}
	if from.Q == TurbineCol && (from.R == TurbineRow-1 || from.R == TurbineRow+1) {
		return false
	}
	next := from.Neighbor(dir)
	if BlockedBoundary(from, next, dir) != BoundaryOuter {
		return false
	}
	return !VoltageReflectsAtOuterWall(from, dir)
}

// Player2VoltageReflects is an alias kept for older call sites.
func Player2VoltageReflects(from Coord, dir int) bool {
	return VoltageReflectsAtOuterWall(from, dir)
}
