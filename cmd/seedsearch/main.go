package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/charts"
	"github.com/jonas/reaktor-sim/internal/energy"
	"github.com/jonas/reaktor-sim/internal/finance"
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
	chartsDir := flag.String("charts-dir", "seedsearch-out", "Ausgabeverzeichnis fuer Chart-PNGs und report.txt (leer = aus)")
	progressBar := flag.Bool("progress", true, "Fortschrittsbalken auf stderr anzeigen")
	energyID := flag.String("energie-karte", energy.DefaultCard().ID, "Energie-Karte (Schichtplan): "+energyCardIDs())
	financeID := flag.String("finanz-karte", finance.DefaultCard().ID, "Finanz-Karte (Schicht-Budget): "+financeCardIDs())
	shifts := flag.Int("schichten", 1, "Anzahl aufeinanderfolgender Schichten (1-5, ganzer Monat = 5)")
	shiftKeep := flag.Int("schicht-keep", 1, "Top-Boards je Rangliste, die in die naechste Schicht weiterverzweigt werden")
	workers := flag.Int("workers", 0, "Parallele Scan-Worker (0 = GOMAXPROCS)")
	topSims := flag.Bool("top-sims", true, "Top-Sim-Ordner mit Trace-Grafiken exportieren")
	tracePNG := flag.Bool("trace-png", true, "Schritt-PNGs in Top-Sim-Traces schreiben")
	topSimCharts := flag.Bool("top-sim-charts", false, "Zusatz-Charts je Top-Sim (fuehrt Monte-Carlo erneut aus)")
	outFull := flag.Bool("out-full", false, "YAML enthaelt alle Outcomes (sonst nur Top-Listen)")
	spillDir := flag.String("spill-dir", "", "Verzeichnis fuer ausgelagerte Schicht-Ergebnisse (leer = bei -schichten>1 auto unter -charts-dir/.spill)")
	spillMemoryMB := flag.Int("spill-memory-mb", 1024, "Schicht-Ergebnisse erst ab dieser Prozess-Speichernutzung auf Disk (0 = immer wenn -spill-dir aktiv)")
	monthFilter := flag.Int("month-filter", 0, "Kampagnenmonat: nur dann verfuegbare Felder kaufen (0 = alle)")
	startBoard := flag.String("start-board", "", "Board-Fingerprint als Startbrett (Folgemonat: Kaufvarianten statt Neugenerierung)")
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
	if *shifts < 1 || *shifts > 5 {
		log.Fatal("-schichten muss zwischen 1 und 5 liegen")
	}
	if *shiftKeep < 1 {
		log.Fatal("-schicht-keep muss >= 1 sein")
	}
	if *workers < 0 {
		log.Fatal("-workers muss >= 0 sein")
	}
	if *spillMemoryMB < 0 {
		log.Fatal("-spill-memory-mb muss >= 0 sein")
	}
	if *monthFilter < 0 {
		log.Fatal("-month-filter muss >= 0 sein")
	}
	if *startBoard != "" {
		if _, err := board.FromFingerprint(*startBoard); err != nil {
			log.Fatalf("start-board: %v", err)
		}
	}

	energyCard, ok := energy.ByID(*energyID)
	if !ok {
		log.Fatalf("unbekannte Energie-Karte %q (verfuegbar: %s)", *energyID, energyCardIDs())
	}
	financeCard, ok := finance.ByID(*financeID)
	if !ok {
		log.Fatalf("unbekannte Finanz-Karte %q (verfuegbar: %s)", *financeID, financeCardIDs())
	}

	opts := seedsearch.Options{
		Runs:                  *runs,
		EnergyCard:            energyCard,
		Finance:               financeCard,
		Shifts:                *shifts,
		ShiftKeep:             *shiftKeep,
		Workers:               *workers,
		MonthFilter:           *monthFilter,
		StartBoardFingerprint: *startBoard,
	}
	export := simExportOptions{
		tracePNGs:    *tracePNG,
		topSimCharts: *topSimCharts,
	}

	if *spillDir == "" && *shifts > 1 {
		if *chartsDir != "" {
			*spillDir = filepath.Join(*chartsDir, ".spill")
		} else {
			var err error
			*spillDir, err = os.MkdirTemp("", "reaktor-seedsearch-spill-*")
			if err != nil {
				log.Fatalf("Spill-Verzeichnis: %v", err)
			}
		}
	}
	opts.SpillDir = *spillDir
	if *spillMemoryMB == 0 {
		opts.SpillMemoryThreshold = 0
	} else {
		opts.SpillMemoryThreshold = uint64(*spillMemoryMB) << 20
	}

	total := *to - *from + 1
	var reportBuf strings.Builder
	out := io.MultiWriter(os.Stdout, &reportBuf)

	fmt.Fprintf(out, "Seed-Suche: Seeds %d–%d (%d Seeds), %d Laeufe je Seed\n", *from, *to, total, *runs)
	fmt.Fprintf(out, "Energie-Karte: %s (Stufe %d)\n", energyCard.Name, energyCard.Level)
	fmt.Fprintf(out, "Finanz-Karte: %s\n", financeCard.Describe())
	fmt.Fprintf(out, "Schichten: %d\n", *shifts)
	if *startBoard != "" {
		fmt.Fprintf(out, "Start-Board: %s (Schicht 1: Kaufvarianten je Seed)\n", *startBoard)
	}
	if *monthFilter > 0 {
		fmt.Fprintf(out, "Monats-Filter: %d (nur ab diesem Monat verfuegbare Felder)\n", *monthFilter)
	}
	if *shifts > 1 {
		fmt.Fprintf(out, "Schicht-Verzweigung: je %d Top-Boards aus %d Ranglisten × alle Seeds in Folgeschicht\n",
			*shiftKeep, seedsearch.KeepTableCount)
	}
	if opts.SpillDir != "" {
		if opts.SpillMemoryThreshold == 0 {
			fmt.Fprintf(out, "RAM-Auslagerung: immer nach %s\n", opts.SpillDir)
		} else {
			fmt.Fprintf(out, "RAM-Auslagerung: ab %d MB Prozess-Speicher nach %s\n", opts.SpillMemoryThreshold>>20, opts.SpillDir)
		}
	}
	for k := 1; k <= *shifts; k++ {
		d := energyCard.ShiftDemands(k)
		fmt.Fprintf(out, "  Schicht %d Bedarf: I=%d W=%d b=%d R=%d\n", k, d.Industry, d.Residential, d.Rail, d.Plant)
	}

	if *chartsDir != "" {
		if err := prepareChartsDir(*chartsDir); err != nil {
			log.Fatalf("Ausgabeverzeichnis vorbereiten: %v", err)
		}
	}

	start := time.Now()

	var bar *progress.Bar
	var onProgress seedsearch.ProgressFunc
	if *progressBar {
		bar = progress.NewBar("Schicht 1", seedsearch.EstimateScanWork(*from, *to, opts), 30)
		onProgress = func(done, total int64, shift int) {
			if shift > 0 {
				bar.SetLabel(fmt.Sprintf("Schicht %d", shift))
			}
			bar.Set(done)
		}
	}

	scan, err := seedsearch.Scan(*from, *to, opts, onProgress)
	if err != nil {
		log.Fatal(err)
	}
	if bar != nil {
		bar.Finish()
	}

	unique := 0
	if len(scan.Shifts) > 0 {
		unique = scan.Shifts[0].Len()
	}
	elapsed := time.Since(start)
	fmt.Fprintf(out, "Fertig in %s (%.1f Seeds/s, %d eindeutige Bretter, %d Duplikate ausgefiltert)\n\n",
		elapsed.Round(time.Millisecond), float64(total)/elapsed.Seconds(), unique, scan.SkippedDuplicates)

	for _, sr := range scan.Shifts {
		printShiftBlock(out, sr, energyCard, *top, *runs)
	}

	if *chartsDir != "" {
		if err := writeCharts(*chartsDir, scan, *runs); err != nil {
			log.Fatal(err)
		}
		if *topSims {
			if err := writeTopSims(*chartsDir, scan, energyCard, financeCard, *runs, *shiftKeep, export); err != nil {
				log.Fatal(err)
			}
		}
		fmt.Fprintf(out, "Charts: %s/\n", *chartsDir)
		if *topSims && len(scan.Shifts) > 0 {
			last := scan.Shifts[len(scan.Shifts)-1].Shift
			fmt.Fprintf(out, "Top-Sims: %s/top_sims_schicht_%d/\n", *chartsDir, last)
		}
	}

	if *outFile != "" {
		if err := writeYAML(*outFile, scan, opts, *from, *to, *top, *outFull); err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(out, "\nVollstaendige Ergebnisse: %s\n", *outFile)
	}

	if *chartsDir != "" {
		reportPath := filepath.Join(*chartsDir, "report.txt")
		if err := os.WriteFile(reportPath, []byte(reportBuf.String()), 0o644); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Report: %s\n", reportPath)
	}

	if opts.SpillDir != "" {
		if err := scan.Cleanup(); err != nil {
			log.Printf("Spill-Verzeichnis aufraeumen: %v", err)
		}
	}
}

