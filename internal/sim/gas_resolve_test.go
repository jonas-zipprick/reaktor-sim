package sim_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestGasBoilerEachDirectionResolves(t *testing.T) {
	pos := hex.Coord{Q: 1, R: 1}
	for dir := 0; dir < 6; dir++ {
		t.Run(hex.DisplayDirName(dir), func(t *testing.T) {
			s := board.NewEmpty()
			s.Tiles[pos.Q][pos.R] = field.NewTile(field.GasBoiler, 0, 0)
			cfg := sim.DefaultConfig()
			cfg.CriticalLimit = 100
			cfg.InitialChips = []sim.Chip{{Type: sim.ChipHeat, Pos: pos, Dir: dir}}
			res := sim.Run(s, rand.New(rand.NewSource(42)), cfg)
			if res.CriticalFailure {
				t.Fatalf("direction %s caused immediate loss", hex.DisplayDirName(dir))
			}
		})
	}
}
