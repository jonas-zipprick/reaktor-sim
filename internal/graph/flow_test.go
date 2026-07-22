package graph_test

import (
	"testing"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/graph"
	"github.com/jonas/reaktor-sim/internal/hex"
)

func TestBuildFlowStartOnlyEmitterEdge(t *testing.T) {
	s := board.NewEmpty()
	s.Tiles[2][0] = field.NewTile(field.GasBoiler, 0, 0)
	s.Tiles[3][0] = field.NewTile(field.CoalChamber, 0, 0)

	flow := graph.BuildFlow(s, []graph.InFlight{{
		Particle: graph.Heat,
		Pos:      hex.Coord{Q: hex.EmitterCol, R: hex.EmitterRow},
		Dir:      hex.RotSE.TravelDir(),
	}})

	emitter := flow.Nodes[hex.Coord{Q: hex.EmitterCol, R: hex.EmitterRow}]
	if len(emitter.Edges) != 1 {
		t.Fatalf("emitter should have 1 edge, got %d", len(emitter.Edges))
	}
	if emitter.Edges[0].Heat != 1 {
		t.Fatalf("emitter edge should be 100%% heat, got %.2f", emitter.Edges[0].Heat)
	}
	want := hex.Coord{Q: 0, R: 3}
	if emitter.Edges[0].To != want {
		t.Fatalf("SE shot should target (%d,%d), got (%d,%d)", want.Q, want.R, emitter.Edges[0].To.Q, emitter.Edges[0].To.R)
	}

	for _, c := range []hex.Coord{{Q: 2, R: 0}, {Q: 3, R: 0}} {
		node, ok := flow.Nodes[c]
		if !ok {
			continue
		}
		for _, e := range node.Edges {
			if e.Heat > 0 || e.Neutron > 0 || e.Voltage > 0 {
				t.Fatalf("node (%d,%d) should not emit at start", c.Q, c.R)
			}
		}
	}
}

func TestBuildPotentialIncludesReactiveFields(t *testing.T) {
	s := board.NewEmpty()
	s.Tiles[1][2] = field.NewTile(field.GasBoiler, 0, 0)

	potential := graph.BuildPotential(s)
	gas := potential.Nodes[hex.Coord{Q: 1, R: 2}]
	if len(gas.Edges) == 0 {
		t.Fatal("potential graph should show reactive edges for gas boiler")
	}
}

func TestBuildPotentialEmitterThreeDirections(t *testing.T) {
	potential := graph.BuildPotential(board.NewEmpty())
	emitter := potential.Nodes[hex.Coord{Q: hex.EmitterCol, R: hex.EmitterRow}]
	if len(emitter.Edges) != 3 {
		t.Fatalf("emitter potential should have 3 edges, got %d", len(emitter.Edges))
	}
	sum := 0.0
	for _, e := range emitter.Edges {
		sum += e.Heat
	}
	if sum < 0.99 || sum > 1.01 {
		t.Fatalf("emitter heat should sum to 1, got %.2f", sum)
	}
}