func energyCardIDs() string {
	ids := make([]string, len(energy.Cards))
	for i, c := range energy.Cards {
		ids[i] = c.ID
	}
	return strings.Join(ids, ", ")
}

func financeCardIDs() string {
	ids := make([]string, len(finance.Cards))
	for i, c := range finance.Cards {
		ids[i] = c.ID
	}
	return strings.Join(ids, ", ")
}

func printShiftBlock(w io.Writer, sr seedsearch.ShiftResult, card energy.Card, top, runs int) {
	d := card.ShiftDemands(sr.Shift)
	title := fmt.Sprintf("################  SCHICHT %d — Bedarf I=%d W=%d b=%d R=%d  ################",
		sr.Shift, d.Industry, d.Residential, d.Rail, d.Plant)
	fmt.Fprintln(w, title)
	fmt.Fprintln(w)

	o, err := sr.LoadOutcomes()
	if err != nil {
		fmt.Fprintf(w, "(Fehler beim Laden der Ergebnisse: %v)\n\n", err)
		return
	}
	printTable(w, "Top Seeds: Alle Bedarfe erfuellt und beliebiger Schaden", seedsearch.TopWins(o, top), runs, winCols)
	printTable(w, "Top Seeds: Alle Bedarfe erfuellt und kein Schaden", seedsearch.TopAllDemandsNoDamage(o, top), runs, allDemandsNoDamageCols)
	printTable(w, "Top Seeds: Max. 1 Bedarf nicht erfuellt und kein Schaden", seedsearch.TopMax1DemandNoDamage(o, top), runs, max1DemandNoDamageCols)
	printTable(w, "Top Seeds: Max. 1 Bedarf nicht erfuellt und maximal 1 Schaden", seedsearch.TopMax1DemandMax1Damage(o, top), runs, max1DemandMax1DamageCols)
	printTable(w, "Top Seeds nach Loops (Schrittlimit)", seedsearch.TopLoops(o, top), runs, loopCols)
}

