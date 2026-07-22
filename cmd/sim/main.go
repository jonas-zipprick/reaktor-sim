package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/charts"
	"github.com/jonas/reaktor-sim/internal/energy"
	"github.com/jonas/reaktor-sim/internal/finance"
	"github.com/jonas/reaktor-sim/internal/render"
	"github.com/jonas/reaktor-sim/internal/rules"
	"github.com/jonas/reaktor-sim/internal/sim"
	"github.com/jonas/reaktor-sim/internal/stats"
)

func main() {
	runs := flag.Int("runs", 20, "Anzahl Monte-Carlo-Durchlaeufe")
	seed := flag.Int64("seed", 0, "Zufallsseed (0 = aktuelle Zeit)")
	outDir := flag.String("out", "output", "Ausgabeverzeichnis für Charts")
	traceFirst := flag.Int("trace-first", -1, "Erste n Laeufe aufzeichnen (-1 = aus, 0 = alle -runs)")
	traceLoop := flag.Int("trace-loop", 0, "Endlosschleifen-Laeufe aufzeichnen (max. Anzahl, 0 = aus)")
	traceWin := flag.Int("trace-win", 0, "Gewonnene Laeufe aufzeichnen (alle Bedarfe erfuellt, max. n, 0 = aus)")
	demandI := flag.Int("demand-i", 1, "Bedarf Industrie (I)")
	demandW := flag.Int("demand-w", 1, "Bedarf Wohnviertel (W)")
	demandB := flag.Int("demand-b", 0, "Bedarf Bahn (b)")
	demandR := flag.Int("demand-r", 1, "Bedarf Reaktoreigenbedarf (R)")
	damageI := flag.Int("damage-i", 0, "Schaden Industrie (I)")
	damageW := flag.Int("damage-w", 0, "Schaden Wohnviertel (W)")
	damageB := flag.Int("damage-b", 0, "Schaden Bahn (b)")
	damageR := flag.Int("damage-r", 0, "Schaden Reaktoreigenbedarf (R)")
	costP1 := flag.Int("cost-p1", 0, "Schicht-Budget Spieler 1 / Reaktor in Geld (0 = zufaellig, nicht alles muss ausgegeben werden)")
	costP2 := flag.Int("cost-p2", 0, "Schicht-Budget Spieler 2 / Stromnetz in Geld (0 = zufaellig, nicht alles muss ausgegeben werden)")
	prevBoard := flag.String("prev-board", "", "Board-Fingerprint des bezahlten Bretts der Vorschicht")
	financeID := flag.String("finanz-karte", "", "Finanz-Karte fuer Schicht-Budget (Schaden-Reparatur): "+financeCardIDs())
	repairBudget := flag.Int("repair-budget", -1, "Max. Geld fuer Schaden-Reparatur je Lauf (1 Geld/Chip); -1 = gesamtes Restbudget Stromnetz nach Feld-Kauf")
	monthFilter := flag.Int("month-filter", 0, "Kampagnenmonat: nur dann verfuegbare Felder kaufen (0 = alle)")
	flag.Parse()

	for name, v := range map[string]int{
		"demand-i": *demandI, "demand-w": *demandW, "demand-b": *demandB, "demand-r": *demandR,
		"damage-i": *damageI, "damage-w": *damageW, "damage-b": *damageB, "damage-r": *damageR,
	} {
		if v < 0 {
			log.Fatalf("-%s muss >= 0 sein", name)
		}
	}

	shiftDemands := board.ShiftDemands{
		Industry:    *demandI,
		Residential: *demandW,
		Rail:        *demandB,
		Plant:       *demandR,
	}
	shiftDamage := board.ShiftDemands{
		Industry:    *damageI,
		Residential: *damageW,
		Rail:        *damageB,
		Plant:       *damageR,
	}

	if *traceFirst < -1 {
		log.Fatal("-trace-first muss >= -1 sein (-1 = aus, 0 = alle -runs)")
	}
	if *traceLoop < 0 {
		log.Fatal("-trace-loop muss >= 0 sein")
	}
	if *traceWin < 0 {
		log.Fatal("-trace-win muss >= 0 sein")
	}
	if *monthFilter < 0 {
		log.Fatal("-month-filter muss >= 0 sein")
	}

	if err := os.RemoveAll(*outDir); err != nil {
		log.Fatalf("Ausgabeverzeichnis leeren: %v", err)
	}

	if *seed == 0 {
		*seed = time.Now().UnixNano()
	}
	rng := rand.New(rand.NewSource(*seed))

	var financeCard finance.Card
	if *financeID != "" {
		var ok bool
		financeCard, ok = finance.ByID(*financeID)
		if !ok {
			log.Fatalf("unbekannte Finanz-Karte %q", *financeID)
		}
	}
	monthRules := rules.Month{FinanceID: financeCard.ID}

	var state *board.State
	var leftover board.PlayerLeftover
	var err error
	switch {
	case *prevBoard != "":
		state, err = board.FromFingerprint(*prevBoard)
		if err != nil {
			log.Fatalf("prev-board: %v", err)
		}
		spendRes, spendErr := board.SpendShiftBudget(rng, state, *costP1, *costP2, *monthFilter, board.MinFirstShiftFieldSpend, monthRules)
		leftover = spendRes.Leftover
		err = spendErr
	case *costP1 > 0 || *costP2 > 0:
		state, leftover, err = board.RandomWithPlayerCosts(rng, *costP1, *costP2, *monthFilter, board.MinFirstShiftFieldSpend, monthRules)
	default:
		state = board.Random(rng, *monthFilter)
	}
	if err != nil {
		log.Fatal(err)
	}
	state.Damage = [4]int{
		shiftDamage.Industry,
		shiftDamage.Residential,
		shiftDamage.Rail,
		shiftDamage.Plant,
	}

	cfg := sim.DefaultConfig()
	cfg.EnergyCard = energy.Card{}
	cfg.FinanceCard = financeCard
	cfg.CriticalLimit = monthRules.CriticalLimit()
	cfg.Shift = 1
	cfg.RandomShift = false
	cfg.ShiftDemands = shiftDemands
	cfg.ReactorRepairBudget, cfg.RepairBudget = repairBudgetsForRun(state, financeCard, *repairBudget, leftover.Player1, leftover.Player2)
	preview := state.Clone()
	preview.ApplyDemands(cfg.ShiftDemands)
	previewChips := sim.EmitterChips(preview, cfg, rng)

	fmt.Printf("Reaktor-Sim: %d Durchläufe (Seed %d)\n", *runs, *seed)
	boardFP := board.Fingerprint(state)
	fmt.Printf("Board-Fingerprint: %s\n", boardFP)
	boardCosts := state.PlayerCostsFor(monthRules)
	fmt.Printf("Board-Kosten: %s (gesamt %d Geld)\n", boardCosts.String(), boardCosts.Total())
	if *monthFilter > 0 {
		fmt.Printf("Monats-Filter: %d (nur ab diesem Monat verfuegbare Felder)\n", *monthFilter)
	}
	if leftover.Player1 > 0 || leftover.Player2 > 0 {
		fmt.Printf("Restbudget: Reaktor %d Geld | Stromnetz %d Geld\n", leftover.Player1, leftover.Player2)
	}
	fmt.Printf("Bedarfe: I=%d W=%d b=%d R=%d | Schaden: I=%d W=%d b=%d R=%d",
		shiftDemands.Industry, shiftDemands.Residential, shiftDemands.Rail, shiftDemands.Plant,
		shiftDamage.Industry, shiftDamage.Residential, shiftDamage.Rail, shiftDamage.Plant)
	if state.EmitterDamage > 0 {
		fmt.Printf(" | Zünder-Schaden: %d", state.EmitterDamage)
	}
	fmt.Println()
	if cfg.ReactorRepairBudget > 0 {
		fmt.Printf("Zünder-Reparatur: bis zu %d Geld je Lauf (Restbudget Reaktor)\n", cfg.ReactorRepairBudget)
	}
	if cfg.RepairBudget > 0 {
		fmt.Printf("Zonen-Reparatur: bis zu %d Geld je Lauf (Restbudget Stromnetz, zufaellige Chips)\n", cfg.RepairBudget)
	} else if state.TotalPlayer2Damage() > 0 && leftover.Player2 > 0 && financeCard.RepairsAllowed() {
		fmt.Println("Zonen-Reparatur: kein Restbudget Stromnetz nach Feld-Kauf")
	} else if state.TotalPlayer2Damage() > 0 && !financeCard.RepairsAllowed() {
		fmt.Println("Zonen-Reparatur: nicht bewilligt (Finanz-Karte)")
	}
	printDemands(preview)

	traceIndex := make([]render.TraceIndexEntry, 0)
	traceCosts := state.PlayerCostsFor(monthRules)

	if *traceFirst >= 0 {
		n := *traceFirst
		if n == 0 {
			n = *runs
		}
		if n < 1 {
			log.Fatal("-trace-first muss >= 1 sein (oder 0 fuer alle -runs)")
		}
		if n > *runs {
			fmt.Printf("Hinweis: -trace-first (%d) > -runs (%d), zeichne nur %d Laeufe auf\n", n, *runs, *runs)
			n = *runs
		}
		totalSteps := 0
		for run := 1; run <= n; run++ {
			traceRNG := rand.New(rand.NewSource(*seed + int64(run)))
			res, snaps := sim.RunTrace(state, traceRNG, cfg)
			meta := render.TraceMeta{
				Shift: res.Shift,
				Costs: traceCosts,
			}
			runDir := filepath.Join(*outDir, fmt.Sprintf("run%d", run))
			if err := render.WriteRunTrace(run, meta, snaps, *outDir); err != nil {
				log.Printf("Warnung: Trace run%d — trace.yaml liegt vor, PNG-Fehler: %v", run, err)
			}
			totalSteps += len(snaps)
			absDir, _ := filepath.Abs(runDir)
			traceIndex = append(traceIndex, render.TraceIndexEntry{
				Run:   run,
				Kind:  render.TraceKindFirst,
				Steps: len(snaps),
				Dir:   absDir,
			})
			fmt.Printf("Trace run%d: %d Schritte → %s/trace.yaml\n", run, len(snaps), absDir)
		}
		fmt.Printf("Trace gesamt: %d Laeufe, %d Schrittbilder\n", n, totalSteps)
	}

	results := sim.RunMonteCarlo(state, *runs, *seed, cfg)

	if *traceLoop > 0 {
		loopRuns := sim.LoopTraceRunIndices(results, *traceLoop)
		for loopNum, mcRun := range loopRuns {
			traceRNG := rand.New(rand.NewSource(*seed + int64(mcRun)))
			res, snaps := sim.RunTrace(state, traceRNG, cfg)
			meta := render.TraceMeta{
				Shift: res.Shift,
				Costs: traceCosts,
			}
			seq := loopNum + 1
			runDir := filepath.Join(*outDir, fmt.Sprintf("loop%d", seq))
			if err := render.WriteLoopTrace(seq, mcRun, meta, snaps, *outDir); err != nil {
				log.Printf("Warnung: Loop-Trace loop%d (MC-Lauf %d) — trace.yaml liegt vor, PNG-Fehler: %v", seq, mcRun, err)
			}
			absDir, _ := filepath.Abs(runDir)
			traceIndex = append(traceIndex, render.TraceIndexEntry{
				Run:           seq,
				MonteCarloRun: mcRun,
				Kind:          render.TraceKindLoop,
				Steps:         len(snaps),
				Dir:           absDir,
			})
			fmt.Printf("Loop-Trace loop%d: MC-Lauf %d, %d Schritte → %s/trace.yaml\n", seq, mcRun, len(snaps), absDir)
		}
		if len(loopRuns) == 0 {
			fmt.Printf("Loop-Trace: kein Lauf mit Schrittlimit in %d Durchlaeufen\n", *runs)
		} else {
			fmt.Printf("Loop-Trace gesamt: %d Laeufe aufgezeichnet (Limit %d)\n", len(loopRuns), *traceLoop)
		}
	}

	if *traceWin > 0 {
		winRuns := sim.WinTraceRunIndices(results, *traceWin)
		for winNum, mcRun := range winRuns {
			traceRNG := rand.New(rand.NewSource(*seed + int64(mcRun)))
			res, snaps := sim.RunTrace(state, traceRNG, cfg)
			meta := render.TraceMeta{
				Shift: res.Shift,
				Costs: traceCosts,
			}
			seq := winNum + 1
			runDir := filepath.Join(*outDir, fmt.Sprintf("win%d", seq))
			if err := render.WriteWinTrace(seq, mcRun, meta, snaps, *outDir); err != nil {
				log.Printf("Warnung: Win-Trace win%d (MC-Lauf %d) — trace.yaml liegt vor, PNG-Fehler: %v", seq, mcRun, err)
			}
			absDir, _ := filepath.Abs(runDir)
			traceIndex = append(traceIndex, render.TraceIndexEntry{
				Run:           seq,
				MonteCarloRun: mcRun,
				Kind:          render.TraceKindWin,
				Steps:         len(snaps),
				Dir:           absDir,
			})
			fmt.Printf("Win-Trace win%d: MC-Lauf %d, %d Schritte → %s/trace.yaml\n", seq, mcRun, len(snaps), absDir)
		}
		if len(winRuns) == 0 {
			fmt.Printf("Win-Trace: kein Lauf mit allen Bedarfen erfuellt in %d Durchlaeufen\n", *runs)
		} else {
			fmt.Printf("Win-Trace gesamt: %d Laeufe aufgezeichnet (Limit %d)\n", len(winRuns), *traceWin)
		}
	}

	if len(traceIndex) > 0 {
		if err := render.WriteTraceIndex(*outDir, traceIndex); err != nil {
			log.Printf("Warnung: trace_index.yaml: %v", err)
		}
	}
	report := stats.Build(state.PlayerCostsFor(monthRules), leftover, results)

	printSummary(report)

	initialView := render.ChipView{Queue: previewChips}
	if err := render.WriteAll(preview, *outDir, initialView, render.BoardMeta{
		Seed:      *seed,
		PrevBoard: *prevBoard,
	}); err != nil {
		log.Fatalf("Board rendern: %v", err)
	}
	if err := charts.WriteAll(report, *outDir); err != nil {
		log.Fatalf("Charts schreiben: %v", err)
	}
	fmt.Printf("\nAusgabe gespeichert in %s/\n", *outDir)
	fmt.Printf("  spielfeld-%s.png / spielfeld-%s.yaml – Brett mit Symbolen\n", boardFP, boardFP)
	if *traceFirst >= 0 {
		fmt.Println("  runN/trace.yaml – Simulationstrace pro Lauf (siehe trace_index.yaml)")
		fmt.Println("  runN/graph_runN_SSS.png – Graph pro Schritt")
	}
	if *traceLoop > 0 {
		fmt.Println("  loopN/trace.yaml – Trace fuer Schrittlimit-Laeufe (siehe trace_index.yaml)")
		fmt.Println("  loopN/graph_loopN_SSS.png – Graph pro Schritt")
	}
	if *traceWin > 0 {
		fmt.Println("  winN/trace.yaml – Trace fuer gewonnene Laeufe (siehe trace_index.yaml)")
		fmt.Println("  winN/graph_winN_SSS.png – Graph pro Schritt")
	}
	if *traceFirst < 0 && *traceLoop == 0 && *traceWin == 0 {
		fmt.Println("  (keine runN/-, loopN/- oder winN/-Ordner ohne Trace-Flags)")
	}
	if len(traceIndex) > 0 {
		fmt.Printf("  trace_index.yaml – Index aller Traces\n")
	}
}

