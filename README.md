# Reaktor-Sim

Monte-Carlo-Simulation für **Reaktor-Architekten** zur Optimierung des Game-Balancings.

## Ablauf

1. **Zufälliger Game State** – Felder werden auf dem Hex-Raster platziert (9 Spalten × 3 Zeilen, siehe `gameRules.md`).
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
- `spielfeld.png` / `spielfeld.txt` – Brettdarstellung mit Symbolen und Bedarfen
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

Separates Meta-Programm zum Durchprobieren vieler Board-Seeds und Finden von Brettern mit vielen Wins oder Loops:

```bash
go run ./cmd/seedsearch -from 1 -to 1000 -runs 100 -top 10
```

Pro Seed wird ein Brett erzeugt und `-runs` Monte-Carlo-Läufe ausgeführt. Am Ende erscheinen zwei Ranglisten (Top Wins / Top Loops). Optional `-out seeds.yaml` für die vollständigen Ergebnisse.

| Flag | Standard | Beschreibung |
|------|----------|--------------|
| `-from` / `-to` | 1 / 1 | Seed-Bereich |
| `-runs` | 100 | Läufe pro Seed |
| `-top` | 10 | Einträge pro Rangliste |
| `-cost-p1` | 0 | Feste Reaktor-Kosten (0 = zufällig) |
| `-cost-p2` | 0 | Feste Stromnetz-Kosten (0 = zufällig) |
| `-out` | — | Vollständige Ergebnisse als YAML |
| `-progress` | true | Fortschrittsbalken auf stderr |

```bash
go run ./cmd/seedsearch -from 1 -to 500 -runs 50 -cost-p1 15 -cost-p2 20 -top 5
```

Simulations-Flags entsprechen `cmd/sim`.

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