func prepareChartsDir(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := os.RemoveAll(filepath.Join(dir, e.Name())); err != nil {
			return err
		}
	}
	return nil
}

func writeCharts(dir string, scan seedsearch.ScanResult, runs int) error {
	multi := len(scan.Shifts) > 1
	for _, sr := range scan.Shifts {
		outcomes, err := sr.LoadOutcomes()
		if err != nil {
			return err
		}
		sub := shiftOutDir(dir, sr.Shift, multi)
		if err := charts.WriteSeedsearchCharts(sub, outcomes, runs); err != nil {
			return err
		}
	}
	return nil
}

func shiftOutDir(dir string, shift int, multiShift bool) string {
	if multiShift {
		return filepath.Join(dir, fmt.Sprintf("schicht%d", shift))
	}
	return dir
}

type col struct {
	title string
	width int
	value func(seedsearch.Outcome) string
}

var avgStepsCol = col{"ø Schritte", 9, func(o seedsearch.Outcome) string { return o.AvgStepsSummary() }}

var boardFpCol = col{"Board", 5, func(o seedsearch.Outcome) string { return o.BoardFingerprint }}

var prevBoardFpCol = col{"Vorschicht-Board", 16, func(o seedsearch.Outcome) string {
	if o.PrevBoardFingerprint == "" {
		return "-"
	}
	return o.PrevBoardFingerprint
}}

