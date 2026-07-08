// Package charts renders balance reports as PNG images.
package charts

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/stats"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// WriteAll saves cost label and histogram charts to outDir.
func WriteAll(report stats.Report, outDir string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	if err := writeCostChart(report, filepath.Join(outDir, "kosten.png")); err != nil {
		return err
	}
	if err := writeHistogram(report.HeatAtTurbine, "Waerme an Turbine", filepath.Join(outDir, "waerme_turbine.png")); err != nil {
		return err
	}

	zones := []board.Zone{
		board.ZoneResidential,
		board.ZoneIndustry,
		board.ZoneRail,
		board.ZonePlant,
	}
	names := map[board.Zone]string{
		board.ZoneResidential: "wohnviertel",
		board.ZoneIndustry:    "industrie",
		board.ZoneRail:        "bahn",
		board.ZonePlant:       "reaktoreigenbedarf",
	}
	for _, z := range zones {
		h := report.ZoneHistograms[z]
		title := fmt.Sprintf("Spannung bei %s", z.String())
		path := filepath.Join(outDir, fmt.Sprintf("spannung_%s.png", names[z]))
		if err := writeHistogram(h, title, path); err != nil {
			return err
		}
	}
	return nil
}

func writeCostChart(report stats.Report, path string) error {
	p := plot.New()
	p.Title.Text = "Game-State Kosten"
	p.Y.Label.Text = "Geld"
	p.X.Label.Text = ""

	c := report.Costs
	values := plotter.Values{float64(c.Player1), float64(c.Player2)}
	bar, err := plotter.NewBarChart(values, vg.Points(40))
	if err != nil {
		return err
	}
	bar.Color = color.RGBA{R: 70, G: 130, B: 180, A: 255}
	p.Add(bar)
	p.NominalX("Spieler 1\n(Reaktor)", "Spieler 2\n(Stromnetz)")

	maxCost := c.Player1
	if c.Player2 > maxCost {
		maxCost = c.Player2
	}
	p.Y.Max = float64(maxCost) * 1.2
	if p.Y.Max < 5 {
		p.Y.Max = 5
	}

	sub := fmt.Sprintf("%d Simulationen | Kritische Masse: %.1f%%", report.Runs, report.CriticalFailRate*100)
	p.Title.Text = fmt.Sprintf("Kosten: %s (%s)", c.String(), sub)

	return p.Save(6*vg.Inch, 4*vg.Inch, path)
}

func writeHistogram(h stats.Histogram, title, path string) error {
	p := plot.New()
	p.Title.Text = title
	p.X.Label.Text = "Anzahl"
	p.Y.Label.Text = "Wahrscheinlichkeit"

	keys := stats.SortedKeys(h)
	if len(keys) == 0 {
		keys = []int{0}
		h = stats.Histogram{0: 1}
	}

	values := make(plotter.Values, len(keys))
	labels := make([]string, len(keys))
	for i, k := range keys {
		values[i] = h[k] * 100
		labels[i] = fmt.Sprintf("%d", k)
	}

	bar, err := plotter.NewBarChart(values, vg.Points(30))
	if err != nil {
		return err
	}
	bar.Color = color.RGBA{R: 60, G: 160, B: 90, A: 255}
	p.Add(bar)

	p.NominalX(labels...)
	p.Y.Label.Text = "Chance (%)"
	maxY := 0.0
	for _, v := range values {
		if float64(v) > maxY {
			maxY = float64(v)
		}
	}
	p.Y.Max = maxY * 1.15
	if p.Y.Max < 10 {
		p.Y.Max = 10
	}

	return p.Save(7*vg.Inch, 4*vg.Inch, path)
}
