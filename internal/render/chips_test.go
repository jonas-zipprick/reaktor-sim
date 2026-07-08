package render

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestLooseCountsAtOnAndIncoming(t *testing.T) {
	c := hex.Coord{Q: 2, R: 2}
	view := ChipView{
		Queue: []sim.Chip{
			{Type: sim.ChipHeat, Pos: c, Dir: 0},
			{Type: sim.ChipNeutron, Pos: hex.Coord{Q: 1, R: 2}, Dir: 0},
		},
		Active: &sim.Chip{Type: sim.ChipVoltage, Pos: hex.Coord{Q: 3, R: 2}, Dir: 3},
	}
	counts := LooseCountsAt(view, c)
	if counts.OnHeat != 1 || counts.InNeutron != 1 || counts.InVoltage != 1 {
		t.Fatalf("on/in: %+v", counts)
	}
	if got := LooseLabel(counts); got != "+1W>1N>1S" {
		t.Fatalf("label = %q", got)
	}
}

func TestLooseLabelIncomingOnly(t *testing.T) {
	c := hex.Coord{Q: 2, R: 2}
	active := sim.Chip{Type: sim.ChipHeat, Pos: hex.Coord{Q: 1, R: 2}, Dir: 0}
	counts := LooseCountsAt(ChipView{Active: &active}, c)
	if got := LooseLabel(counts); got != ">1W" {
		t.Fatalf("label = %q", got)
	}
}

func TestLooseCountsAtDoesNotDoubleCountActiveInQueue(t *testing.T) {
	turbine := hex.Coord{Q: hex.TurbineCol, R: hex.TurbineRow}
	chip := sim.Chip{Type: sim.ChipVoltage, Pos: hex.Coord{Q: 5, R: 2}, Dir: hex.RotNW.TravelDir()}
	view := ChipView{
		Queue:  []sim.Chip{chip},
		Active: &chip,
	}
	counts := LooseCountsAt(view, turbine)
	if counts.InVoltage != 1 {
		t.Fatalf("incoming voltage = %d, want 1 (not double-counted)", counts.InVoltage)
	}
	if got := LooseLabel(counts); got != ">1S" {
		t.Fatalf("label = %q, want >1S", got)
	}
}
