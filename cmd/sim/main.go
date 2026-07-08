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
	trace := flag.Bool("trace", false, "Schichten aufzeichnen und Graph pro Schritt speichern")
	traceRuns := flag.Int("trace-runs", 0, "Aufgezeichnete Laeufe (0 = gleich -runs)")
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
	state.ApplyDemands(cfg.ShiftDemands)
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
	printDemands(state)

	if *trace {
		n := *traceRuns
		if n == 0 {
			n = *runs
		}
		if n < 1 {
			log.Fatal("-trace-runs muss >= 1 sein (oder 0 fuer -runs)")
		}
		if n > *runs {
			fmt.Printf("Hinweis: -trace-runs (%d) > -runs (%d), zeichne nur %d Laeufe auf\n", n, *runs, *runs)
			n = *runs
		}
		totalSteps := 0
		index := make([]render.TraceIndexEntry, 0, n)
		for run := 1; run <= n; run++ {
			traceRNG := rand.New(rand.NewSource(*seed + int64(run)))
			_, snaps := sim.RunTrace(state, traceRNG, cfg)
			runDir := filepath.Join(*outDir, fmt.Sprintf("run%d", run))
			if err := render.WriteRunTrace(run, snaps, *outDir); err != nil {
				log.Printf("Warnung: Trace run%d — trace.txt liegt vor, PNG-Fehler: %v", run, err)
			}
			totalSteps += len(snaps)
			absDir, _ := filepath.Abs(runDir)
			index = append(index, render.TraceIndexEntry{
				Run:   run,
				Steps: len(snaps),
				Dir:   absDir,
			})
			fmt.Printf("Trace run%d: %d Schritte → %s/trace.txt\n", run, len(snaps), absDir)
		}
		if err := render.WriteTraceIndex(*outDir, index); err != nil {
			log.Printf("Warnung: trace_index.txt: %v", err)
		}
		fmt.Printf("Trace gesamt: %d Laeufe, %d Schrittbilder (Index: %s/trace_index.txt)\n", n, totalSteps, *outDir)
	}

	results := sim.RunMonteCarlo(state, *runs, rng, cfg)
	report := stats.Build(state.PlayerCosts(), results)

	printSummary(report)

	initialView := render.ChipView{Queue: previewChips}
	if err := render.WriteAll(state, *outDir, initialView); err != nil {
		log.Fatalf("Board rendern: %v", err)
	}
	if err := charts.WriteAll(report, *outDir); err != nil {
		log.Fatalf("Charts schreiben: %v", err)
	}
	fmt.Printf("\nAusgabe gespeichert in %s/\n", *outDir)
	fmt.Println("  spielfeld.png / spielfeld.txt – Brett mit Symbolen")
	if *trace {
		fmt.Println("  runN/trace.txt – Simulationstrace pro Lauf (siehe trace_index.txt)")
		fmt.Println("  runN/graph_runN_SSS.png – Graph pro Schritt")
	} else {
		fmt.Println("  (keine runN/-Ordner ohne -trace)")
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
	fmt.Printf("\nKritische Masse überschritten: %.1f%%\n", report.CriticalFailRate*100)
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
