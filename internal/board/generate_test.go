package board

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/rules"
)

func TestPlaceableSlots(t *testing.T) {
	slots := PlaceableSlots()
	if len(slots) != 33 {
		t.Fatalf("expected 33 slots, got %d", len(slots))
	}
}

func TestRandomWithCostExact(t *testing.T) {
	targets := []int{0, 1, 7, 42, 76, 100}
	for _, target := range targets {
		t.Run("", func(t *testing.T) {
			t.Parallel()
			rng := rand.New(rand.NewSource(int64(target * 17)))
			s, err := RandomWithCost(rng, target)
			if err != nil {
				t.Fatalf("target %d: %v", target, err)
			}
			if got := s.TotalCost(); got != target {
				t.Fatalf("target %d: got cost %d", target, got)
			}
		})
	}
}

func TestRandomWithCostVariesBySeed(t *testing.T) {
	const target = 50
	a, err := RandomWithCost(rand.New(rand.NewSource(1)), target)
	if err != nil {
		t.Fatal(err)
	}
	b, err := RandomWithCost(rand.New(rand.NewSource(2)), target)
	if err != nil {
		t.Fatal(err)
	}
	if a.TotalCost() != target || b.TotalCost() != target {
		t.Fatalf("costs %d and %d", a.TotalCost(), b.TotalCost())
	}
	same := true
	for q := 0; q < len(a.Tiles); q++ {
		for r := 0; r < len(a.Tiles[q]); r++ {
			if a.Tiles[q][r].Type != b.Tiles[q][r].Type {
				same = false
				break
			}
		}
	}
	if same {
		t.Fatal("expected different random boards for different seeds")
	}
}

func TestRandomWithCostNegative(t *testing.T) {
	_, err := RandomWithCost(rand.New(rand.NewSource(1)), -1)
	if err == nil {
		t.Fatal("expected error for negative cost")
	}
}

func TestAllCostsAchievableUpToMax(t *testing.T) {
	slots := PlaceableSlots()
	planner, err := newCostPlanner(slots, 0, rules.Month{})
	if err != nil {
		t.Fatal(err)
	}
	if !planner.achievable[planner.maxCost] {
		t.Fatalf("expected max total %d to be achievable", planner.maxCost)
	}
	for target := 0; target <= planner.maxCost; target++ {
		if !planner.achievable[target] {
			t.Fatalf("expected target %d to be achievable", target)
		}
	}
}
