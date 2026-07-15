# Reaktor-Sim

Monte-Carlo-Simulation für **Reaktor-Architekten** zur Optimierung des Game-Balancings.

## Ablauf

1. **Zufälliger Game State** – Felder werden auf dem Hex-Raster platziert (9 Spalten × 5 Zeilen, siehe `gameRules.md`).
2. **Graph** – Aus dem Board werden Knoten (Felder) und gerichtete Kanten mit drei Übergangswahrscheinlichkeiten erzeugt: Wärme, Neutron, Spannung.
3. **Simulation** – Monte-Carlo-Durchläufe starten mit Wärme am Zünder; Chips bewegen sich, Felder reagieren und können ausbrennen (Graph wird neu berechnet).
4. **Charts** – PNG-Histogramme für Kosten, Wärme an der Turbine und Spannung pro Bedarfszone.

## Voraussetzungen

- Go 1.21+

## Ausführen

```bash
go run ./cmd/sim -runs 20 -out output
```

### Optionen

| Flag | Standard | Beschreibung |
|------|----------|--------------|
| `-runs` | 20 | Anzahl Monte-Carlo-Durchläufe (Histogramme/Charts) |
| `-seed` | Zeitstempel | Reproduzierbarer Zufallsseed |
| `-out` | `output` | Verzeichnis für PNG-Charts |
| `-heat` | 1 | Wärme-Chips am Zünder pro Schicht |
| `-trace-first` | -1 (aus) | Erste n Läufe aufzeichnen; `0` = alle `-runs` |
| `-trace-loop` | 0 (aus) | Schrittlimit-Läufe aufzeichnen (max. n Stück) |
| `-trace-win` | 0 (aus) | Gewonnene Läufe aufzeichnen (alle Bedarfe erfüllt, max. n) |

## Ausgabe

- `kosten.png` – Gesamtkosten des zufälligen Boards
- `waerme_turbine.png` – Verteilung der Wärme an der Turbine
- `spannung_wohnviertel.png` – Spannung am rechten Rand
- `spannung_industrie.png` – Spannung am oberen Rand
- `spannung_bahn.png` – Spannung am unteren Rand
- `spannung_reaktoreigenbedarf.png` – Spannung am Kraftwerksrand
- `spielfeld-<board-fingerprint>.png` / `spielfeld-<board-fingerprint>.yaml` – Brettdarstellung mit Symbolen und Bedarfen
- `graph.png` / `graph.txt` – Flussgraph (rot=Wärme, grün=Neutron, blau=Spannung)
- `runN/graph_runN_SSS.png` – Graph pro Simulationsschritt (mit `-trace-first`)
- `runN/trace.yaml` – Ereignisprotokoll pro Lauf

### Trace-Modus

```bash
go run ./cmd/sim -seed 1 -trace-first 0 -out output
```

Mit `-trace-first 0` werden alle `-runs` Läufe aufgezeichnet (`run1/` … `runN/`).
Weniger Traces: `-trace-first 5`. Mehr Statistik: `-runs 5000` (ohne Trace-Flags schneller).

## Seed-Suche

Separates Meta-Programm zum Durchprobieren vieler Board-Seeds und Finden guter Bretter. Bedarfe kommen aus dem Schichtplan der **Energie-Karte**, das Schicht-Budget aus der **Finanz-Karte**:

```bash
go run ./cmd/seedsearch -from 1 -to 1000 -runs 100 -top 10
```

Pro Seed wird ein Brett erzeugt (mit dem Schicht-Budget der Finanz-Karte) und `-runs` Monte-Carlo-Läufe ausgeführt. Doppelte Startbretter werden automatisch ausgefiltert. Am Ende erscheinen pro Schicht mehrere Ranglisten. Optional `-out seeds.yaml` für die vollständigen Ergebnisse.

| Flag | Standard | Beschreibung |
|------|----------|--------------|
| `-from` / `-to` | 1 / 1 | Seed-Bereich |
| `-runs` | 100 | Läufe pro Seed |
| `-top` | 10 | Einträge pro Rangliste |
| `-energie-karte` | `eroeffnungsfeier` | Energie-Karte (Schichtplan/Bedarfe) |
| `-finanz-karte` | `schwerindustrie` | Finanz-Karte (Schicht-Budget) |
| `-schichten` | 1 | Anzahl aufeinanderfolgender Schichten (1-5, ganzer Monat = 5) |
| `-schicht-keep` | 1 | Top-Boards je Erfolgs-Rangliste, die in die nächste Schicht verzweigen |
| `-month-filter` | 0 | Kampagnenmonat für Feldverfügbarkeit |
| `-start-board` | — | Board-Fingerprint als Startbrett (Folgemonat) |
| `-charts-dir` | `seedsearch-out` | Ausgabe für Chart-PNGs und `report.txt` |
| `-out` | — | Vollständige Ergebnisse als YAML |
| `-progress` | true | Fortschrittsbalken auf stderr |

### Ganzen Monat simulieren

Mit `-schichten n` werden `n` Schichten hintereinander gerechnet. Schicht 1 baut pro Seed ein Brett; jede Folgeschicht nimmt die Top-Boards der vorherigen Schicht (`-schicht-keep` je Rangliste aus 4 Erfolgs-Tabellen; Loops-Tabellen-Einträge werden übersprungen und durch nächste Erfolgs-Boards ersetzt) und probiert sie mit **allen Seeds** erneut durch (Kauf + Simulation). Ungestillte Bedarfe und Schaden werden als Median über die Läufe kumulierend übernommen.

```bash
go run ./cmd/seedsearch -from 1 -to 500 -runs 50 -schichten 5 -schicht-keep 2 -energie-karte schturmowschtschina -finanz-karte sparmassnahmen -top 5
```

`cmd/sim` behält dagegen die Einzel-Flags `-demand-*`, `-damage-*` und `-prev-board`.

### Folgemonat mit festem Startbrett

Mit `-start-board` lädt Schicht 1 ein bestehendes Brett per Fingerprint (z. B. aus `spielfeld-*.yaml` oder `carry_board_fingerprint` eines Seedsearch-Ergebnisses). Pro Seed werden nur noch Kaufvarianten mit dem Finanz-Budget ausprobiert; ab Schicht 2 verzweigt der Lauf wie gewohnt.

```bash
go run ./cmd/seedsearch -from 1 -to 500 -runs 100 \
  -start-board b2_AQIIAAAAAwMABAAABAoAAwAABwkIAAAACgICAAAADQkCAAAAEAwAAQAAFAQAAAAA \
  -energie-karte netzoptimierung -finanz-karte sparmassnahmen \
  -schichten 5 -month-filter 2
```

## Projektstruktur

```
cmd/sim/          Monte-Carlo-Simulator mit Charts und Trace
cmd/seedsearch/   Seed-Suche (Wins/Loops über viele Bretter)
internal/
  board/          Game State & Zufallsgenerator
  field/          Feldtypen & Kosten
  hex/            Hex-Koordinaten
  graph/          Kanten-Wahrscheinlichkeiten
  sim/            Monte-Carlo-Simulation
  stats/          Histogramme
  charts/         PNG-Erzeugung (gonum/plot)
gameRules.md      Spielregeln
```

## Hinweise

Die Simulation vereinfacht einige Spielerentscheidungen (z. B. zufällige Schussrichtungen, automatisches Abfeuern von Spannung an der Turbine). Für feinere Balancing-Tests können Feldwahrscheinlichkeiten und Spielerstrategien in `internal/sim` erweitert werden.
