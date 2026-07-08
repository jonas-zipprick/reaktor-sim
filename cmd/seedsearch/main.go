package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/energy"
	"github.com/jonas/reaktor-sim/internal/progress"
	"github.com/jonas/reaktor-sim/internal/seedsearch"
	"gopkg.in/yaml.v3"
)

func main() {
	from := flag.Int64("from", 1, "Erster Seed (inklusive)")
	to := flag.Int64("to", 0, "Letzter Seed (inklusive); 0 = from allein")
	runs := flag.Int("runs", 100, "Monte-Carlo-Laeufe pro Seed")
	top := flag.Int("top", 10, "Anzahl Seeds in jeder Rangliste")
	outFile := flag.String("out", "", "Ergebnis als YAML speichern (optional)")
	progressBar := flag.Bool("progress", true, "Fortschrittsbalken auf stderr anzeigen")
	initialHeat := flag.Int("heat", 1, "Waerme-Chips am Zuender pro Schicht (nur ohne -mixed-trigger)")
	initialNeutron := flag.Int("neutron", 0, "Neutron-Chips am Zuender pro Schicht (nur ohne -mixed-trigger)")
	mixedTrigger := flag.Bool("mixed-trigger", true, "1 Basis-Trigger pro Schicht: zufaellig Waerme oder Neutron")
	energyCard := flag.String("energy-card", energy.DefaultCard().ID, "Energiekarte (ID)")
	shift := flag.Int("shift", 1, "Schicht 1-5 auf der Energiekarte (0 = zufaellig pro Lauf)")
	costP1 := flag.Int("cost-p1", 0, "Brettkosten Spieler 1 / Reaktor in Geld (0 = zufaellig)")
	costP2 := flag.Int("cost-p2", 0, "Brettkosten Spieler 2 / Stromnetz in Geld (0 = zufaellig)")
	flag.Parse()

	if *to == 0 {
		*to = *from
	}
	if *from > *to {
		log.Fatal("-from muss <= -to sein")
	}
	if *runs < 1 {
		log.Fatal("-runs muss >= 1 sein")
	}
	if *top < 1 {
		log.Fatal("-top muss >= 1 sein")
	}
	if *shift < 0 || *shift > 5 {
		log.Fatal("-shift muss 0-5 sein")
	}

	opts := seedsearch.Options{
		Runs:                *runs,
		InitialHeat:         *initialHeat,
		InitialNeutron:      *initialNeutron,
		MixedEmitterTrigger: *mixedTrigger,
		EnergyCardID:        *energyCard,
		Shift:               *shift,
		CostP1:              *costP1,
		CostP2:              *costP2,
	}

	total := *to - *from + 1
	fmt.Printf("Seed-Suche: Seeds %d–%d (%d Bretter), %d Laeufe je Seed\n", *from, *to, total, *runs)
	if *costP1 > 0 || *costP2 > 0 {
		fmt.Printf("Brettkosten (fest): Reaktor %d Geld | Stromnetz %d Geld\n", *costP1, *costP2)
	} else {
		fmt.Println("Brettkosten: zufaellig pro Seed (-cost-p1/-cost-p2 zum Fixieren)")
	}
	if card, ok := energy.ByID(*energyCard); ok {
		if *shift == 0 {
			fmt.Printf("Energiekarte: %s (Stufe %d), Schicht zufaellig 1-5\n", card.Name, card.Level)
		} else {
			fmt.Println(card.DescribeShift(*shift))
		}
	}
	start := time.Now()

	var bar *progress.Bar
	var onProgress seedsearch.ProgressFunc
	if *progressBar {
		bar = progress.NewBar("Seeds", total, 30)
		onProgress = func(done, _ int64) { bar.Set(done) }
	}

	outcomes, err := seedsearch.Scan(*from, *to, opts, onProgress)
	if err != nil {
		log.Fatal(err)
	}
	if bar != nil {
		bar.Finish()
	}

	elapsed := time.Since(start)
	fmt.Printf("Fertig in %s (%.1f Seeds/s)\n\n", elapsed.Round(time.Millisecond), float64(total)/elapsed.Seconds())

	printTable("Top Seeds nach Wins (alle Bedarfe erfuellt)", seedsearch.TopWins(outcomes, *top), *runs, winCols)
	printTable("Top Seeds nach Loops (Schrittlimit)", seedsearch.TopLoops(outcomes, *top), *runs, loopCols)

	if *outFile != "" {
		if err := writeYAML(*outFile, outcomes, opts, *from, *to, *top); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("\nVollstaendige Ergebnisse: %s\n", *outFile)
	}
}

type col struct {
	title string
	width int
	value func(seedsearch.Outcome) string
}

var winCols = []col{
	{"Seed", 10, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.Seed) }},
	{"P1", 4, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.BoardCosts.Player1) }},
	{"P2", 4, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.BoardCosts.Player2) }},
	{"Wins", 6, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.Wins) }},
	{"Win%", 7, func(o seedsearch.Outcome) string { return fmt.Sprintf("%.1f", o.WinRate()*100) }},
	{"Loops", 6, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.Loops) }},
	{"Kritisch P1", 11, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.CriticalP1) }},
	{"Kritisch P2", 11, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.CriticalP2) }},
	{"End-Bedarf", 16, func(o seedsearch.Outcome) string { return o.EndDemandSummary() }},
	{"End-Schaden", 14, func(o seedsearch.Outcome) string { return o.EndDamageSummary() }},
}

