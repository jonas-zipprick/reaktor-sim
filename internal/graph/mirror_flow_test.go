package graph_test

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/graph"
	"github.com/jonas/reaktor-sim/internal/hex"
)

func TestBuildFlowMirrorSingleReflectionEdge(t *testing.T) {
	mirror := hex.Coord{Q: 2, R: 1}
	from := hex.Coord{Q: 1, R: 1}

	s := board.NewEmpty()
	tile := field.NewTile(field.Mirror, hex.RotNW, 0)
	s.Tiles[mirror.Q][mirror.R] = tile

	incoming := hex.RotE.TravelDir() // from (1,1) into mirror at (2,1)
	outDir := tile.Orientation.WireOutgoing(hex.RotW.TravelDir())

	flow := graph.BuildFlow(s, []graph.InFlight{
		{Particle: graph.Heat, Pos: from, Dir: incoming},
		// Chip waiting on mirror must not add a second outgoing edge.
		{Particle: graph.Heat, Pos: mirror, Dir: hex.RotSE.TravelDir()},
	})

	node := flow.Nodes[mirror]
	if len(node.Edges) != 1 {
		t.Fatalf("mirror should have 1 reflection edge, got %d", len(node.Edges))
	}
	want := mirror.Neighbor(outDir)
	if node.Edges[0].To != want {
		t.Fatalf("reflection edge to %v, want %v", node.Edges[0].To, want)
	}
	if node.Edges[0].Heat != 1 {
		t.Fatalf("reflection edge heat = %.2f, want 1", node.Edges[0].Heat)
	}
}
