package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/charts"
	"github.com/jonas/reaktor-sim/internal/energy"
	"github.com/jonas/reaktor-sim/internal/finance"
	"github.com/jonas/reaktor-sim/internal/render"
	"github.com/jonas/reaktor-sim/internal/seedsearch"
	"github.com/jonas/reaktor-sim/internal/sim"
	"github.com/jonas/reaktor-sim/internal/stats"
)

type simExportOptions struct {
	tracePNGs    bool
	topSimCharts bool
}

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

func writeTopSims(dir string, scan seedsearch.ScanResult, card energy.Card, fin finance.Card, runs, keep int, export simExportOptions) error {
	if len(scan.Shifts) == 0 {
		return nil
	}
	last := scan.Shifts[len(scan.Shifts)-1]
	lastOutcomes, err := last.LoadOutcomes()
	if err != nil {
		return err
	}
	base := filepath.Join(dir, fmt.Sprintf("top_sims_schicht_%d", last.Shift))

	seen := make(map[string]struct{})
	for _, tbl := range topTables {
		rows := seedsearch.PickUniqueOutcomes(lastOutcomes, tbl.pick, keep, seen)
		if len(rows) == 0 {
			continue
		}
		tableDir := filepath.Join(base, tbl.slug)
		boardDirs := make(shiftBoardDirs)
		type pendingLink struct {
			outDir string
			shift  int
			prevFP string
		}
		var pending []pendingLink
		for _, row := range rows {
			chain, err := seedsearch.TraceChain(scan, row, card)
			if err != nil {
				return fmt.Errorf("%s: %w", tbl.slug, err)
			}
			for _, o := range chain {
				outDir := filepath.Join(tableDir, seedsearch.ShiftDirName(o))
				if err := writeSimExport(outDir, o, chain[:o.Shift], card, fin, runs, export); err != nil {
					return fmt.Errorf("%s %s: %w", tbl.slug, seedsearch.ShiftDirName(o), err)
				}
				boardDirs.register(o.Shift, o.BoardFingerprint, outDir)
				if o.Shift > 1 && o.PrevBoardFingerprint != "" {
					pending = append(pending, pendingLink{
						outDir: outDir,
						shift:  o.Shift,
						prevFP: o.PrevBoardFingerprint,
					})
				}
			}
		}
		for _, p := range pending {
			prevDir, ok := boardDirs.lookup(p.shift-1, p.prevFP)
			if !ok {
				continue
			}
			if err := linkToPrevShift(p.outDir, prevDir); err != nil {
				return fmt.Errorf("%s: vorschicht link (prev %s): %w", tbl.slug, p.prevFP, err)
			}
		}
	}
	return nil
}

