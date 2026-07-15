package field

import "github.com/jonas/reaktor-sim/internal/hex"

// AllowedOnCell reports whether t may be placed on c per sector and board layout.
func AllowedOnCell(t Type, c hex.Coord) bool {
	if t == Empty {
		return true
	}
	info, ok := Catalog[t]
	if !ok {
		return false
	}
	switch info.Sector {
	case "reactor":
		if !c.IsPlayer1() || c.IsEmitter() || c.IsTurbine() {
			return false
		}
		// Turbine column slots are interface/wall cells: no fuel emitters there.
		if c.Q == hex.TurbineCol && isFuelEmitter(t) {
			return false
		}
		return true
	case "grid":
		return c.IsPlayer2()
	default:
		return false
	}
}

func isFuelEmitter(t Type) bool {
	switch t {
	case CoalChamber, GasBoiler, UraniumPlate, Tokamak:
		return true
	default:
		return false
	}
}

// FilterForCell returns market types that may be placed on c.
func FilterForCell(market []Type, c hex.Coord) []Type {
	out := make([]Type, 0, len(market))
	for _, t := range market {
		if AllowedOnCell(t, c) {
			out = append(out, t)
		}
	}
	return out
}
