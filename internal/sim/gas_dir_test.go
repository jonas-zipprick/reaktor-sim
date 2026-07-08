package sim_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestGasBoilerEmitsAllSixDirections(t *testing.T) {
	pos := hex.Coord{Q: 1, R: 1}
	seen := map[int]bool{}
	for seed := int64(0); seed < 500; seed++ {
		s := board.NewEmpty()
		s.Tiles[pos.Q][pos.R] = field.NewTile(field.GasBoiler, 0, 0)
		cfg := sim.DefaultConfig()
		cfg.InitialChips = []sim.Chip{{Type: sim.ChipHeat, Pos: hex.Coord{Q: 0, R: 1}, Dir: hex.RotE.TravelDir()}}
		_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(seed)), cfg)
		for _, snap := range snaps {
			if snap.Event != "Feldreaktion" {
				continue
			}
			for _, c := range snap.Queue {
				if c.Pos == pos && c.Type == sim.ChipHeat {
					seen[c.Dir] = true
				}
			}
		}
		if len(seen) == 6 {
			return
		}
	}
	t.Fatalf("gas boiler never emitted all 6 travel dirs in 500 seeds; saw %d dirs: %v", len(seen), seen)
}
