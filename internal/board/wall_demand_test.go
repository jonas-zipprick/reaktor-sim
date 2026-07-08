package board

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/hex"
)

func TestZonesForOuterWallHitSEFromBottomRow(t *testing.T) {
	zones := ZonesForOuterWallHit(hex.Coord{Q: 5, R: 2}, hex.RotSE.TravelDir())
	if len(zones) != 1 || zones[0] != ZoneRail {
		t.Fatalf("SE from (5,2) = %v, want [Bahn]", zones)
	}
}

func TestZonesForOuterWallHitSEFromRightColumn(t *testing.T) {
	zones := ZonesForOuterWallHit(hex.Coord{Q: 8, R: 1}, hex.RotSE.TravelDir())
	if len(zones) != 1 || zones[0] != ZoneResidential {
		t.Fatalf("SE from (8,1) = %v, want [Wohnviertel]", zones)
	}
}

func TestZonesForOuterWallHitNWTopLeftCorner(t *testing.T) {
	zones := ZonesForOuterWallHit(hex.Coord{Q: 5, R: 0}, hex.RotNW.TravelDir())
	if len(zones) != 2 {
		t.Fatalf("NW from (5,0) = %v, want Industrie+Reaktor", zones)
	}
}

func TestTryConsumeWallDemandSEBottomRowRailOnly(t *testing.T) {
	s := NewEmpty()
	s.ApplyDemands(ShiftDemands{Residential: 1, Rail: 1})

	z, ok := s.TryConsumeWallDemand(hex.Coord{Q: 5, R: 2}, hex.RotSE.TravelDir(), rand.New(rand.NewSource(1)))
	if !ok || z != ZoneRail {
		t.Fatalf("got zone %v ok=%v, want Rail", z, ok)
	}
	if s.TotalDemand(ZoneResidential) != 1 {
		t.Fatal("residential demand must not be consumed")
	}
}

func TestTryConsumeWallDemandSEBottomRowSpikesWhenRailEmpty(t *testing.T) {
	s := NewEmpty()
	s.ApplyDemands(ShiftDemands{Residential: 1})

	_, ok := s.TryConsumeWallDemand(hex.Coord{Q: 5, R: 2}, hex.RotSE.TravelDir(), rand.New(rand.NewSource(1)))
	if ok {
		t.Fatal("expected no delivery when only residential demand remains for bottom wall")
	}
}

func TestTryConsumeWallDemandEastResidential(t *testing.T) {
	s := NewEmpty()
	s.ApplyDemands(ShiftDemands{Residential: 1})

	z, ok := s.TryConsumeWallDemand(hex.Coord{Q: 8, R: 1}, hex.RotE.TravelDir(), rand.New(rand.NewSource(1)))
	if !ok || z != ZoneResidential {
		t.Fatalf("got zone %v ok=%v, want Residential", z, ok)
	}
}

func TestTryConsumeWallDemandNWCornerPlantOnly(t *testing.T) {
	s := NewEmpty()
	s.ApplyDemands(ShiftDemands{Plant: 1})

	z, ok := s.TryConsumeWallDemand(hex.Coord{Q: 5, R: 0}, hex.RotNW.TravelDir(), rand.New(rand.NewSource(1)))
	if !ok || z != ZonePlant {
		t.Fatalf("got zone %v ok=%v, want Plant", z, ok)
	}
}
