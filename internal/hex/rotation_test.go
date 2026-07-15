package hex_test

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/hex"
)

func TestRotationTravelDirs(t *testing.T) {
	cases := []struct {
		rot  hex.Rotation
		want int
	}{
		{hex.RotNW, 2},
		{hex.RotNE, 1},
		{hex.RotE, 0},
		{hex.RotSE, 5},
		{hex.RotSW, 4},
		{hex.RotW, 3},
	}
	for _, tc := range cases {
		if got := tc.rot.TravelDir(); got != tc.want {
			t.Fatalf("rotation %d travel dir: got %d want %d", tc.rot, got, tc.want)
		}
	}
}

func TestDisplayDir(t *testing.T) {
	cases := []struct {
		travel int
		want   int
	}{
		{hex.RotNW.TravelDir(), int(hex.RotNW)},
		{hex.RotNE.TravelDir(), int(hex.RotNE)},
		{hex.RotE.TravelDir(), int(hex.RotE)},
		{hex.RotSE.TravelDir(), int(hex.RotSE)},
		{hex.RotSW.TravelDir(), int(hex.RotSW)},
		{hex.RotW.TravelDir(), int(hex.RotW)},
	}
	for _, tc := range cases {
		if got := hex.DisplayDir(tc.travel); got != tc.want {
			t.Fatalf("DisplayDir(%d) = %d, want %d", tc.travel, got, tc.want)
		}
	}
}

func TestDisplayDirName(t *testing.T) {
	if got := hex.DisplayDirName(hex.RotE.TravelDir()); got != "E" {
		t.Fatalf("DisplayDirName(east) = %q, want E", got)
	}
	if got := hex.DisplayDirName(hex.RotNE.TravelDir()); got != "NE" {
		t.Fatalf("DisplayDirName(NE) = %q, want NE", got)
	}
}

func TestEmitterShotSoutheast(t *testing.T) {
	next := hex.EmitterShotTarget(hex.RotSE.TravelDir())
	if next != (hex.Coord{Q: 1, R: 3}) {
		t.Fatalf("SE from emitter: got (%d,%d) want (1,3)", next.Q, next.R)
	}
}

func TestRelayOrientation5PassesEastWest(t *testing.T) {
	r := hex.RotW // orientation 5: horizontal E-W line
	// Voltage traveling E (incoming from W).
	if got := r.WireOutgoing(hex.RotW.TravelDir()); got != hex.RotE.TravelDir() {
		t.Fatalf("E pass-through: got %s want E", hex.DisplayDirName(got))
	}
	// Voltage traveling W (incoming from E).
	if got := r.WireOutgoing(hex.RotE.TravelDir()); got != hex.RotW.TravelDir() {
		t.Fatalf("W pass-through: got %s want W", hex.DisplayDirName(got))
	}
}

func TestRelayOrientation0DeflectsWestToSouthwest(t *testing.T) {
	r := hex.RotNW // orientation 0: NW-SE line
	if got := r.WireOutgoing(hex.RotW.TravelDir()); got != hex.RotSW.TravelDir() {
		t.Fatalf("W -> SW: got %s want SW", hex.DisplayDirName(got))
	}
}

func TestRelayOrientation0PassesNorthwestSoutheast(t *testing.T) {
	r := hex.RotNW
	if got := r.WireOutgoing(hex.RotSE.TravelDir()); got != hex.RotNW.TravelDir() {
		t.Fatalf("SE pass-through: got %s want NW", hex.DisplayDirName(got))
	}
	if got := r.WireOutgoing(hex.RotNW.TravelDir()); got != hex.RotSE.TravelDir() {
		t.Fatalf("NW pass-through: got %s want SE", hex.DisplayDirName(got))
	}
}

func TestParallelToAxis(t *testing.T) {
	r := hex.RotNW
	if !r.ParallelToAxis(hex.RotSE.TravelDir()) {
		t.Fatal("SE should be parallel to NW-SE axis")
	}
	if r.ParallelToAxis(hex.RotE.TravelDir()) {
		t.Fatal("E should not be parallel to NW-SE axis")
	}
}
