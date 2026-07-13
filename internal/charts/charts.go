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
	if err := writeHistogram(report.HeatAtTurbine, "Waerme an Turbine", "Anzahl", filepath.Join(outDir, "waerme_turbine.png")); err != nil {
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
		if err := writeHistogram(h, title, "Anzahl", path); err != nil {
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
	fieldVals := plotter.Values{float64(c.Player1), float64(c.Player2)}
	var rebuildVals plotter.Values
	if report.Campaign != nil {
		rebuildVals = plotter.Values{report.Campaign.RebuildF[0], report.Campaign.RebuildF[1]}
	}
	repairVals := plotter.Values{report.AvgRepairF[0], report.AvgRepairF[1]}
	savedVals := plotter.Values{report.AvgSavedF[0], report.AvgSavedF[1]}

	fieldBar, err := plotter.NewBarChart(fieldVals, vg.Points(40))
	if err != nil {
		return err
	}
	fieldBar.Color = color.RGBA{R: 70, G: 130, B: 180, A: 255}
	fieldBar.LineStyle.Width = 0

	var rebuildBar *plotter.BarChart
	if report.Campaign != nil {
		rebuildBar, err = plotter.NewBarChart(rebuildVals, vg.Points(40))
		if err != nil {
			return err
		}
		rebuildBar.Color = color.RGBA{R: 160, G: 110, B: 70, A: 255}
		rebuildBar.LineStyle.Width = 0
		rebuildBar.StackOn(fieldBar)
	}

	repairBar, err := plotter.NewBarChart(repairVals, vg.Points(40))
	if err != nil {
		return err
	}
	repairBar.Color = color.RGBA{R: 220, G: 140, B: 60, A: 255}
	repairBar.LineStyle.Width = 0
	if rebuildBar != nil {
		repairBar.StackOn(rebuildBar)
	} else {
		repairBar.StackOn(fieldBar)
	}

	savedBar, err := plotter.NewBarChart(savedVals, vg.Points(40))
	if err != nil {
		return err
	}
	savedBar.Color = color.RGBA{R: 90, G: 170, B: 90, A: 255}
	savedBar.LineStyle.Width = 0
	savedBar.StackOn(repairBar)

	if rebuildBar != nil {
		p.Add(fieldBar, rebuildBar, repairBar, savedBar)
		p.Legend.Add("Feldkosten (Brett)", fieldBar)
		p.Legend.Add("Umbau", rebuildBar)
	} else {
		p.Add(fieldBar, repairBar, savedBar)
		p.Legend.Add("Feldkosten", fieldBar)
	}
	p.Legend.Add("Reparatur", repairBar)
	p.Legend.Add("Gespart", savedBar)
	p.Legend.Top = true
	p.NominalX("Spieler 1\n(Reaktor)", "Spieler 2\n(Stromnetz)")

	maxCost := playerStackTotal(c.Player1, 0, report.Campaign, report.AvgRepairF[0], report.AvgSavedF[0])
	if total := playerStackTotal(c.Player2, 1, report.Campaign, report.AvgRepairF[1], report.AvgSavedF[1]); total > maxCost {
		maxCost = total
	}
	if report.Campaign != nil {
		if report.Campaign.MonthBudget[0] > maxCost {
			maxCost = report.Campaign.MonthBudget[0]
		}
		if report.Campaign.MonthBudget[1] > maxCost {
			maxCost = report.Campaign.MonthBudget[1]
		}
	}
	p.Y.Max = float64(maxCost) * 1.15
	if p.Y.Max < 5 {
		p.Y.Max = 5
	}

	sub := fmt.Sprintf("%d Simulationen | Kritische Masse: %.1f%%", report.Runs, report.CriticalFailRate*100)
	if report.Campaign != nil && report.Campaign.Shifts > 1 {
		p.Title.Text = fmt.Sprintf("Monatsbudget %d/%d Geld (%d Schichten) | Reaktor %d+%.0f+%.1f+%.1f | Stromnetz %d+%.0f+%.1f+%.1f (%s)",
			report.Campaign.MonthBudget[0], report.Campaign.MonthBudget[1], report.Campaign.Shifts,
			c.Player1, report.Campaign.RebuildF[0], report.AvgRepairF[0], report.AvgSavedF[0],
			c.Player2, report.Campaign.RebuildF[1], report.AvgRepairF[1], report.AvgSavedF[1], sub)
	} else {
		p.Title.Text = fmt.Sprintf("Kosten: Reaktor %d+%.1f+%.1f | Stromnetz %d+%.1f+%.1f (%s)",
			c.Player1, report.AvgRepairF[0], report.AvgSavedF[0],
			c.Player2, report.AvgRepairF[1], report.AvgSavedF[1], sub)
	}

	return p.Save(6*vg.Inch, 4*vg.Inch, path)
}

func playerStackTotal(field, player int, campaign *stats.CampaignMoney, repair, saved float64) int {
	total := field + int(repair+0.5) + int(saved+0.5)
	if campaign != nil {
		total += int(campaign.RebuildF[player] + 0.5)
	}
	return total
}

func writeHistogram(h stats.Histogram, title, xLabel, path string) error {
	p := plot.New()
	p.Title.Text = title
	p.X.Label.Text = xLabel
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

	width := 7 * vg.Inch
	if len(keys) > 14 {
		width = vg.Inch * vg.Length(4+float64(len(keys))*0.35)
		if width > 20*vg.Inch {
			width = 20 * vg.Inch
		}
	}

	return p.Save(width, 4*vg.Inch, path)
}
