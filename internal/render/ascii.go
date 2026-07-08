package render

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/graph"
	"github.com/jonas/reaktor-sim/internal/hex"
)

// WriteBoardYAML saves structured board state for downstream tooling.
func WriteBoardYAML(state *board.State, path string) error {
	return writeYAML(path, buildBoardYAML(state))
}

// WriteGraphText saves a textual edge list of the flow graph.
func WriteGraphText(state *board.State, g *graph.Graph, path string) error {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Graph (%d Knoten)\n\n", len(g.Nodes)))
	b.WriteString("Kanten (von -> nach: Waerme/Neutron/Spannung):\n")

	coords := make([]hex.Coord, 0, len(g.Nodes))
	for c := range g.Nodes {
		coords = append(coords, c)
	}
	sortCoords(coords)

	for _, c := range coords {
		node := g.Nodes[c]
		if len(node.Edges) == 0 {
			continue
		}
		b.WriteString(fmt.Sprintf("(%d,%d) %s", c.Q, c.R, Label(state, c)))
		if ch := ChargeLabel(state.Tiles[c.Q][c.R]); ch != "" {
			b.WriteString(" Ladung=" + ch)
		}
		b.WriteString(":\n")
		for _, e := range node.Edges {
			if e.Heat < minEdgeProb && e.Neutron < minEdgeProb && e.Voltage < minEdgeProb {
				continue
			}
			b.WriteString(fmt.Sprintf("  -> (%d,%d)  H:%.2f  N:%.2f  V:%.2f\n",
				e.To.Q, e.To.R, e.Heat, e.Neutron, e.Voltage))
		}
	}

	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func sortCoords(coords []hex.Coord) {
	for i := 0; i < len(coords); i++ {
		for j := i + 1; j < len(coords); j++ {
			if coords[j].R < coords[i].R || (coords[j].R == coords[i].R && coords[j].Q < coords[i].Q) {
				coords[i], coords[j] = coords[j], coords[i]
			}
		}
	}
}

// WriteAll saves board renderings to outDir using spielfeld-<fingerprint> filenames.
func WriteAll(state *board.State, outDir string, view ChipView) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	fp := board.Fingerprint(state)
	base := "spielfeld-" + fp
	files := []struct {
		fn   func() error
		name string
	}{
		{func() error { return WriteBoardPNG(state, filepath.Join(outDir, base+".png"), view) }, base + ".png"},
		{func() error { return WriteBoardYAML(state, filepath.Join(outDir, base+".yaml")) }, base + ".yaml"},
	}
	for _, f := range files {
		if err := f.fn(); err != nil {
			return fmt.Errorf("%s: %w", f.name, err)
		}
	}
	return nil
}
