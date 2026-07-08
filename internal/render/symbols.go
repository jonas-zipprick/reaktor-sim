package render

import (
	"fmt"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
)

// BottomLabel returns bound charge below a cell symbol.
func BottomLabel(state *board.State, c hex.Coord, tile field.Tile) string {
	return ChargeLabel(tile)
}

// ChargeLabel returns current charge as "n", "n/max", or "*" for unlimited fuel.
// Storage fields show stored voltage chips instead of fuel charge.
func ChargeLabel(tile field.Tile) string {
	if tile.BurnedOut || tile.Type == field.Empty {
		return ""
	}
	info, ok := field.Catalog[tile.Type]
	if !ok {
		return ""
	}

	switch tile.Type {
	case field.CapacitorBank, field.PumpedStorage, field.LeadAccumulator:
		return fmt.Sprintf("%d/%d", tile.StoredVoltage, info.MaxCharge)
	case field.Tokamak:
		return "*"
	}

	if info.InitialCharge == -1 {
		return "*"
	}
	if info.MaxCharge > 0 {
		return fmt.Sprintf("%d/%d", tile.Charge, info.MaxCharge)
	}
	if info.InitialCharge > 0 {
		return fmt.Sprintf("%d", tile.Charge)
	}
	return ""
}

// Label returns the cell symbol including orientation where relevant.
func Label(state *board.State, c hex.Coord) string {
	if c.IsEmitter() {
		return "Z"
	}
	if c.IsTurbine() {
		return "@"
	}

	tile := state.Tiles[c.Q][c.R]
	if tile.BurnedOut {
		return "x"
	}

	base := baseSymbol(tile.Type)
	if base == "" {
		return "?"
	}

	switch tile.Type {
	case field.Mirror, field.Relay:
		return fmt.Sprintf("%s%d", base, tile.Orientation)
	case field.Superconductor:
		return fmt.Sprintf("%s%d", base, tile.SuperTarget)
	default:
		return base
	}
}

func baseSymbol(t field.Type) string {
	switch t {
	case field.Empty:
		return ""
	case field.Mirror:
		return "%"
	case field.CoalChamber:
		return "C"
	case field.CoolingTower:
		return "~"
	case field.GasBoiler:
		return "G"
	case field.AbsorberRod:
		return "A"
	case field.UraniumPlate:
		return "U"
	case field.Tokamak:
		return "O"
	case field.Relay:
		return "Re"
	case field.Transformer:
		return "t"
	case field.Ground:
		return "E"
	case field.EmergencyGenerator:
		return "N"
	case field.LeadAccumulator:
		return "B"
	case field.CapacitorBank:
		return "Ko"
	case field.PumpedStorage:
		return "P"
	case field.HVCascade:
		return "H"
	case field.Superconductor:
		return "S"
	default:
		return ""
	}
}

// Legend returns a human-readable symbol key.
func Legend() []string {
	return []string{
		"Z = Zuender  @ = Turbine (Schussrichtung pro Chip zufaellig: NO/O/SO)",
		"x=ausgebrannt  %=Spiegel  Re=Relais  S=Supraleiter  (leere Felder nur Umriss)",
		"Rotation an Spiegel/Relais/S: NW, NE, E, SE, SW, W (im Uhrzeigersinn)",
		"C=Kohle  ~=Kuehlturm  G=Erdgas  A=Absorber  U=Uran  O=Tokamak",
		"t=Trafo  E=Erdung  N=Notgenerator  B=Blei  Ko=Kondensator  P=Pumpspeicher  H=HV",
		"Zahl darunter = Ladung gebunden (n/max, *=unendlich)",
		"+nW/+nN/+nS = ungebunden (Waerme/Neutron/Spannung)  >nW/>nN/>nS = einkommend",
		"Rand-Bedarf ausserhalb: I oben  W rechts  b unten  R oben (Turbine)",
		"| = Schnittstelle (zwischen Spalte 5 und 6)",
	}
}

// ZoneMarker returns a border demand hint for player-2 edges.
func ZoneMarker(c hex.Coord) string {
	z, ok := board.ZoneOf(c)
	if !ok {
		return ""
	}
	switch z {
	case board.ZoneIndustry:
		return "I"
	case board.ZoneResidential:
		return "W"
	case board.ZoneRail:
		return "b"
	case board.ZonePlant:
		return "R"
	default:
		return ""
	}
}
