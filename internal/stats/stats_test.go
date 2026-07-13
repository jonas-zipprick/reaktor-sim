package stats

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/sim"
)

func TestBuildAvgRepair(t *testing.T) {
	costs := board.PlayerCosts{Player1: 10, Player2: 20}
	results := []sim.Result{
		{ReactorRepairSpent: 2, RepairSpent: 1},
		{ReactorRepairSpent: 0, RepairSpent: 3},
	}
	report := Build(costs, board.PlayerLeftover{Player1: 5, Player2: 4}, results)
	if report.AvgRepairF[0] != 1 {
		t.Fatalf("avg repair P1 = %v, want 1", report.AvgRepairF[0])
	}
	if report.AvgRepairF[1] != 2 {
		t.Fatalf("avg repair P2 = %v, want 2", report.AvgRepairF[1])
	}
	if report.AvgSavedF[0] != 4 {
		t.Fatalf("avg saved P1 = %v, want 4", report.AvgSavedF[0])
	}
	if report.AvgSavedF[1] != 2 {
		t.Fatalf("avg saved P2 = %v, want 2", report.AvgSavedF[1])
	}
}
