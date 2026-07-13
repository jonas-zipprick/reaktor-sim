package render

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
)

func TestBuildBoardYAMLCosts(t *testing.T) {
	s := board.NewEmpty()
	s.ApplyDemands(board.ShiftDemands{Plant: 1})
	doc := buildBoardYAML(s, BoardMeta{})
	if doc.Costs.Total != 0 {
		t.Fatalf("costs = %+v", doc.Costs)
	}
	if len(doc.Demands) != 1 || doc.Demands[0].Zone != "Reaktoreigenbedarf" {
		t.Fatalf("demands = %+v", doc.Demands)
	}
}
