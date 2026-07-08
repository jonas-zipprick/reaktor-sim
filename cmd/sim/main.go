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
	"github.com/jonas/reaktor-sim/internal/render"
	"github.com/jonas/reaktor-sim/internal/sim"
	"github.com/jonas/reaktor-sim/internal/stats"
)

func main() {
	runs := flag.Int("runs", 20, "Anzahl Monte-Carlo-Durchlaeufe")
	seed := flag.Int64("seed", 0, "Zufallsseed (0 = aktuelle Zeit)")
	outDir := flag.String("out", "output", "Ausgabeverzeichnis für Charts")
	initialHeat := flag.Int("heat", 1, "Waerme-Chips am Zuender pro Schicht (nur ohne -mixed-trigger)")
	initialNeutron := flag.Int("neutron", 0, "Neutron-Chips am Zuender pro Schicht (nur ohne -mixed-trigger)")
	mixedTrigger := flag.Bool("mixed-trigger", true, "1 Basis-Trigger pro Schicht: zufaellig Waerme oder Neutron")
	traceFirst := flag.Int("trace-first", -1, "Erste n Laeufe aufzeichnen (-1 = aus, 0 = alle -runs)")
	traceLoop := flag.Int("trace-loop", 0, "Endlosschleifen-Laeufe aufzeichnen (max. Anzahl, 0 = aus)")
	traceWin := flag.Int("trace-win", 0, "Gewonnene Laeufe aufzeichnen (alle Bedarfe erfuellt, max. n, 0 = aus)")
	energyCard := flag.String("energy-card", energy.DefaultCard().ID, "Energiekarte (ID, z.B. eroeffnungsfeier)")
	shift := flag.Int("shift", 1, "Schicht 1-5 auf der Energiekarte (0 = zufaellig pro Lauf)")
	costP1 := flag.Int("cost-p1", 0, "Brettkosten Spieler 1 / Reaktor in Geld (0 = zufaellig)")
	costP2 := flag.Int("cost-p2", 0, "Brettkosten Spieler 2 / Stromnetz in Geld (0 = zufaellig)")
	flag.Parse()

	card, ok := energy.ByID(*energyCard)
	if !ok {
		log.Fatalf("unbekannte Energiekarte %q (verfuegbar: %s)", *energyCard, listEnergyCardIDs())
	}
	if *shift < 0 || *shift > 5 {
		log.Fatal("-shift muss 0-5 sein")
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

	if err := os.RemoveAll(*outDir); err != nil {
		log.Fatalf("Ausgabeverzeichnis leeren: %v", err)
	}

	if *seed == 0 {
		*seed = time.Now().UnixNano()
	}
	rng := rand.New(rand.NewSource(*seed))

	var state *board.State
	if *costP1 > 0 || *costP2 > 0 {
		var err error
		state, err = board.RandomWithPlayerCosts(rng, *costP1, *costP2)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		state = board.Random(rng)
	}
	cfg := sim.DefaultConfig()
	cfg.InitialHeat = *initialHeat
	cfg.InitialNeutron = *initialNeutron
	cfg.MixedEmitterTrigger = *mixedTrigger
	cfg.EnergyCard = card
	cfg.Shift = *shift
	cfg.RandomShift = *shift == 0
	if cfg.RandomShift {
		cfg.ShiftDemands = card.ShiftDemands(1)
	} else {
		cfg.ShiftDemands = card.ShiftDemands(*shift)
	}
	preview := state.Clone()
	preview.ApplyDemands(cfg.ShiftDemands)
	previewChips := sim.EmitterChips(cfg, rng)

	fmt.Printf("Reaktor-Sim: %d Durchläufe (Seed %d)\n", *runs, *seed)
	boardCosts := state.PlayerCosts()
	fmt.Printf("Board-Kosten: %s (gesamt %d Geld)\n", boardCosts.String(), boardCosts.Total())
	if cfg.RandomShift {
		fmt.Printf("Energiekarte: %s (Stufe %d), Schicht zufaellig 1-5\n", card.Name, card.Level)
	} else {
		fmt.Println(card.DescribeShift(*shift))
	}
	if card.SpecialRule != "" {
		fmt.Printf("Sonderregel (noch nicht simuliert): %s\n", card.SpecialRule)
	}
	printDemands(preview)

	traceIndex := make([]render.TraceIndexEntry, 0)
	traceCosts := state.PlayerCosts()

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
				EnergyCardID:    card.ID,
				EnergyCardName:  card.Name,
				EnergyCardLevel: card.Level,
				Shift:           res.Shift,
				Costs:           traceCosts,
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
				EnergyCardID:    card.ID,
				EnergyCardName:  card.Name,
				EnergyCardLevel: card.Level,
				Shift:           res.Shift,
				Costs:           traceCosts,
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
				EnergyCardID:    card.ID,
				EnergyCardName:  card.Name,
				EnergyCardLevel: card.Level,
				Shift:           res.Shift,
				Costs:           traceCosts,
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
	report := stats.Build(state.PlayerCosts(), results)

	printSummary(report)

	initialView := render.ChipView{Queue: previewChips}
	if err := render.WriteAll(preview, *outDir, initialView); err != nil {
		log.Fatalf("Board rendern: %v", err)
	}
	if err := charts.WriteAll(report, *outDir); err != nil {
		log.Fatalf("Charts schreiben: %v", err)
	}
	fmt.Printf("\nAusgabe gespeichert in %s/\n", *outDir)
	fmt.Println("  spielfeld.png / spielfeld.yaml – Brett mit Symbolen")
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

func listEnergyCardIDs() string {
	ids := make([]string, len(energy.Cards))
	for i, c := range energy.Cards {
		ids[i] = c.ID
	}
	return strings.Join(ids, ", ")
}
