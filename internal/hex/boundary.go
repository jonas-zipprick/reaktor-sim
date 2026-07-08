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
	if from.WallBlocksEast() && to.Q > from.Q {
		return BoundaryInternalWall
	}
	if !to.Valid() {
		return BoundaryOuter
	}
	return BoundaryInternalWall
}

// ReflectOffOuterWall returns the outbound direction after bouncing off an outer wall.
func ReflectOffOuterWall(dir int) int {
	return ReflectDirection(dir)
}
