package sim_test

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/energy"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestStartDemandsMatchShiftPlan(t *testing.T) {
	card, ok := energy.ByID("eroeffnungsfeier")
	if !ok {
		t.Fatal("missing card")
	}
	want := card.ShiftDemands(1)

	cfg := sim.DefaultConfig()
	cfg.EnergyCard = card
	cfg.Shift = 1
	cfg.InitialChips = []sim.Chip{}

	_, snaps := sim.RunTrace(board.NewEmpty(), rand.New(rand.NewSource(1)), cfg)
	if len(snaps) == 0 || snaps[0].Event != "Start" {
		t.Fatalf("expected Start snapshot, got %#v", snaps)
	}
	b := snaps[0].Board
	checkZoneDemand(t, b, board.ZoneIndustry, want.Industry)
	checkZoneDemand(t, b, board.ZoneResidential, want.Residential)
	checkZoneDemand(t, b, board.ZoneRail, want.Rail)
	checkZoneDemand(t, b, board.ZonePlant, want.Plant)
}

func TestPreappliedDemandsMustNotBeDoubled(t *testing.T) {
	card := energy.DefaultCard()
	cfg := sim.DefaultConfig()
	cfg.EnergyCard = card
	cfg.Shift = 1
	cfg.InitialChips = []sim.Chip{}

	s := board.NewEmpty()
	s.ApplyDemands(cfg.ShiftDemands)

	_, snaps := sim.RunTrace(s, rand.New(rand.NewSource(2)), cfg)
	b := snaps[0].Board
	want := card.ShiftDemands(1)
	if b.TotalDemand(board.ZoneIndustry) != want.Industry {
		t.Fatalf("industry demand doubled: got %d want %d", b.TotalDemand(board.ZoneIndustry), want.Industry)
	}
}

func checkZoneDemand(t *testing.T, b *board.State, z board.Zone, want int) {
	t.Helper()
	if got := b.TotalDemand(z); got != want {
		t.Fatalf("%s demand = %d, want %d", z.String(), got, want)
	}
}
