package charts

import (
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/jonas/reaktor-sim/internal/seedsearch"
	"github.com/jonas/reaktor-sim/internal/stats"
)

// WriteSeedsearchCharts saves distribution histograms per seed.
func WriteSeedsearchCharts(outDir string, outcomes []seedsearch.Outcome, runs int) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	allSeedMetrics := []seedsearchMetric{
		{
			title:    "Verteilung uebrige Bedarfe je Seed",
			filename: "verteilung_end_bedarf.png",
			xLabel:   "Bedarf gesamt",
			bucket:   func(o seedsearch.Outcome) int { return o.TotalMedianEndDemand() },
		},
		{
			title:    "Verteilung Schaden je Seed",
			filename: "verteilung_end_schaden.png",
			xLabel:   "Schaden gesamt",
			bucket:   func(o seedsearch.Outcome) int { return o.TotalMedianEndDamage() },
		},
	}
	for _, m := range allSeedMetrics {
		h := outcomeHistogram(outcomes, m.bucket)
		title := fmt.Sprintf("%s (%d Seeds, %d Laeufe je Seed)", m.title, len(outcomes), runs)
		path := filepath.Join(outDir, m.filename)
		if err := writeHistogram(h, title, m.xLabel, path); err != nil {
			return fmt.Errorf("%s: %w", m.filename, err)
		}
	}

	winning := seedsearch.WinningOnly(outcomes)

	metrics := []seedsearchMetric{
		{
			title:    "Verteilung ø Schritte je Seed",
			filename: "verteilung_schritte.png",
			xLabel:   "ø Schritte",
			bucket:   func(o seedsearch.Outcome) int { return int(math.Round(o.AvgSteps)) },
		},
		{
			title:    "Verteilung Wins je Seed (alle Bedarfe erfuellt)",
			filename: "verteilung_wins.png",
			xLabel:   "Wins",
			bucket:   func(o seedsearch.Outcome) int { return o.Wins },
		},
		{
			title:    "Verteilung Treffer je Seed (alle Bedarfe, kein Schaden)",
			filename: "verteilung_all_demands_no_damage.png",
			xLabel:   "Treffer",
			bucket:   func(o seedsearch.Outcome) int { return o.AllDemandsNoDamage },
		},
		{
			title:    "Verteilung Treffer je Seed (max. 1 Bedarf, kein Schaden)",
			filename: "verteilung_max1_demand_no_damage.png",
			xLabel:   "Treffer",
			bucket:   func(o seedsearch.Outcome) int { return o.Max1DemandNoDamage },
		},
		{
			title:    "Verteilung Treffer je Seed (max. 1 Bedarf, max. 1 Schaden)",
			filename: "verteilung_max1_demand_max1_damage.png",
			xLabel:   "Treffer",
			bucket:   func(o seedsearch.Outcome) int { return o.Max1DemandMax1Damage },
		},
		{
			title:    "Verteilung Loops je Seed (Schrittlimit)",
			filename: "verteilung_loops.png",
			xLabel:   "Loops",
			bucket:   func(o seedsearch.Outcome) int { return o.Loops },
		},
		{
			title:    "Verteilung Kritisch P1 je Seed",
			filename: "verteilung_kritisch_p1.png",
			xLabel:   "Kritisch P1",
			bucket:   func(o seedsearch.Outcome) int { return o.CriticalP1 },
		},
		{
			title:    "Verteilung Kritisch P2 je Seed",
			filename: "verteilung_kritisch_p2.png",
			xLabel:   "Kritisch P2",
			bucket:   func(o seedsearch.Outcome) int { return o.CriticalP2 },
		},
		{
			title:    "Verteilung Brettkosten Reaktor je Seed",
			filename: "verteilung_kosten_p1.png",
			xLabel:   "Geld P1",
			bucket:   func(o seedsearch.Outcome) int { return o.BoardCosts.Player1 },
		},
		{
			title:    "Verteilung Brettkosten Stromnetz je Seed",
			filename: "verteilung_kosten_p2.png",
			xLabel:   "Geld P2",
			bucket:   func(o seedsearch.Outcome) int { return o.BoardCosts.Player2 },
		},
	}

	seedCount := len(winning)
	for _, m := range metrics {
		h := outcomeHistogram(winning, m.bucket)
		title := fmt.Sprintf("%s (%d gewinnende Seeds, %d Laeufe je Seed)", m.title, seedCount, runs)
		path := filepath.Join(outDir, m.filename)
		if err := writeHistogram(h, title, m.xLabel, path); err != nil {
			return fmt.Errorf("%s: %w", m.filename, err)
		}
	}
	return nil
}

type seedsearchMetric struct {
	title    string
	filename string
	xLabel   string
	bucket   func(seedsearch.Outcome) int
}

func outcomeHistogram(outcomes []seedsearch.Outcome, bucket func(seedsearch.Outcome) int) stats.Histogram {
	counts := make(map[int]int)
	for _, o := range outcomes {
		counts[bucket(o)]++
	}
	total := len(outcomes)
	if total == 0 {
		return stats.Histogram{0: 1}
	}
	h := make(stats.Histogram, len(counts))
	for k, v := range counts {
		h[k] = float64(v) / float64(total)
	}
	return h
}
