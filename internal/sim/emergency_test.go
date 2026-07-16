package sim_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/energy"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestEmergencyGeneratorFiresAlongOrientation(t *testing.T) {
	pos := hex.Coord{Q: 6, R: 1}
	s := board.NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.Tile{
		Type:        field.EmergencyGenerator,
		Charge:      1,
		Orientation: hex.RotE,
	}
	s.ApplyDemands(board.ShiftDemands{Residential: 1})

	cfg := sim.DefaultConfig()
	cfg.EnergyCard = energy.Card{}
	cfg.ShiftDemands = board.ShiftDemands{Residential: 1}
	cfg.InitialChips = []sim.Chip{}

	wantDir := hex.RotE.TravelDir()
	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(1)), cfg)
	for _, snap := range snaps {
		if snap.Event != "Freiwilliger Schuss" {
			continue
		}
		for _, c := range snap.Queue {
			if c.Pos == pos && c.Type == sim.ChipVoltage {
				if c.Dir != wantDir {
					t.Fatalf("fired dir = %d (%s), want %d (%s)",
						c.Dir, hex.DisplayDirName(c.Dir), wantDir, hex.DisplayDirName(wantDir))
				}
				return
			}
		}
	}
	t.Fatal("expected voluntary notgenerator shot")
}

func TestEmergencyGeneratorSurvivesVoltageHit(t *testing.T) {
	pos := hex.Coord{Q: 6, R: 1}
	s := board.NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.NewTile(field.EmergencyGenerator, 0, 0)

	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  hex.Coord{Q: 5, R: 1},
		Dir:  hex.RotE.TravelDir(),
	}}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(1)), cfg)
	if s.Tiles[pos.Q][pos.R].Type != field.EmergencyGenerator {
		t.Fatalf("tile type = %v, want emergency generator", s.Tiles[pos.Q][pos.R].Type)
	}
	if got := s.Tiles[pos.Q][pos.R].Charge; got != 1 {
		t.Fatalf("charge = %d, want 1 (incoming voltage must not add charge)", got)
	}
	destroyed := false
	for _, snap := range snaps {
		if snap.Narrative == "Spannung trifft Notgenerator und wird vernichtet." {
			destroyed = true
			break
		}
	}
	if !destroyed {
		t.Fatal("expected loaded notgenerator to destroy incoming voltage")
	}
}

func TestEmergencyGeneratorDestroysIncomingVoltageWhileLoaded(t *testing.T) {
	pos := hex.Coord{Q: 6, R: 1}
	s := board.NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.NewTile(field.EmergencyGenerator, 0, 0)

	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  hex.Coord{Q: 5, R: 1},
		Dir:  hex.RotE.TravelDir(),
	}}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(1)), cfg)
	destroyed := false
	for _, snap := range snaps {
		if snap.Event == "Weiterleitung" {
			t.Fatal("loaded notgenerator must not redirect incoming voltage")
		}
		if snap.Narrative == "Spannung trifft Notgenerator und wird vernichtet." {
			destroyed = true
		}
	}
	if !destroyed {
		t.Fatal("expected incoming voltage to be destroyed at loaded notgenerator")
	}
}

func TestBurnedEmergencyGeneratorRedirectsVoltage(t *testing.T) {
	pos := hex.Coord{Q: 6, R: 1}
	s := board.NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.Tile{Type: field.EmergencyGenerator, BurnedOut: true}

	cfg := sim.DefaultConfig()
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  hex.Coord{Q: 5, R: 1},
		Dir:  hex.RotE.TravelDir(),
	}}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(3)), cfg)

	redirected := false
	for _, snap := range snaps {
		if snap.Event == "Weiterleitung" {
			redirected = true
			break
		}
	}
	if !redirected {
		t.Fatal("expected burned emergency generator to redirect voltage")
	}
	if s.Tiles[pos.Q][pos.R].Type != field.EmergencyGenerator {
		t.Fatalf("tile type = %v, want emergency generator still on board", s.Tiles[pos.Q][pos.R].Type)
	}
}

