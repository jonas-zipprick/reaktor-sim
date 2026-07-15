package graph_test

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/graph"
	"github.com/jonas/reaktor-sim/internal/hex"
)

func TestGasBoilerPotentialSixDirections(t *testing.T) {
	s := board.NewEmpty()
	pos := hex.Coord{Q: 2, R: 2}
	s.Tiles[pos.Q][pos.R] = field.NewTile(field.GasBoiler, 0, 0)

	potential := graph.BuildPotential(s)
	node, ok := potential.Nodes[pos]
	if !ok {
		t.Fatal("gas boiler missing from potential graph")
	}
	if len(node.Edges) != 6 {
		t.Fatalf("gas boiler should have 6 edges, got %d", len(node.Edges))
	}
}

func TestGasBoilerFlowShowsAllDiceEdges(t *testing.T) {
	s := board.NewEmpty()
	pos := hex.Coord{Q: 2, R: 2}
	s.Tiles[pos.Q][pos.R] = field.NewTile(field.GasBoiler, 0, 0)

	// Only two directions rolled, but graph should show all six dice edges.
	flow := graph.BuildFlow(s, []graph.InFlight{
		{Particle: graph.Heat, Pos: pos, Dir: hex.RotE.TravelDir()},
		{Particle: graph.Heat, Pos: pos, Dir: hex.RotNW.TravelDir()},
	})
	node := flow.Nodes[pos]
	if len(node.Edges) != 6 {
		t.Fatalf("gas with chips in flight should show 6 dice edges, got %d", len(node.Edges))
	}
}
