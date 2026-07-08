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
		return "Zu"
	}
	if c.IsTurbine() {
		return "Tu"
	}

	tile := state.Tiles[c.Q][c.R]
	if tile.BurnedOut {
		return "au"
	}

	base := baseSymbol(tile.Type)
	if base == "" {
		return "??"
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
		return "Sp"
	case field.CoalChamber:
		return "Kh"
	case field.CoolingTower:
		return "Ku"
	case field.GasBoiler:
		return "Eg"
	case field.AbsorberRod:
		return "Ab"
	case field.UraniumPlate:
		return "Ur"
	case field.Tokamak:
		return "Tk"
	case field.Relay:
		return "Re"
	case field.Transformer:
		return "Tr"
	case field.Ground:
		return "Er"
	case field.EmergencyGenerator:
		return "Ng"
	case field.LeadAccumulator:
		return "Bl"
	case field.CapacitorBank:
		return "Kd"
	case field.PumpedStorage:
		return "Pu"
	case field.HVCascade:
		return "Hv"
	case field.Superconductor:
		return "Su"
	default:
		return ""
	}
}

// Legend returns a human-readable symbol key.
func Legend() []string {
	return []string{
		"Zu = Zuender  Tu = Turbine (Schussrichtung pro Chip zufaellig: NO/O/SO)",
		"au = ausgebrannt  Sp = Spiegel  Re = Relais  Su = Supraleiter  (leere Felder nur Umriss)",
		"Rotation an Spiegel/Relais/Su: NW, NE, E, SE, SW, W (im Uhrzeigersinn)",
		"Kh = Kohle  Ku = Kuehlturm  Eg = Erdgas  Ab = Absorber  Ur = Uran  Tk = Tokamak",
		"Tr = Trafo  Er = Erdung  Ng = Notgenerator  Bl = Blei  Kd = Kondensator  Pu = Pumpspeicher  Hv = HV",
		"Zahl darunter = Ladung gebunden (n/max, *=unendlich)",
		"+nW/+nN/+nS = ungebunden (Waerme/Neutron/Spannung)  >nW/>nN/>nS = einkommend",
		"Rand-Bedarf ausserhalb: I oben  W rechts  b unten  R oben (!n = Schaden)",
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
