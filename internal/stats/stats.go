// Package stats aggregates Monte-Carlo results into histograms.
package stats

import (
	"sort"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/sim"
)

// Histogram maps an observed count to its empirical probability.
type Histogram map[int]float64

// Report contains all balance metrics for one board state.
type Report struct {
	Costs            board.PlayerCosts
	Runs             int
	HeatAtTurbine    Histogram
	ZoneHistograms   map[board.Zone]Histogram
	CriticalFailRate float64
	CriticalP1Rate   float64
	CriticalP2Rate   float64
}

// Build aggregates simulation results into histograms.
func Build(costs board.PlayerCosts, results []sim.Result) Report {
	runs := len(results)
	if runs == 0 {
		return Report{Costs: costs, ZoneHistograms: make(map[board.Zone]Histogram)}
	}

	heatCounts := make(map[int]int)
	zoneCounts := [4]map[int]int{}
	for i := range zoneCounts {
		zoneCounts[i] = make(map[int]int)
	}
	critical := 0
	criticalP1 := 0
	criticalP2 := 0

	for _, r := range results {
		heatCounts[r.HeatAtTurbine]++
		for z := board.Zone(0); int(z) < 4; z++ {
			zoneCounts[z][r.ZoneDeliveries[z]]++
		}
		if r.CriticalFailure {
			critical++
		}
		if r.CriticalP1 {
			criticalP1++
		}
		if r.CriticalP2 {
			criticalP2++
		}
	}

	report := Report{
		Costs:            costs,
		Runs:             runs,
		HeatAtTurbine:    normalize(heatCounts, runs),
		ZoneHistograms:   make(map[board.Zone]Histogram, 4),
		CriticalFailRate: float64(critical) / float64(runs),
		CriticalP1Rate:   float64(criticalP1) / float64(runs),
		CriticalP2Rate:   float64(criticalP2) / float64(runs),
	}
	for z := board.Zone(0); int(z) < 4; z++ {
		report.ZoneHistograms[z] = normalize(zoneCounts[z], runs)
	}
	return report
}

func normalize(counts map[int]int, total int) Histogram {
	h := make(Histogram, len(counts))
	for k, v := range counts {
		h[k] = float64(v) / float64(total)
	}
	return h
}

// SortedKeys returns histogram bucket keys in ascending order.
func SortedKeys(h Histogram) []int {
	keys := make([]int, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

// MaxKey returns the highest bucket key in a histogram.
func MaxKey(h Histogram) int {
	max := 0
	for k := range h {
		if k > max {
			max = k
		}
	}
	return max
}
