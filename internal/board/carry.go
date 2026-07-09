package board

// CopyPlaceableTiles overwrites placeable slots on dst from src.
func CopyPlaceableTiles(dst, src *State) {
	for _, c := range PlaceableSlots() {
		dst.Tiles[c.Q][c.R] = src.Tiles[c.Q][c.R]
	}
}