var avgSavedCol = col{"ø Gespart", 9, func(o seedsearch.Outcome) string { return o.AvgSavedSummary() }}
var avgRepairCol = col{"ø Reparatur", 11, func(o seedsearch.Outcome) string { return o.AvgRepairSummary() }}

var winCols = []col{
	{"Seed", 10, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.Seed) }},
	boardFpCol,
	prevBoardFpCol,
	{"P1", 4, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.BoardCosts.Player1) }},
	{"P2", 4, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.BoardCosts.Player2) }},
	avgSavedCol,
	avgRepairCol,
	{"Wins", 6, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.Wins) }},
	{"Win%", 7, func(o seedsearch.Outcome) string { return fmt.Sprintf("%.1f", o.WinRate()*100) }},
	{"Loops", 6, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.Loops) }},
	avgStepsCol,
	{"Kritisch P1", 11, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.CriticalP1) }},
	{"Kritisch P2", 11, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.CriticalP2) }},
	{"End-Bedarf", 16, func(o seedsearch.Outcome) string { return o.EndDemandSummary() }},
	{"End-Schaden", 14, func(o seedsearch.Outcome) string { return o.EndDamageSummary() }},
}

var loopCols = []col{
	{"Seed", 10, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.Seed) }},
	boardFpCol,
	prevBoardFpCol,
	{"P1", 4, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.BoardCosts.Player1) }},
	{"P2", 4, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.BoardCosts.Player2) }},
	avgSavedCol,
	avgRepairCol,
	{"Loops", 6, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.Loops) }},
	{"Loop%", 7, func(o seedsearch.Outcome) string { return fmt.Sprintf("%.1f", o.LoopRate()*100) }},
	{"Wins", 6, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.Wins) }},
	avgStepsCol,
	{"Kritisch P1", 11, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.CriticalP1) }},
	{"Kritisch P2", 11, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.CriticalP2) }},
	{"End-Bedarf", 16, func(o seedsearch.Outcome) string { return o.EndDemandSummary() }},
	{"End-Schaden", 14, func(o seedsearch.Outcome) string { return o.EndDamageSummary() }},
}

var allDemandsNoDamageCols = scoreCols("Treffer", func(o seedsearch.Outcome) int { return o.AllDemandsNoDamage })
var max1DemandNoDamageCols = scoreCols("Treffer", func(o seedsearch.Outcome) int { return o.Max1DemandNoDamage })
var max1DemandMax1DamageCols = scoreCols("Treffer", func(o seedsearch.Outcome) int { return o.Max1DemandMax1Damage })

func scoreCols(scoreTitle string, score func(seedsearch.Outcome) int) []col {
	return []col{
		{"Seed", 10, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.Seed) }},
		boardFpCol,
		prevBoardFpCol,
		{"P1", 4, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.BoardCosts.Player1) }},
		{"P2", 4, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.BoardCosts.Player2) }},
		avgSavedCol,
		avgRepairCol,
		{scoreTitle, 7, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", score(o)) }},
		{"%", 7, func(o seedsearch.Outcome) string {
			if o.Runs == 0 {
				return "0.0"
			}
			return fmt.Sprintf("%.1f", float64(score(o))/float64(o.Runs)*100)
		}},
		{"Wins", 6, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.Wins) }},
		{"Loops", 6, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.Loops) }},
		avgStepsCol,
		{"Kritisch P1", 11, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.CriticalP1) }},
		{"Kritisch P2", 11, func(o seedsearch.Outcome) string { return fmt.Sprintf("%d", o.CriticalP2) }},
		{"End-Bedarf", 16, func(o seedsearch.Outcome) string { return o.EndDemandSummary() }},
		{"End-Schaden", 14, func(o seedsearch.Outcome) string { return o.EndDamageSummary() }},
	}
}