var loopCols = []col{
	{"Seed", 10, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.Seed) }},
	{"P1", 4, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.BoardCosts.Player1) }},
	{"P2", 4, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.BoardCosts.Player2) }},
	{"Loops", 6, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.Loops) }},
	{"Loop%", 7, func(o seedsearch.Outcome) string { return fmt.Sprintf("%.1f", o.LoopRate()*100) }},
	{"Wins", 6, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.Wins) }},
	{"Kritisch P1", 11, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.CriticalP1) }},
	{"Kritisch P2", 11, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.CriticalP2) }},
	{"End-Bedarf", 16, func(o seedsearch.Outcome) string { return o.EndDemandSummary() }},
	{"End-Schaden", 14, func(o seedsearch.Outcome) string { return o.EndDamageSummary() }},
}

func printTable(title string, rows []seedsearch.Outcome, runs int, cols []col) {
	fmt.Println(title)
	fmt.Println(strings.Repeat("-", len(title)))
	header := make([]string, len(cols))
	under := make([]string, len(cols))
	for i, c := range cols {
		header[i] = fmt.Sprintf("%-*s", c.width, c.title)
		under[i] = strings.Repeat("-", c.width)
	}
	fmt.Println(strings.Join(header, " "))
	fmt.Println(strings.Join(under, " "))
	for _, row := range rows {
		cells := make([]string, len(cols))
		for i, c := range cols {
			cells[i] = fmt.Sprintf("%-*s", c.width, c.value(row))
		}
		fmt.Println(strings.Join(cells, " "))
	}
	if len(rows) == 0 {
		fmt.Println("(keine)")
	}
	fmt.Printf("(%d Laeufe pro Seed)\n\n", runs)
}

type reportYAML struct {
	From     int64         `yaml:"from"`
	To       int64         `yaml:"to"`
	Runs     int           `yaml:"runs_per_seed"`
	Top      int           `yaml:"top"`
	Settings settingsYAML  `yaml:"settings"`
	Outcomes []outcomeYAML `yaml:"outcomes"`
	TopWins  []outcomeYAML `yaml:"top_wins"`
	TopLoops []outcomeYAML `yaml:"top_loops"`
}

type settingsYAML struct {
	EnergyCard string    `yaml:"energy_card"`
	Shift      int       `yaml:"shift"`
	CostP1     int       `yaml:"cost_p1"`
	CostP2     int       `yaml:"cost_p2"`
	CostMode   string    `yaml:"cost_mode"`
}

type costsYAML struct {
	Reaktor   int `yaml:"reaktor"`
	Stromnetz int `yaml:"stromnetz"`
	Total     int `yaml:"total"`
}

type zoneTotalsYAML struct {
	Industry    float64 `yaml:"industrie"`
	Residential float64 `yaml:"wohnviertel"`
	Rail        float64 `yaml:"bahn"`
	Plant       float64 `yaml:"reaktorbedarf"`
}

type outcomeYAML struct {
	Seed         int64          `yaml:"seed"`
	BoardCosts   costsYAML      `yaml:"board_costs"`
	Wins         int            `yaml:"wins"`
	Loops        int            `yaml:"loops"`
	CriticalP1   int            `yaml:"critical_p1"`
	CriticalP2   int            `yaml:"critical_p2"`
	AvgEndDemand zoneTotalsYAML `yaml:"avg_end_demand"`
	AvgEndDamage zoneTotalsYAML `yaml:"avg_end_damage"`
	WinRate      float64        `yaml:"win_rate"`
	LoopRate     float64        `yaml:"loop_rate"`
}

func writeYAML(path string, outcomes []seedsearch.Outcome, opts seedsearch.Options, from, to int64, top int) error {
	costMode := "random"
	if opts.CostP1 > 0 || opts.CostP2 > 0 {
		costMode = "fixed"
	}
	doc := reportYAML{
		From: from,
		To:   to,
		Runs: outcomes[0].Runs,
		Top:  top,
		Settings: settingsYAML{
			EnergyCard: opts.EnergyCardID,
			Shift:      opts.Shift,
			CostP1:     opts.CostP1,
			CostP2:     opts.CostP2,
			CostMode:   costMode,
		},
		Outcomes: toOutcomeYAML(outcomes),
		TopWins:  toOutcomeYAML(seedsearch.TopWins(outcomes, top)),
		TopLoops: toOutcomeYAML(seedsearch.TopLoops(outcomes, top)),
	}
	data, err := yaml.Marshal(doc)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func toOutcomeYAML(in []seedsearch.Outcome) []outcomeYAML {
	out := make([]outcomeYAML, len(in))
	for i, o := range in {
		out[i] = outcomeYAML{
			Seed: o.Seed,
			BoardCosts: costsYAML{
				Reaktor:   o.BoardCosts.Player1,
				Stromnetz: o.BoardCosts.Player2,
				Total:     o.BoardCosts.Total(),
			},
			Wins:         o.Wins,
			Loops:        o.Loops,
			CriticalP1:   o.CriticalP1,
			CriticalP2:   o.CriticalP2,
			AvgEndDemand: toZoneTotalsYAML(o.AvgEndDemand),
			AvgEndDamage: toZoneTotalsYAML(o.AvgEndDamage),
			WinRate:      o.WinRate(),
			LoopRate:     o.LoopRate(),
		}
	}
	return out
}

func toZoneTotalsYAML(t seedsearch.ZoneTotals) zoneTotalsYAML {
	return zoneTotalsYAML{
		Industry:    t[board.ZoneIndustry],
		Residential: t[board.ZoneResidential],
		Rail:        t[board.ZoneRail],
		Plant:       t[board.ZonePlant],
	}
}

func init() {
	log.SetOutput(os.Stderr)
}
