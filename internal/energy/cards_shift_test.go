package energy_test

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/energy"
)

func TestEnergyCardShiftPlans(t *testing.T) {
	want := map[string][5]board.ShiftDemands{
		"eroeffnungsfeier": {
			{Industry: 1, Residential: 1, Rail: 0, Plant: 1},
			{Industry: 1, Residential: 1, Rail: 0, Plant: 1},
			{Industry: 0, Residential: 1, Rail: 1, Plant: 1},
			{Industry: 1, Residential: 1, Rail: 1, Plant: 1},
			{Industry: 2, Residential: 1, Rail: 1, Plant: 1},
		},
		"netzoptimierung": {
			{Industry: 0, Residential: 1, Rail: 1, Plant: 1},
			{Industry: 0, Residential: 1, Rail: 1, Plant: 1},
			{Industry: 0, Residential: 2, Rail: 0, Plant: 1},
			{Industry: 1, Residential: 1, Rail: 1, Plant: 1},
			{Industry: 1, Residential: 2, Rail: 1, Plant: 1},
		},
		"technologische-transformation": {
			{Industry: 1, Residential: 1, Rail: 0, Plant: 1},
			{Industry: 2, Residential: 1, Rail: 0, Plant: 1},
			{Industry: 1, Residential: 1, Rail: 1, Plant: 1},
			{Industry: 2, Residential: 1, Rail: 0, Plant: 1},
			{Industry: 2, Residential: 1, Rail: 1, Plant: 1},
		},
		"gossnab": {
			{Industry: 1, Residential: 1, Rail: 0, Plant: 1},
			{Industry: 2, Residential: 0, Rail: 1, Plant: 1},
			{Industry: 0, Residential: 3, Rail: 1, Plant: 1},
			{Industry: 1, Residential: 1, Rail: 0, Plant: 1},
			{Industry: 2, Residential: 1, Rail: 1, Plant: 1},
		},
		"testlauf-volllast": {
			{Industry: 1, Residential: 1, Rail: 1, Plant: 1},
			{Industry: 2, Residential: 1, Rail: 1, Plant: 1},
			{Industry: 2, Residential: 2, Rail: 1, Plant: 1},
			{Industry: 3, Residential: 1, Rail: 1, Plant: 1},
			{Industry: 2, Residential: 1, Rail: 2, Plant: 1},
		},
		"schturmowschtschina": {
			{Industry: 1, Residential: 1, Rail: 1, Plant: 1},
			{Industry: 2, Residential: 2, Rail: 1, Plant: 1},
			{Industry: 3, Residential: 2, Rail: 2, Plant: 2},
			{Industry: 4, Residential: 1, Rail: 2, Plant: 2},
			{Industry: 4, Residential: 2, Rail: 1, Plant: 2},
		},
	}

	for id, shifts := range want {
		card, ok := energy.ByID(id)
		if !ok {
			t.Fatalf("card %q missing", id)
		}
		for shift := 1; shift <= 5; shift++ {
			got := card.ShiftDemands(shift)
			if got != shifts[shift-1] {
				t.Fatalf("%s shift %d = %+v, want %+v", id, shift, got, shifts[shift-1])
			}
		}
	}
}
