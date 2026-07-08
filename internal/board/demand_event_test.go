package board

import "testing"

func TestBorderDemandEvent(t *testing.T) {
	got := BorderDemandEvent(ZoneIndustry)
	if got != "Rand-Bedarf Industrie erfuellt" {
		t.Fatalf("got %q", got)
	}
	z, ok := ZoneFromBorderDemandEvent(got)
	if !ok || z != ZoneIndustry {
		t.Fatalf("parse = %v %v", z, ok)
	}
}