func writeSimExport(outDir string, o seedsearch.Outcome, chainPrefix []seedsearch.Outcome, card energy.Card, fin finance.Card, runs int, export simExportOptions) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	state, err := board.FromFingerprint(o.BoardFingerprint)
	if err != nil {
		return fmt.Errorf("board: %w", err)
	}
	state.Damage = o.StartDamage
	state.EmitterDamage = o.StartEmitterDamage

	cfg := sim.DefaultConfig()
	cfg.EnergyCard = card
	cfg.FinanceCard = fin
	cfg.CriticalLimit = fin.CriticalLimit()
	cfg.Shift = o.Shift
	cfg.RandomShift = false
	cfg.ShiftDemands = o.StartDemands
	cfg.ReactorRepairBudget = o.ReactorRepairBudget
	cfg.RepairBudget = o.RepairBudget

	rng := rand.New(rand.NewSource(o.Seed))
	preview := state.Clone()
	preview.ApplyDemands(cfg.ShiftDemands)
	previewChips := sim.EmitterChips(preview, cfg, rng)

	traceCosts := state.PlayerCosts()
	traceMeta := render.TraceMeta{
		Shift: o.Shift,
		Costs: traceCosts,
	}

	var traceIndex []render.TraceIndexEntry
	if export.tracePNGs {
		traceRNG := rand.New(rand.NewSource(o.Seed + 1))
		runWriter, err := render.NewRunTraceStreamWriter(outDir, 1, traceMeta)
		if err != nil {
			return err
		}
		cfg.Trace = true
		cfg.TraceStep = runWriter.Record
		sim.RunTrace(state, traceRNG, cfg)
		if err := runWriter.Finish(); err != nil {
			return fmt.Errorf("trace run1: %w", err)
		}
		traceIndex = append(traceIndex, render.TraceIndexEntry{
			Run:   1,
			Kind:  render.TraceKindFirst,
			Steps: runWriter.Steps(),
			Dir:   runWriter.Dir(),
		})

		if o.TraceLoopRun > 0 {
			loopWriter, err := render.NewLoopTraceStreamWriter(outDir, 1, o.TraceLoopRun, traceMeta)
			if err != nil {
				return err
			}
			loopCfg := cfg
			loopCfg.TraceStep = loopWriter.Record
			loopRNG := rand.New(rand.NewSource(o.Seed + int64(o.TraceLoopRun)))
			sim.RunTrace(state, loopRNG, loopCfg)
			if err := loopWriter.Finish(); err != nil {
				return fmt.Errorf("trace loop1 (MC %d): %w", o.TraceLoopRun, err)
			}
			traceIndex = append(traceIndex, render.TraceIndexEntry{
				Run:           1,
				MonteCarloRun: o.TraceLoopRun,
				Kind:          render.TraceKindLoop,
				Steps:         loopWriter.Steps(),
				Dir:           loopWriter.Dir(),
			})
		}

		if o.TraceWinRun > 0 {
			winWriter, err := render.NewWinTraceStreamWriter(outDir, 1, o.TraceWinRun, traceMeta)
			if err != nil {
				return err
			}
			winCfg := cfg
			winCfg.TraceStep = winWriter.Record
			winRNG := rand.New(rand.NewSource(o.Seed + int64(o.TraceWinRun)))
			sim.RunTrace(state, winRNG, winCfg)
			if err := winWriter.Finish(); err != nil {
				return fmt.Errorf("trace win1 (MC %d): %w", o.TraceWinRun, err)
			}
			traceIndex = append(traceIndex, render.TraceIndexEntry{
				Run:           1,
				MonteCarloRun: o.TraceWinRun,
				Kind:          render.TraceKindWin,
				Steps:         winWriter.Steps(),
				Dir:           winWriter.Dir(),
			})
		}
	}

	if len(traceIndex) > 0 {
		if err := render.WriteTraceIndex(outDir, traceIndex); err != nil {
			return fmt.Errorf("trace index: %w", err)
		}
	}

	if err := render.WriteAll(preview, outDir, render.ChipView{Queue: previewChips}, render.BoardMeta{
		Seed:      o.Seed,
		PrevBoard: o.PrevBoardFingerprint,
	}); err != nil {
		return fmt.Errorf("render: %w", err)
	}

	if export.topSimCharts {
		results := sim.RunMonteCarlo(state, runs, o.Seed, cfg)
		report := stats.Build(state.PlayerCosts(), o.EndLeftover, results)
		if len(chainPrefix) > 1 {
			cm := seedsearch.CampaignMoneyFromChain(chainPrefix, fin)
			report.Campaign = &cm
		}
		if err := charts.WriteAll(report, outDir); err != nil {
			return fmt.Errorf("charts: %w", err)
		}
	}
	return nil
}

// shiftBoardDirs maps exported shift folders by board fingerprint (prev_board target).
type shiftBoardDirs map[int]map[string]string

func (m shiftBoardDirs) register(shift int, boardFP, dir string) {
	if m[shift] == nil {
		m[shift] = make(map[string]string)
	}
	m[shift][boardFP] = dir
}

func (m shiftBoardDirs) lookup(shift int, boardFP string) (string, bool) {
	dirs, ok := m[shift]
	if !ok {
		return "", false
	}
	dir, ok := dirs[boardFP]
	return dir, ok
}