func TestEmergencyGeneratorBurnsOutAfterVoluntaryFire(t *testing.T) {
	pos := hex.Coord{Q: 6, R: 1}
	s := board.NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.NewTile(field.EmergencyGenerator, 0, 0)
	s.ApplyDemands(board.ShiftDemands{Residential: 1})
	if s.Tiles[pos.Q][pos.R].Charge != 1 {
		t.Fatalf("setup charge = %d, want 1", s.Tiles[pos.Q][pos.R].Charge)
	}

	cfg := sim.DefaultConfig()
	cfg.ShiftDemands = board.ShiftDemands{Residential: 1}
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  hex.Coord{Q: 5, R: 1},
		Dir:  hex.RotE.TravelDir(),
	}}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(42)), cfg)
	for i := len(snaps) - 1; i >= 0; i-- {
		tile := snaps[i].Board.Tiles[pos.Q][pos.R]
		if tile.Type == field.EmergencyGenerator && tile.BurnedOut {
			return
		}
	}
	t.Fatal("expected emergency generator to burn out after charge spent")
}

func TestEmergencyGeneratorNoVoluntaryFireWithoutDemand(t *testing.T) {
	pos := hex.Coord{Q: 6, R: 1}
	s := board.NewEmpty()
	s.Tiles[pos.Q][pos.R] = field.NewTile(field.EmergencyGenerator, 0, 0)

	cfg := sim.DefaultConfig()
	cfg.ShiftDemands = board.ShiftDemands{}
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  hex.Coord{Q: 5, R: 1},
		Dir:  hex.RotE.TravelDir(),
	}}

	sim.RunTrace(s, rand.New(rand.NewSource(7)), cfg)
	if got := s.Tiles[pos.Q][pos.R].Charge; got != 1 {
		t.Fatalf("charge = %d, want 1 (no fire without demand)", got)
	}
}

func TestEmergencyGeneratorWaitsForOtherVoltage(t *testing.T) {
	ngPos := hex.Coord{Q: 6, R: 1}
	vPos := hex.Coord{Q: 5, R: 1}
	s := board.NewEmpty()
	s.Tiles[ngPos.Q][ngPos.R] = field.NewTile(field.EmergencyGenerator, 0, 0)
	s.ApplyDemands(board.ShiftDemands{Residential: 1})

	cfg := sim.DefaultConfig()
	cfg.ShiftDemands = board.ShiftDemands{Residential: 1}
	cfg.InitialChips = []sim.Chip{{
		Type: sim.ChipVoltage,
		Pos:  vPos,
		Dir:  hex.RotE.TravelDir(),
	}}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(11)), cfg)
	for _, snap := range snaps {
		hasOtherVoltage := snap.Active != nil &&
			snap.Active.Type == sim.ChipVoltage && snap.Active.Pos != ngPos
		if !hasOtherVoltage {
			for _, c := range snap.Queue {
				if c.Type == sim.ChipVoltage && c.Pos != ngPos {
					hasOtherVoltage = true
					break
				}
			}
		}
		if !hasOtherVoltage {
			continue
		}
		if snap.Event == "Freiwilliger Schuss" {
			for _, c := range snap.Queue {
				if c.Pos == ngPos && c.Type == sim.ChipVoltage {
					t.Fatalf("notgenerator fired at step %d while other voltage still on board", snap.Step)
				}
			}
		}
	}
}

