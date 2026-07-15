package board

import (
	"math/rand"
	"testing"

	"github.com/jonas/reaktor-sim/internal/hex"
)

func TestZonesForOuterWallHitSEFromBottomRow(t *testing.T) {
	zones := ZonesForOuterWallHit(hex.Coord{Q: 6, R: 4}, hex.RotSE.TravelDir())
	if len(zones) != 1 || zones[0] != ZoneRail {
		t.Fatalf("SE from (6,4) = %v, want [Bahn]", zones)
	}
}

func TestZonesForOuterWallHitSEFromRightColumn(t *testing.T) {
	zones := ZonesForOuterWallHit(hex.Coord{Q: 8, R: 2}, hex.RotE.TravelDir())
	if len(zones) != 1 || zones[0] != ZoneResidential {
		t.Fatalf("E from (8,2) = %v, want [Wohnviertel]", zones)
	}
}

func TestZonesForOuterWallHitNWTopLeftCorner(t *testing.T) {
	zones := ZonesForOuterWallHit(hex.Coord{Q: 6, R: 0}, hex.RotNW.TravelDir())
	if len(zones) != 1 || zones[0] != ZoneIndustry {
		t.Fatalf("NW from (6,0) = %v, want [Industrie]", zones)
	}
}

func TestTryConsumeWallDemandSEBottomRowRailOnly(t *testing.T) {
	s := NewEmpty()
	s.ApplyDemands(ShiftDemands{Residential: 1, Rail: 1})

	z, ok := s.TryConsumeWallDemand(hex.Coord{Q: 6, R: 4}, hex.RotSE.TravelDir(), rand.New(rand.NewSource(1)))
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

	_, ok := s.TryConsumeWallDemand(hex.Coord{Q: 6, R: 4}, hex.RotSE.TravelDir(), rand.New(rand.NewSource(1)))
	if ok {
		t.Fatal("expected no delivery when only residential demand remains for bottom wall")
	}
}

func TestTryConsumeWallDemandEastResidential(t *testing.T) {
	s := NewEmpty()
	s.ApplyDemands(ShiftDemands{Residential: 1})

	z, ok := s.TryConsumeWallDemand(hex.Coord{Q: 8, R: 2}, hex.RotE.TravelDir(), rand.New(rand.NewSource(1)))
	if !ok || z != ZoneResidential {
		t.Fatalf("got zone %v ok=%v, want Residential", z, ok)
	}
}

func TestBottomRowNeverWiresPlantDemand(t *testing.T) {
	c := hex.Coord{Q: 6, R: 4}
	for _, z := range ZonesOf(c) {
		if z == ZonePlant {
			t.Fatalf("(%d,%d) must not wire Reaktoreigenbedarf", c.Q, c.R)
		}
	}

	s := NewEmpty()
	s.ApplyDemands(ShiftDemands{Plant: 1, Rail: 1})
	if s.DemandAt(c, ZonePlant) != 0 {
		t.Fatalf("plant demand placed on bottom row at (%d,%d)", c.Q, c.R)
	}
	if s.DemandAt(c, ZoneRail) == 0 {
		t.Fatal("expected rail demand on bottom-left player-2 cell")
	}
}

func TestTopRowNeverWiresPlantDemand(t *testing.T) {
	c := hex.Coord{Q: 6, R: 0}
	for _, z := range ZonesOf(c) {
		if z == ZonePlant {
			t.Fatalf("(%d,%d) must not wire Reaktoreigenbedarf", c.Q, c.R)
		}
	}

	s := NewEmpty()
	s.ApplyDemands(ShiftDemands{Plant: 1, Industry: 1})
	if s.DemandAt(c, ZonePlant) != 0 {
		t.Fatalf("plant demand placed on top row at (%d,%d)", c.Q, c.R)
	}
	if s.DemandAt(c, ZoneIndustry) == 0 {
		t.Fatal("expected industry demand on top-left player-2 cell")
	}
	if s.DemandAt(hex.Coord{Q: hex.TurbineCol, R: hex.TurbineRow}, ZonePlant) == 0 {
		t.Fatal("expected plant demand on turbine")
	}
}

func TestTryConsumeWallDemandNWCornerIndustryOnly(t *testing.T) {
	s := NewEmpty()
	s.ApplyDemands(ShiftDemands{Plant: 1, Industry: 1})

	z, ok := s.TryConsumeWallDemand(hex.Coord{Q: 6, R: 0}, hex.RotNW.TravelDir(), rand.New(rand.NewSource(1)))
	if !ok || z != ZoneIndustry {
		t.Fatalf("got zone %v ok=%v, want Industry", z, ok)
	}
	if s.TotalDemand(ZonePlant) != 1 {
		t.Fatal("plant demand must not be consumed from top row")
	}
}

func TestTryConsumeWallDemandPlantWallBesideTurbine(t *testing.T) {
	cases := []struct {
		c   hex.Coord
		dir hex.Rotation
	}{
		{hex.Coord{Q: hex.TurbineCol, R: 1}, hex.RotNW},
		{hex.Coord{Q: hex.TurbineCol, R: 1}, hex.RotW},
		{hex.Coord{Q: hex.TurbineCol, R: 3}, hex.RotSW},
		{hex.Coord{Q: hex.TurbineCol, R: 3}, hex.RotW},
		{hex.Coord{Q: hex.TurbineCol, R: hex.TurbineRow}, hex.RotNW},
		{hex.Coord{Q: hex.TurbineCol, R: hex.TurbineRow}, hex.RotSW},
	}
	for _, tc := range cases {
		s := NewEmpty()
		s.ApplyDemands(ShiftDemands{Plant: 1})
		z, ok := s.TryConsumeWallDemand(tc.c, tc.dir.TravelDir(), rand.New(rand.NewSource(1)))
		if !ok || z != ZonePlant {
			t.Fatalf("(%d,%d) %s got zone %v ok=%v, want Plant", tc.c.Q, tc.c.R, tc.dir, z, ok)
		}
	}
}
