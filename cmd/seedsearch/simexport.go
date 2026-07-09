package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/charts"
	"github.com/jonas/reaktor-sim/internal/energy"
	"github.com/jonas/reaktor-sim/internal/render"
	"github.com/jonas/reaktor-sim/internal/seedsearch"
	"github.com/jonas/reaktor-sim/internal/sim"
	"github.com/jonas/reaktor-sim/internal/stats"
)

type topTable struct {
	slug string
	pick func([]seedsearch.Outcome, int) []seedsearch.Outcome
}

var topTables = []topTable{
	{"top_wins", seedsearch.TopWins},
	{"top_all_demands_no_damage", seedsearch.TopAllDemandsNoDamage},
	{"top_max1_demand_no_damage", seedsearch.TopMax1DemandNoDamage},
	{"top_max1_demand_max1_damage", seedsearch.TopMax1DemandMax1Damage},
	{"top_loops", seedsearch.TopLoops},
}

func writeTopSims(dir string, scan seedsearch.ScanResult, card energy.Card, runs, keep int) error {
	if len(scan.Shifts) == 0 {
		return nil
	}
	last := scan.Shifts[len(scan.Shifts)-1]
	base := filepath.Join(dir, fmt.Sprintf("top_sims_schicht_%d", last.Shift))

	seen := make(map[string]struct{})
	for _, tbl := range topTables {
		rows := seedsearch.PickUniqueOutcomes(last.Outcomes, tbl.pick, keep, seen)
		if len(rows) == 0 {
			continue
		}
		tableDir := filepath.Join(base, tbl.slug)
		for _, row := range rows {
			chain, err := seedsearch.TraceChain(scan, row, card)
			if err != nil {
				return fmt.Errorf("%s: %w", tbl.slug, err)
			}
			for _, o := range chain {
				outDir := filepath.Join(tableDir, seedsearch.ShiftDirName(o))
				if err := writeSimExport(outDir, o, runs); err != nil {
					return fmt.Errorf("%s %s: %w", tbl.slug, seedsearch.ShiftDirName(o), err)
				}
			}
		}
	}
	return nil
}

func writeSimExport(outDir string, o seedsearch.Outcome, runs int) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	state, err := board.FromFingerprint(o.BoardFingerprint)
	if err != nil {
		return fmt.Errorf("board: %w", err)
	}
	state.Damage = o.StartDamage

	cfg := sim.DefaultConfig()
	cfg.EnergyCard = energy.Card{}
	cfg.Shift = o.Shift
	cfg.RandomShift = false
	cfg.ShiftDemands = o.StartDemands
	cfg.RepairBudget = o.RepairBudget

	rng := rand.New(rand.NewSource(o.Seed))
	preview := state.Clone()
	preview.ApplyDemands(cfg.ShiftDemands)
	previewChips := sim.EmitterChips(preview, cfg, rng)

	traceCosts := state.PlayerCosts()
	traceRNG := rand.New(rand.NewSource(o.Seed + 1))
	res, snaps := sim.RunTrace(state, traceRNG, cfg)
	meta := render.TraceMeta{
		Shift: res.Shift,
		Costs: traceCosts,
	}
	var traceIndex []render.TraceIndexEntry
	if err := render.WriteRunTrace(1, meta, snaps, outDir); err != nil {
		return fmt.Errorf("trace run1: %w", err)
	}
	runDir, _ := filepath.Abs(filepath.Join(outDir, "run1"))
	traceIndex = append(traceIndex, render.TraceIndexEntry{
		Run:   1,
		Kind:  render.TraceKindFirst,
		Steps: len(snaps),
		Dir:   runDir,
	})

	results := sim.RunMonteCarlo(state, runs, o.Seed, cfg)
	report := stats.Build(state.PlayerCosts(), results)

	for loopNum, mcRun := range sim.LoopTraceRunIndices(results, 1) {
		loopRNG := rand.New(rand.NewSource(o.Seed + int64(mcRun)))
		loopRes, loopSnaps := sim.RunTrace(state, loopRNG, cfg)
		loopMeta := render.TraceMeta{
			Shift: loopRes.Shift,
			Costs: traceCosts,
		}
		seq := loopNum + 1
		if err := render.WriteLoopTrace(seq, mcRun, loopMeta, loopSnaps, outDir); err != nil {
			return fmt.Errorf("trace loop%d (MC %d): %w", seq, mcRun, err)
		}
		loopDir, _ := filepath.Abs(filepath.Join(outDir, fmt.Sprintf("loop%d", seq)))
		traceIndex = append(traceIndex, render.TraceIndexEntry{
			Run:           seq,
			MonteCarloRun: mcRun,
			Kind:          render.TraceKindLoop,
			Steps:         len(loopSnaps),
			Dir:           loopDir,
		})
	}

	for winNum, mcRun := range sim.WinTraceRunIndices(results, 1) {
		winRNG := rand.New(rand.NewSource(o.Seed + int64(mcRun)))
		winRes, winSnaps := sim.RunTrace(state, winRNG, cfg)
		winMeta := render.TraceMeta{
			Shift: winRes.Shift,
			Costs: traceCosts,
		}
		seq := winNum + 1
		if err := render.WriteWinTrace(seq, mcRun, winMeta, winSnaps, outDir); err != nil {
			return fmt.Errorf("trace win%d (MC %d): %w", seq, mcRun, err)
		}
		winDir, _ := filepath.Abs(filepath.Join(outDir, fmt.Sprintf("win%d", seq)))
		traceIndex = append(traceIndex, render.TraceIndexEntry{
			Run:           seq,
			MonteCarloRun: mcRun,
			Kind:          render.TraceKindWin,
			Steps:         len(winSnaps),
			Dir:           winDir,
		})
	}
	if err := render.WriteTraceIndex(outDir, traceIndex); err != nil {
		return fmt.Errorf("trace index: %w", err)
	}

	if err := render.WriteAll(preview, outDir, render.ChipView{Queue: previewChips}); err != nil {
		return fmt.Errorf("render: %w", err)
	}
	if err := charts.WriteAll(report, outDir); err != nil {
		return fmt.Errorf("charts: %w", err)
	}
	return nil
}
