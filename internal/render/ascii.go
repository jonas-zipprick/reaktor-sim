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

// WriteBoardText saves an ASCII grid view of the board.
func WriteBoardText(state *board.State, path string) error {
	var b strings.Builder
	b.WriteString("Spielfeld (Spalten 1-9, Zeilen 1-3)\n\n")
	b.WriteString("      1  2  3  4  5  |  6  7  8  9\n")

	for r := 0; r < hex.Rows; r++ {
		if r&1 == 0 {
			b.WriteString("   ")
		}
		b.WriteString(fmt.Sprintf("r%d ", r+1))
		for q := 0; q < hex.Cols; q++ {
			c := hex.Coord{Q: q, R: r}
			sep := " "
			if q == hex.Player2MinCol {
				sep = "|"
			}
			if !c.Valid() {
				b.WriteString(sep + "  ")
				continue
			}
			sym := Label(state, c)
			if sym == "" {
				b.WriteString(sep + "  ")
				continue
			}
			if ch := ChargeLabel(state.Tiles[c.Q][c.R]); ch != "" {
				sym += ":" + ch
			}
			b.WriteString(sep + sym + " ")
		}
		b.WriteString("\n")
	}

	b.WriteString(fmt.Sprintf("\nKosten: %s (gesamt %d Geld)\n", state.PlayerCosts().String(), state.TotalCost()))
	if lines := DemandSummaryLines(state); len(lines) > 0 {
		b.WriteString("\nVerbleibender Bedarf (Rand):\n")
		for _, line := range lines {
			b.WriteString("  " + line + "\n")
		}
	}
	b.WriteString("\n")
	for _, line := range Legend() {
		b.WriteString(line + "\n")
	}
	b.WriteString("Rand-Bedarf ausserhalb: I oben  W rechts  b unten  R oben (Turbine)\n")

	return os.WriteFile(path, []byte(b.String()), 0o644)
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

// WriteAll saves board renderings to outDir.
func WriteAll(state *board.State, outDir string, view ChipView) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	files := []struct {
		fn   func() error
		name string
	}{
		{func() error { return WriteBoardPNG(state, filepath.Join(outDir, "spielfeld.png"), view) }, "spielfeld.png"},
		{func() error { return WriteBoardText(state, filepath.Join(outDir, "spielfeld.txt")) }, "spielfeld.txt"},
	}
	for _, f := range files {
		if err := f.fn(); err != nil {
			return fmt.Errorf("%s: %w", f.name, err)
		}
	}
	return nil
}
