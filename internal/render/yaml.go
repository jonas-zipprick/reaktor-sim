package render

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
	"gopkg.in/yaml.v3"
)

func writeYAML(path string, v any) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Errorf("yaml marshal: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

type costsYAML struct {
	Reaktor   int `yaml:"reaktor"`
	Stromnetz int `yaml:"stromnetz"`
	Total     int `yaml:"total"`
}

type demandYAML struct {
	Zone   string `yaml:"zone"`
	Letter string `yaml:"letter"`
	Count  int    `yaml:"count"`
}

type cellYAML struct {
	Q      int    `yaml:"q"`
	R      int    `yaml:"r"`
	Symbol string `yaml:"symbol,omitempty"`
	Charge string `yaml:"charge,omitempty"`
	Field  string `yaml:"field,omitempty"`
	Demand string `yaml:"demand,omitempty"`
}

type damageYAML struct {
	Zone   string `yaml:"zone"`
	Letter string `yaml:"letter"`
	Count  int    `yaml:"count"`
}

type boardYAML struct {
	Seed      int64        `yaml:"seed,omitempty"`
	PrevBoard string       `yaml:"prev_board,omitempty"`
	Costs     costsYAML    `yaml:"costs"`
	Demands []demandYAML `yaml:"demands"`
	Damage  []damageYAML `yaml:"damage,omitempty"`
	Cells   []cellYAML   `yaml:"cells"`
	Legend  []string     `yaml:"legend"`
}

func buildBoardYAML(state *board.State, meta BoardMeta) boardYAML {
	costs := state.PlayerCosts()
	doc := boardYAML{
		Seed:      meta.Seed,
		PrevBoard: meta.PrevBoard,
		Costs: costsYAML{
			Reaktor:   costs.Player1,
			Stromnetz: costs.Player2,
			Total:     costs.Total(),
		},
		Legend: Legend(),
	}
	for _, z := range []board.Zone{
		board.ZoneIndustry,
		board.ZoneResidential,
		board.ZoneRail,
		board.ZonePlant,
	} {
		if n := state.TotalDemand(z); n > 0 {
			doc.Demands = append(doc.Demands, demandYAML{
				Zone:   z.String(),
				Letter: board.ZoneLetter(z),
				Count:  n,
			})
		}
		if n := state.TotalDamage(z); n > 0 {
			doc.Damage = append(doc.Damage, damageYAML{
				Zone:   z.String(),
				Letter: board.ZoneLetter(z),
				Count:  n,
			})
		}
	}
	if n := state.EmitterDamage; n > 0 {
		doc.Damage = append(doc.Damage, damageYAML{
			Zone:   "Zünder",
			Letter: "Z",
			Count:  n,
		})
	}
	for _, c := range hex.AllBoardCoords {
		if !c.Valid() {
			continue
		}
		tile := state.Tiles[c.Q][c.R]
		sym := Label(state, c)
		ch := ChargeLabel(tile)
		dem := state.DemandLabel(c)
		if sym == "" && ch == "" && dem == "" && tile.Type == field.Empty {
			continue
		}
		entry := cellYAML{Q: c.Q, R: c.R, Symbol: sym, Charge: ch, Demand: dem}
		if tile.Type != field.Empty {
			if info, ok := field.Catalog[tile.Type]; ok {
				entry.Field = info.Name
			}
		}
		doc.Cells = append(doc.Cells, entry)
	}
	return doc
}

type traceStepYAML struct {
	Index     int    `yaml:"index"`
	Step      int    `yaml:"step"`
	Event     string `yaml:"event"`
	Narrative string `yaml:"narrative"`
	QueueSize int    `yaml:"queue_size"`
}

type traceEnergyYAML struct {
	ID    string `yaml:"id"`
	Name  string `yaml:"name"`
	Level int    `yaml:"level"`
}

type traceSetupYAML struct {
	EnergyCard traceEnergyYAML `yaml:"energy_card"`
	Shift      int             `yaml:"shift"`
	Costs      costsYAML       `yaml:"costs"`
}

type traceYAML struct {
	Run           int             `yaml:"run"`
	MonteCarloRun int             `yaml:"monte_carlo_run,omitempty"`
	Setup         traceSetupYAML  `yaml:"setup"`
	Steps         int             `yaml:"steps"`
	Outcome       string          `yaml:"outcome,omitempty"`
	Events        []traceStepYAML `yaml:"events"`
}

func buildTraceSetupYAML(meta TraceMeta) traceSetupYAML {
	return traceSetupYAML{
		EnergyCard: traceEnergyYAML{
			ID:    meta.EnergyCardID,
			Name:  meta.EnergyCardName,
			Level: meta.EnergyCardLevel,
		},
		Shift: meta.Shift,
		Costs: costsYAML{
			Reaktor:   meta.Costs.Player1,
			Stromnetz: meta.Costs.Player2,
			Total:     meta.Costs.Total(),
		},
	}
}

func buildTraceYAML(run int, monteCarloRun int, meta TraceMeta, snaps []sim.Snapshot) traceYAML {
	doc := traceYAML{
		Run:           run,
		MonteCarloRun: monteCarloRun,
		Setup:         buildTraceSetupYAML(meta),
		Steps:         len(snaps),
		Events:        make([]traceStepYAML, 0, len(snaps)),
	}
	if len(snaps) == 0 {
		return doc
	}
	last := snaps[len(snaps)-1]
	doc.Outcome = traceOutcomeNote(last.Event, last.Board)
	for i, snap := range snaps {
		doc.Events = append(doc.Events, traceStepYAML{
			Index:     i,
			Step:      snap.Step,
			Event:     snap.Event,
			Narrative: ASCII(snap.Narrative),
			QueueSize: snap.QueueSize,
		})
	}
	return doc
}

type traceIndexRunYAML struct {
	Run           int    `yaml:"run"`
	Kind          string `yaml:"kind"`
	MonteCarloRun int    `yaml:"monte_carlo_run,omitempty"`
	Steps         int    `yaml:"steps"`
	Directory     string `yaml:"directory"`
	TraceFile     string `yaml:"trace_file"`
}

type traceIndexYAML struct {
	Runs []traceIndexRunYAML `yaml:"runs"`
}

func buildTraceIndexYAML(entries []TraceIndexEntry) traceIndexYAML {
	doc := traceIndexYAML{Runs: make([]traceIndexRunYAML, 0, len(entries))}
	for _, e := range entries {
		doc.Runs = append(doc.Runs, traceIndexRunYAML{
			Run:           e.Run,
			Kind:          string(e.Kind),
			MonteCarloRun: e.MonteCarloRun,
			Steps:         e.Steps,
			Directory:     e.Dir,
			TraceFile:     filepath.Join(e.Dir, "trace.yaml"),
		})
	}
	return doc
}
