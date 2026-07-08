package hex

// Rotation is a hex orientation: 0=northwest, then clockwise (NE, E, SE, SW, W).
type Rotation int

const (
	RotNW Rotation = 0
	RotNE Rotation = 1
	RotE  Rotation = 2
	RotSE Rotation = 3
	RotSW Rotation = 4
	RotW  Rotation = 5
)

// ShootRotations are valid firing orientations for emitter and turbine.
var ShootRotations = []Rotation{RotNE, RotE, RotSE}

// travelDirs maps rotation to internal movement direction indices.
var travelDirs = [6]int{2, 1, 0, 5, 4, 3}

// TravelDir returns the movement direction index for this rotation.
func (r Rotation) TravelDir() int {
	return travelDirs[normalizeRot(int(r))]
}

// displayDirForTravel maps internal travel indices to labeled edges 0=NW..5=W clockwise.
var displayDirForTravel = [6]Rotation{RotE, RotNE, RotNW, RotW, RotSW, RotSE}

// DisplayDir converts an internal travel direction to the labeled edge index (0=NW, clockwise).
func DisplayDir(travelDir int) int {
	return int(displayDirForTravel[normalizeRot(travelDir)])
}

// DirNames are compass labels for rotations 0=NW, then clockwise.
var DirNames = [6]string{"NW", "NE", "E", "SE", "SW", "W"}

// String returns the compass label for this rotation.
func (r Rotation) String() string {
	return DirNames[normalizeRot(int(r))]
}

// DisplayDirName returns the compass label for an internal travel direction.
func DisplayDirName(travelDir int) string {
	return Rotation(DisplayDir(travelDir)).String()
}

// ValidShootTravelDir reports whether travelDir is a legal emitter/turbine shot (NE, E, SE).
func ValidShootTravelDir(travelDir int) bool {
	return Rotation(DisplayDir(travelDir)).ValidShoot()
}

// ValidShoot reports whether r is a legal emitter/turbine shot (NE, E, or SE).
func (r Rotation) ValidShoot() bool {
	return r == RotNE || r == RotE || r == RotSE
}

// WireOutgoing returns the outbound travel direction for a chip crossing a mirror
// or relay. Orientation marks the open line between opposite edges (o and o+3).
// Approaches parallel to that line pass through; all others are deflected 60°
// away from the line.
func (r Rotation) WireOutgoing(incoming int) int {
	incoming = normalizeRot(incoming)
	parallelA := r.TravelDir()
	parallelB := Rotation(normalizeRot(int(r) + 3)).TravelDir()
	if incoming == parallelA || incoming == parallelB {
		return (incoming + 3) % 6
	}
	prev := (incoming + 5) % 6
	next := (incoming + 1) % 6
	if prev == parallelA || prev == parallelB {
		return next
	}
	return prev
}

func normalizeRot(v int) int {
	v %= 6
	if v < 0 {
		v += 6
	}
	return v
}

// RandomRotation returns a uniform random orientation 0-5.
func RandomRotation(rng RandIntn) Rotation {
	return Rotation(rng.Intn(6))
}

// RandomShootRotation returns a random legal shot orientation (1, 2, or 3).
func RandomShootRotation(rng RandIntn) Rotation {
	return ShootRotations[rng.Intn(len(ShootRotations))]
}

// RandomTravelDir returns a uniformly random travel direction (Richtungswuerfel: alle 6 Kanten).
func RandomTravelDir(rng RandIntn) int {
	return RandomRotation(rng).TravelDir()
}

// RandomShootDir returns a uniformly random emitter/turbine travel direction (NE, E, SE).
func RandomShootDir(rng RandIntn) int {
	return ShootRotations[rng.Intn(len(ShootRotations))].TravelDir()
}

// RandIntn is the subset of rand.Rand used for board generation.
type RandIntn interface {
	Intn(n int) int
}