func TestEmergencyGeneratorWaitsForStoredVoltage(t *testing.T) {
	ngPos := hex.Coord{Q: 6, R: 1}
	capPos := hex.Coord{Q: 7, R: 1}
	s := board.NewEmpty()
	s.Tiles[ngPos.Q][ngPos.R] = field.NewTile(field.EmergencyGenerator, 0, 0)
	s.Tiles[capPos.Q][capPos.R] = field.NewTile(field.CapacitorBank, 0, 0)
	s.Tiles[capPos.Q][capPos.R].StoredVoltage = 2
	s.ApplyDemands(board.ShiftDemands{Residential: 1})

	cfg := sim.DefaultConfig()
	cfg.ShiftDemands = board.ShiftDemands{Residential: 1}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(17)), cfg)
	for _, snap := range snaps {
		if snap.Event != "Freiwilliger Schuss" {
			continue
		}
		for _, c := range snap.Queue {
			if c.Pos == ngPos && c.Type == sim.ChipVoltage {
				if snap.Board.Tiles[capPos.Q][capPos.R].StoredVoltage > 0 {
					t.Fatalf("notgenerator fired at step %d while capacitor still held voltage", snap.Step)
				}
			}
		}
	}
}

func TestEmergencyGeneratorWaitsForPlayer1Heat(t *testing.T) {
	ngPos := hex.Coord{Q: 6, R: 1}
	s := board.NewEmpty()
	s.Tiles[ngPos.Q][ngPos.R] = field.NewTile(field.EmergencyGenerator, 0, 0)
	s.ApplyDemands(board.ShiftDemands{Residential: 1})

	cfg := sim.DefaultConfig()
	cfg.ShiftDemands = board.ShiftDemands{Residential: 1}
	cfg.InitialChips = []sim.Chip{
		{Type: sim.ChipHeat, Pos: hex.Coord{Q: 1, R: 1}, Dir: hex.RotE.TravelDir()},
		{Type: sim.ChipVoltage, Pos: hex.Coord{Q: 5, R: 1}, Dir: hex.RotE.TravelDir()},
	}

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(11)), cfg)
	lastHeatStep := -1
	firstNGFireStep := -1
	for _, snap := range snaps {
		for _, c := range snap.Queue {
			if c.Type == sim.ChipHeat && c.Pos.IsPlayer1() {
				lastHeatStep = snap.Step
			}
		}
		if snap.Event != "Freiwilliger Schuss" {
			continue
		}
		for _, c := range snap.Queue {
			if c.Pos == ngPos && c.Type == sim.ChipVoltage {
				if firstNGFireStep < 0 {
					firstNGFireStep = snap.Step
				}
			}
		}
	}
	if firstNGFireStep >= 0 && firstNGFireStep <= lastHeatStep {
		t.Fatalf("notgenerator fired at step %d while P1 heat still present until step %d",
			firstNGFireStep, lastHeatStep)
	}
}

func TestEmitterSkippedWithoutDemand(t *testing.T) {
	s := board.NewEmpty()
	cfg := sim.DefaultConfig()
	cfg.EnergyCard = energy.Card{}
	cfg.ShiftDemands = board.ShiftDemands{}
	cfg.InitialChips = nil

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(3)), cfg)
	if len(snaps) == 0 {
		t.Fatal("no snapshots")
	}
	for _, c := range snaps[0].Queue {
		if c.Pos.IsEmitter() {
			t.Fatalf("unexpected emitter chip at shift start: %+v", c)
		}
	}
	if chips := sim.EmitterChips(s, cfg, rand.New(rand.NewSource(1))); len(chips) != 0 {
		t.Fatalf("EmitterChips = %d, want 0 without demand", len(chips))
	}
}

func TestEmitterFiresWithDemand(t *testing.T) {
	s := board.NewEmpty()
	cfg := sim.DefaultConfig()
	cfg.EnergyCard = energy.Card{}
	cfg.ShiftDemands = board.ShiftDemands{Plant: 1}
	cfg.InitialChips = nil

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(5)), cfg)
	if len(snaps) == 0 {
		t.Fatal("no snapshots")
	}
	found := false
	for _, c := range snaps[0].Queue {
		if c.Pos.IsEmitter() {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected emitter chip at shift start when demand exists")
	}
}
