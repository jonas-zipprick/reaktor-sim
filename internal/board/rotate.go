package board

import (
	"math/rand"

	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
)

const shiftRotationLikelihood = 0.5

// ApplyRandomShiftRotations reorients already placed fields at shift start (phase 2
// in gameRules.md). Each orientable field is rotated independently with
// shiftRotationLikelihood; mirrors, relays, cooling towers, grounds, notgenerators,
// pressure valves, and superconductors pick a new facing.
func ApplyRandomShiftRotations(rng *rand.Rand, s *State) {
	for _, c := range PlaceableSlots() {
		t := s.tileAt(c)
		if t == nil || t.Type == field.Empty || t.BurnedOut {
			continue
		}
		if !field.HasRotation(t.Type) {
			continue
		}
		if rng.Float64() >= shiftRotationLikelihood {
			continue
		}
		rotateTile(t, rng)
	}
}

func rotateTile(t *field.Tile, rng *rand.Rand) {
	switch t.Type {
	case field.Mirror, field.Relay, field.CoolingTower, field.Ground, field.EmergencyGenerator, field.PressureValve, field.DistributionStation:
		t.Orientation = randomDifferentRotation(rng, t.Orientation)
	case field.Superconductor:
		t.SuperTarget = randomDifferentRotation(rng, t.SuperTarget)
	}
}

func randomDifferentRotation(rng *rand.Rand, current hex.Rotation) hex.Rotation {
	r := hex.RandomRotation(rng)
	for r == current {
		r = hex.RandomRotation(rng)
	}
	return r
}