func printTable(w io.Writer, title string, rows []seedsearch.Outcome, runs int, cols []col) {
	fmt.Fprintln(w, title)
	fmt.Fprintln(w, strings.Repeat("-", len(title)))

	// Size each column to the widest of its title and cell values so that the
	// variable-length fingerprints stay aligned without excessive padding.
	widths := make([]int, len(cols))
	for i, c := range cols {
		widths[i] = c.width
		if len(c.title) > widths[i] {
			widths[i] = len(c.title)
		}
		for _, row := range rows {
			if v := len(c.value(row)); v > widths[i] {
				widths[i] = v
			}
		}
	}

	header := make([]string, len(cols))
	under := make([]string, len(cols))
	for i, c := range cols {
		header[i] = fmt.Sprintf("%-*s", widths[i], c.title)
		under[i] = strings.Repeat("-", widths[i])
	}
	fmt.Fprintln(w, strings.Join(header, " "))
	fmt.Fprintln(w, strings.Join(under, " "))
	for _, row := range rows {
		cells := make([]string, len(cols))
		for i, c := range cols {
			cells[i] = fmt.Sprintf("%-*s", widths[i], c.value(row))
		}
		fmt.Fprintln(w, strings.Join(cells, " "))
	}
	if len(rows) == 0 {
		fmt.Fprintln(w, "(keine)")
	}
	fmt.Fprintf(w, "(%d Laeufe pro Seed)\n\n", runs)
}

type reportYAML struct {
	From     int64        `yaml:"from"`
	To       int64        `yaml:"to"`
	Runs     int          `yaml:"runs_per_seed"`
	Top      int          `yaml:"top"`
	Settings settingsYAML `yaml:"settings"`
	Shifts   []shiftYAML  `yaml:"shifts"`
}

type settingsYAML struct {
	EnergyCard     string `yaml:"energy_card"`
	FinanceCard    string `yaml:"finance_card"`
	FinanceReactor int    `yaml:"finance_reactor_budget"`
	FinanceGrid    int    `yaml:"finance_grid_budget"`
	Shifts         int    `yaml:"shifts"`
	ShiftKeep      int    `yaml:"shift_keep"`
	MonthFilter    int    `yaml:"month_filter,omitempty"`
	StartBoard     string `yaml:"start_board,omitempty"`
}

