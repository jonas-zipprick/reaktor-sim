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
| `-trace` | false | Schichten aufzeichnen, Graph pro Schritt in `runN/` |
| `-trace-runs` | 0 (= `-runs`) | Aufgezeichnete Läufe; `0` zeichnet alle `-runs` auf |

## Ausgabe

- `kosten.png` – Gesamtkosten des zufälligen Boards
- `waerme_turbine.png` – Verteilung der Wärme an der Turbine
- `spannung_wohnviertel.png` – Spannung am rechten Rand
- `spannung_industrie.png` – Spannung am oberen Rand
- `spannung_bahn.png` – Spannung am unteren Rand
- `spannung_reaktoreigenbedarf.png` – Spannung am Kraftwerksrand
- `spielfeld.png` / `spielfeld.txt` – Brettdarstellung mit Symbolen und Bedarfen
- `graph.png` / `graph.txt` – Flussgraph (rot=Wärme, grün=Neutron, blau=Spannung)
- `runN/graph_runN_SSS.png` – Graph pro Simulationsschritt (mit `-trace`)
- `runN/trace.txt` – Ereignisprotokoll pro Lauf

### Trace-Modus

```bash
go run ./cmd/sim -seed 1 -trace -out output
```

Mit Standard `-runs 20` entstehen `run1/` … `run20/` mit je `graph_runN_SSS.png`.
Weniger Traces: `-trace -trace-runs 5`. Mehr Statistik: `-runs 5000` (ohne `-trace` schneller).

## Projektstruktur

```
cmd/sim/          CLI-Einstiegspunkt
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
