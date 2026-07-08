package sim_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestHeatReachesTurbineOnOpenPath(t *testing.T) {
	s := board.NewEmpty()
	s.Tiles[1][1] = field.NewTile(field.CoalChamber, 0, 0)
	s.Tiles[2][1] = field.NewTile(field.CoalChamber, 0, 0)
	s.Tiles[3][1] = field.NewTile(field.CoalChamber, 0, 0)

	rng := rand.New(rand.NewSource(1))
	cfg := sim.DefaultConfig()
	cfg.StartDir = hex.RotE.TravelDir()

	hits := 0
	for i := 0; i < 200; i++ {
		res := sim.Run(s, rng, cfg)
		hits += res.HeatAtTurbine
	}
	if hits == 0 {
		t.Fatal("expected some heat at turbine on open path")
	}
}

func TestDirectLineToTurbine(t *testing.T) {
	s := board.NewEmpty()
	rng := rand.New(rand.NewSource(99))
	cfg := sim.DefaultConfig()
	cfg.StartDir = hex.RotE.TravelDir()

	hits := 0
	for i := 0; i < 50; i++ {
		res := sim.Run(s, rng, cfg)
		hits += res.HeatAtTurbine
	}
	if hits == 0 {
		t.Fatal("expected heat through empty corridor")
	}
}

func TestErdgasChainToTurbine(t *testing.T) {
	s := board.NewEmpty()
	s.Tiles[1][1] = field.NewTile(field.GasBoiler, 0, 0)

	rng := rand.New(rand.NewSource(7))
	cfg := sim.DefaultConfig()
	cfg.StartDir = hex.RotE.TravelDir()

	hits := 0
	for i := 0; i < 100; i++ {
		res := sim.Run(s, rng, cfg)
		hits += res.HeatAtTurbine
	}
	if hits == 0 {
		t.Fatal("erdgas chain should reach turbine sometimes")
	}
}

func TestRandomBoardsProduceHeat(t *testing.T) {
	hits := 0
	for seed := int64(1); seed < 30; seed++ {
		rng := rand.New(rand.NewSource(seed))
		s := board.Random(rng)
		for i := 0; i < 100; i++ {
			hits += sim.Run(s, rng, sim.DefaultConfig()).HeatAtTurbine
		}
	}
	if hits == 0 {
		t.Fatal("expected heat on at least some random boards")
	}
}

func TestBoardHas25Cells(t *testing.T) {
	if len(hex.AllBoardCoords) != 25 {
		t.Fatalf("expected 25 cells, got %d", len(hex.AllBoardCoords))
	}
}

func TestShootRotations(t *testing.T) {
	for _, r := range hex.ShootRotations {
		if !r.ValidShoot() {
			t.Fatalf("rotation %d should be valid shoot", r)
		}
	}
	if hex.RotNW.ValidShoot() || hex.RotSE.ValidShoot() == false {
		t.Fatal("unexpected shoot validation")
	}
}