type shiftYAML struct {
	Shift                   int            `yaml:"shift"`
	Demand                  zoneTotalsYAML `yaml:"demand"`
	Outcomes                []outcomeYAML  `yaml:"outcomes"`
	TopWins                 []outcomeYAML  `yaml:"top_wins"`
	TopAllDemandsNoDamage   []outcomeYAML  `yaml:"top_all_demands_no_damage"`
	TopMax1DemandNoDamage   []outcomeYAML  `yaml:"top_max1_demand_no_damage"`
	TopMax1DemandMax1Damage []outcomeYAML  `yaml:"top_max1_demand_max1_damage"`
	TopLoops                []outcomeYAML  `yaml:"top_loops"`
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

type leftoverYAML struct {
	Reaktor   int `yaml:"reaktor"`
	Stromnetz int `yaml:"stromnetz"`
}

type outcomeYAML struct {
	Seed                   int64          `yaml:"seed"`
	Shift                  int            `yaml:"shift"`
	BoardFingerprint       string         `yaml:"board_fingerprint"`
	PrevBoardFingerprint   string         `yaml:"prev_board_fingerprint,omitempty"`
	CarryBoardFingerprint  string         `yaml:"carry_board_fingerprint,omitempty"`
	BoardCosts             costsYAML      `yaml:"board_costs"`
	StartLeftover          leftoverYAML   `yaml:"start_leftover,omitempty"`
	EndLeftover            leftoverYAML   `yaml:"end_leftover"`
	Wins                   int            `yaml:"wins"`
	AllDemandsNoDamage   int            `yaml:"all_demands_no_damage"`
	Max1DemandNoDamage   int            `yaml:"max1_demand_no_damage"`
	Max1DemandMax1Damage int            `yaml:"max1_demand_max1_damage"`
	Loops                int            `yaml:"loops"`
	CriticalP1           int            `yaml:"critical_p1"`
	CriticalP2           int            `yaml:"critical_p2"`
	AvgEndDemand         zoneTotalsYAML `yaml:"avg_end_demand"`
	AvgEndDamage         zoneTotalsYAML `yaml:"avg_end_damage"`
	AvgSaved             leftoverYAML   `yaml:"avg_saved"`
	AvgRepairSpent       float64        `yaml:"avg_repair_spent"`
	AvgSteps             float64        `yaml:"avg_steps"`
	WinRate              float64        `yaml:"win_rate"`
	LoopRate             float64        `yaml:"loop_rate"`
}

func writeYAML(path string, scan seedsearch.ScanResult, opts seedsearch.Options, from, to int64, top int, fullOutcomes bool) error {
	doc := reportYAML{
		From: from,
		To:   to,
		Runs: opts.Runs,
		Top:  top,
		Settings: settingsYAML{
			EnergyCard:     opts.EnergyCard.ID,
			FinanceCard:    opts.Finance.ID,
			FinanceReactor: opts.Finance.ReactorBudget,
			FinanceGrid:    opts.Finance.GridBudget,
			Shifts:         opts.Shifts,
			ShiftKeep:      opts.ShiftKeep,
			MonthFilter:    opts.MonthFilter,
			StartBoard:     opts.StartBoardFingerprint,
		},
	}
	for _, sr := range scan.Shifts {
		outcomes, err := sr.LoadOutcomes()
		if err != nil {
			return err
		}
		entry := shiftYAML{
			Shift:                   sr.Shift,
			Demand:                  shiftDemandsYAML(opts.EnergyCard.ShiftDemands(sr.Shift)),
			TopWins:                 toOutcomeYAML(seedsearch.TopWins(outcomes, top)),
			TopAllDemandsNoDamage:   toOutcomeYAML(seedsearch.TopAllDemandsNoDamage(outcomes, top)),
			TopMax1DemandNoDamage:   toOutcomeYAML(seedsearch.TopMax1DemandNoDamage(outcomes, top)),
			TopMax1DemandMax1Damage: toOutcomeYAML(seedsearch.TopMax1DemandMax1Damage(outcomes, top)),
			TopLoops:                toOutcomeYAML(seedsearch.TopLoops(outcomes, top)),
		}
		if fullOutcomes {
			entry.Outcomes = toOutcomeYAML(outcomes)
		}
		doc.Shifts = append(doc.Shifts, entry)
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
			Seed:                  o.Seed,
			Shift:                 o.Shift,
			BoardFingerprint:      o.BoardFingerprint,
			PrevBoardFingerprint:  o.PrevBoardFingerprint,
			CarryBoardFingerprint: o.CarryBoardFingerprint,
			BoardCosts: costsYAML{
				Reaktor:   o.BoardCosts.Player1,
				Stromnetz: o.BoardCosts.Player2,
				Total:     o.BoardCosts.Total(),
			},
			StartLeftover: leftoverYAML{
				Reaktor:   o.StartLeftover.Player1,
				Stromnetz: o.StartLeftover.Player2,
			},
			EndLeftover: leftoverYAML{
				Reaktor:   o.EndLeftover.Player1,
				Stromnetz: o.EndLeftover.Player2,
			},
			Wins:                 o.Wins,
			AllDemandsNoDamage:   o.AllDemandsNoDamage,
			Max1DemandNoDamage:   o.Max1DemandNoDamage,
			Max1DemandMax1Damage: o.Max1DemandMax1Damage,
			Loops:                o.Loops,
			CriticalP1:           o.CriticalP1,
			CriticalP2:           o.CriticalP2,
			AvgEndDemand:         toZoneTotalsYAML(o.AvgEndDemand),
			AvgEndDamage:         toZoneTotalsYAML(o.AvgEndDamage),
			AvgSaved: leftoverYAML{
				Reaktor:   int(math.Round(o.AvgSavedP1)),
				Stromnetz: int(math.Round(o.AvgSavedP2)),
			},
			AvgRepairSpent:       o.AvgRepairSpent,
			AvgSteps:             o.AvgSteps,
			WinRate:              o.WinRate(),
			LoopRate:             o.LoopRate(),
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

func shiftDemandsYAML(d board.ShiftDemands) zoneTotalsYAML {
	return zoneTotalsYAML{
		Industry:    float64(d.Industry),
		Residential: float64(d.Residential),
		Rail:        float64(d.Rail),
		Plant:       float64(d.Plant),
	}
}

func init() {
	log.SetOutput(os.Stderr)
}
