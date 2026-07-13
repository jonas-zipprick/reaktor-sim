// Package field defines board tile types and their game properties.
package field

import "github.com/jonas/reaktor-sim/internal/hex"

// Type identifies a placed field on the hex grid.
type Type int

const (
	Empty Type = iota

	// Reactor (player 1).
	Remove
	Mirror
	CoalChamber
	CoolingTower
	GasBoiler
	AbsorberRod
	UraniumPlate
	Tokamak

	// Grid (player 2).
	Relay
	Transformer
	Ground
	EmergencyGenerator
	LeadAccumulator
	CapacitorBank
	PumpedStorage
	HVCascade
	Superconductor
)

// Info holds static metadata for a field type.
type Info struct {
	Name               string
	Cost               int
	InitialCharge      int // -1 = infinite, -2 = special storage max
	MaxCharge          int
	Sector             string // "reactor" or "grid"
	AvailableFromMonth int    // earliest campaign month (1 = from start)
}

var Catalog = map[Type]Info{
	Remove:             {Name: "Feld entfernen", Cost: 1, Sector: "reactor"},
	Mirror:             {Name: "Ablenk-Spiegel", Cost: 1, Sector: "reactor"},
	CoalChamber:        {Name: "Kohle-Brennkammer", Cost: 2, InitialCharge: 4, MaxCharge: 4, Sector: "reactor"},
	CoolingTower:       {Name: "Kühlturm", Cost: 2, Sector: "reactor"},
	GasBoiler:          {Name: "Erdgas-Kessel", Cost: 3, InitialCharge: 8, MaxCharge: 8, Sector: "reactor", AvailableFromMonth: 2},
	AbsorberRod:        {Name: "Absorber-Stab", Cost: 3, Sector: "reactor", AvailableFromMonth: 3},
	UraniumPlate:       {Name: "Uran-Platte", Cost: 5, InitialCharge: 2, MaxCharge: 2, Sector: "reactor", AvailableFromMonth: 3},
	Tokamak:            {Name: "Tokamak-Kammer", Cost: 8, InitialCharge: -1, Sector: "reactor", AvailableFromMonth: 4},
	Relay:              {Name: "Relais/Weiche", Cost: 1, Sector: "grid"},
	Transformer:        {Name: "Transformator", Cost: 3, InitialCharge: 6, MaxCharge: 6, Sector: "grid"},
	Ground:             {Name: "Erdung/Widerstand", Cost: 1, Sector: "grid"},
	EmergencyGenerator: {Name: "Notgenerator", Cost: 3, InitialCharge: 2, MaxCharge: 2, Sector: "grid"},
	LeadAccumulator:    {Name: "Blei-Akkumulator", Cost: 3, MaxCharge: 2, Sector: "grid", AvailableFromMonth: 2},
	CapacitorBank:      {Name: "Kondensator-Bank", Cost: 4, MaxCharge: 5, Sector: "grid", AvailableFromMonth: 2},
	PumpedStorage:      {Name: "Pumpspeicherwerk", Cost: 4, MaxCharge: 5, Sector: "grid", AvailableFromMonth: 3},
	HVCascade:          {Name: "Hochspannungs-Kaskade", Cost: 3, InitialCharge: 8, MaxCharge: 8, Sector: "grid", AvailableFromMonth: 3},
	Superconductor:     {Name: "Supraleiter", Cost: 4, Sector: "grid", AvailableFromMonth: 4},
}

// ReactorMarket lists placeable reactor fields for random generation.
var ReactorMarket = []Type{
	Mirror, CoalChamber, CoolingTower, GasBoiler, AbsorberRod, UraniumPlate, Tokamak,
}

// GridMarket lists placeable grid fields for random generation.
var GridMarket = []Type{
	Relay, Transformer, Ground, EmergencyGenerator,
	LeadAccumulator, CapacitorBank, PumpedStorage, HVCascade, Superconductor,
}

// Tile is a field instance on the board.
type Tile struct {
	Type           Type
	Charge         int
	BurnedOut      bool
	Orientation    hex.Rotation // mirror/relay facing (0-5)
	TokamakCounter int
	StoredVoltage  int
	SuperTarget    hex.Rotation // superconductor aim (0-5)
}

func NewTile(t Type, orientation, superTarget hex.Rotation) Tile {
	info := Catalog[t]
	charge := info.InitialCharge
	if charge < 0 {
		charge = 0
	}
	max := info.MaxCharge
	if info.InitialCharge == -1 {
		charge = 0
	}
	if max == 0 && info.MaxCharge > 0 {
		max = info.MaxCharge
	}
	return Tile{
		Type:        t,
		Charge:      charge,
		Orientation: orientation,
		SuperTarget: superTarget,
	}
}

func (t *Tile) IsBurnedOut() bool {
	return t.BurnedOut
}

func (t *Tile) Cost() int {
	if info, ok := Catalog[t.Type]; ok {
		return info.Cost
	}
	return 0
}

// HasRotation reports whether a placed field's facing affects simulation.
func HasRotation(t Type) bool {
	return t == Mirror || t == Relay || t == Superconductor
}
