package energy_test

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/energy"
)

func TestDefaultCardShift1NotAllOnes(t *testing.T) {
	c := energy.DefaultCard()
	d := c.ShiftDemands(1)
	if d.Industry != 1 || d.Residential != 1 || d.Rail != 0 || d.Plant != 1 {
		t.Fatalf("eroeffnungsfeier shift 1 = %+v, want I1 W1 b0 R1", d)
	}
}

func TestNetzoptimierungShift1(t *testing.T) {
	c, ok := energy.ByID("netzoptimierung")
	if !ok {
		t.Fatal("card not found")
	}
	d := c.ShiftDemands(1)
	if d.Industry != 0 || d.Residential != 1 || d.Rail != 0 || d.Plant != 1 {
		t.Fatalf("netzoptimierung shift 1 = %+v, want I0 W1 b0 R1", d)
	}
}

func TestEroeffnungsfeierShift3(t *testing.T) {
	c := energy.DefaultCard()
	d := c.ShiftDemands(3)
	if d.Industry != 2 || d.Residential != 1 || d.Rail != 0 || d.Plant != 1 {
		t.Fatalf("eroeffnungsfeier shift 3 = %+v, want I2 W1 b0 R1", d)
	}
}

func TestSchturmowShift5HighDemand(t *testing.T) {
	c, ok := energy.ByID("schturmowschtschina")
	if !ok {
		t.Fatal("card not found")
	}
	d := c.ShiftDemands(5)
	if d.Industry != 7 || d.Plant != 2 {
		t.Fatalf("shift 5 = %+v", d)
	}
}

func TestByIDUnknown(t *testing.T) {
	if _, ok := energy.ByID("missing"); ok {
		t.Fatal("expected missing card")
	}
}

func TestEveryShiftHasValidDemands(t *testing.T) {
	for _, c := range energy.Cards {
		for shift := 1; shift <= 5; shift++ {
			d := c.ShiftDemands(shift)
			if d.Industry < 0 || d.Residential < 0 || d.Rail < 0 || d.Plant < 0 {
				t.Fatalf("%s shift %d negative demand: %+v", c.ID, shift, d)
			}
			_ = board.ShiftDemands(d)
		}
	}
}