func printDemands(state *board.State) {
	zones := []board.Zone{
		board.ZoneIndustry,
		board.ZoneResidential,
		board.ZoneRail,
		board.ZonePlant,
	}
	fmt.Print("Bedarfe: ")
	for i, z := range zones {
		if i > 0 {
			fmt.Print("  ")
		}
		fmt.Printf("%s=%d", z.String(), state.TotalDemand(z))
	}
	fmt.Print(" | Schaden: ")
	for i, z := range zones {
		if i > 0 {
			fmt.Print("  ")
		}
		fmt.Printf("%s=%d", z.String(), state.TotalDamage(z))
	}
	fmt.Println()
}

func printSummary(report stats.Report) {
	fmt.Printf("\n--- Wärme an Turbine ---\n")
	printHist(report.HeatAtTurbine)

	zones := []board.Zone{
		board.ZoneResidential,
		board.ZoneIndustry,
		board.ZoneRail,
		board.ZonePlant,
	}
	for _, z := range zones {
		fmt.Printf("\n--- Spannung %s ---\n", z.String())
		printHist(report.ZoneHistograms[z])
	}
	fmt.Printf("\nKritische Masse überschritten: %.1f%% (P1: %.1f%%, P2: %.1f%%)\n",
		report.CriticalFailRate*100, report.CriticalP1Rate*100, report.CriticalP2Rate*100)
}

func printHist(h stats.Histogram) {
	keys := stats.SortedKeys(h)
	for _, k := range keys {
		fmt.Printf("  %d: %.1f%%\n", k, h[k]*100)
	}
}

func init() {
	log.SetOutput(os.Stderr)
}

func financeCardIDs() string {
	ids := make([]string, len(finance.Cards))
	for i, c := range finance.Cards {
		ids[i] = c.ID
	}
	return strings.Join(ids, ", ")
}

func repairBudgetsForRun(state *board.State, fin finance.Card, flagGridBudget, leftoverP1, leftoverP2 int) (reactor, grid int) {
	if !fin.RepairsAllowed() {
		return 0, 0
	}
	grid = flagGridBudget
	if grid < 0 {
		total := state.TotalPlayer2Damage()
		if leftoverP2 > total {
			grid = total
		} else {
			grid = leftoverP2
		}
	} else if total := state.TotalPlayer2Damage(); grid > total {
		grid = total
	}
	reactor = leftoverP1
	if reactor > state.EmitterDamage {
		reactor = state.EmitterDamage
	}
	return reactor, grid
}
