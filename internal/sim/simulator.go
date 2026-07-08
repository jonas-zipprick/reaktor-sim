// Package sim runs Monte-Carlo simulations of one reactor shift.
package sim

import (
	"math/rand"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/energy"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/graph"
	"github.com/jonas/reaktor-sim/internal/hex"
)

// ChipType mirrors graph particle types for the event queue.
type ChipType int

const (
	ChipHeat ChipType = iota
	ChipNeutron
	ChipVoltage
)

// Chip is a particle in flight.
type Chip struct {
	Type ChipType
	Pos  hex.Coord
	Dir  int
}

// MaxStepsPerRun is the hard cap on chip resolutions per simulation run.
const MaxStepsPerRun = 100

// Result holds metrics from one simulation run.
type Result struct {
	HeatAtTurbine     int
	ZoneDeliveries    [4]int
	CriticalFailure   bool
	StepLimitExceeded bool
}

// Snapshot captures board and graph state after a simulation step.
type Snapshot struct {
	Step      int
	Event     string
	Narrative string
	Board     *board.State
	Graph     *graph.Graph
	Queue     []Chip
	Active    *Chip
	QueueSize int
}

// Config controls simulation parameters.
type Config struct {
	MaxSteps      int
	CriticalLimit int
	InitialHeat            int
	InitialNeutron         int
	MixedEmitterTrigger    bool // one random heat/neutron trigger per run (game rules)
	StartDir               int // 0-5 travel override for tests, or -1 = random NO/O/SO per chip
	ShiftDemands  board.ShiftDemands
	EnergyCard    energy.Card // when set, overrides ShiftDemands per run
	Shift         int         // 1-5 for EnergyCard
	RandomShift   bool        // pick shift 1-5 per Monte-Carlo run
	Trace         bool
	InitialChips  []Chip
}

func DefaultConfig() Config {
	card := energy.DefaultCard()
	return Config{
		MaxSteps:      MaxStepsPerRun,
		CriticalLimit: 8,
		InitialHeat:   1,
		EnergyCard:    card,
		Shift:         1,
		ShiftDemands:  card.ShiftDemands(1),
	}
}

type engine struct {
	board *board.State
	graph *graph.Graph
	queue []Chip
	res   Result
	rng   *rand.Rand
	cfg   Config
	trace []Snapshot
	step  int
	lastResolved Chip
	hasResolved  bool
	lostRecorded bool
}

// Run executes one full shift simulation on a board clone.
func Run(state *board.State, rng *rand.Rand, cfg Config) Result {
	res, _ := run(state, rng, cfg)
	return res
}

// RunTrace executes one shift and records snapshots after each chip resolution.
func RunTrace(state *board.State, rng *rand.Rand, cfg Config) (Result, []Snapshot) {
	cfg.Trace = true
	return run(state, rng, cfg)
}

func run(state *board.State, rng *rand.Rand, cfg Config) (Result, []Snapshot) {
	e := &engine{
		board: state.Clone(),
		rng:   rng,
		cfg:   cfg,
	}
	e.board.ApplyDemands(cfg.demandsForRun(e.rng))
	e.queue = make([]Chip, 0, 64)
	e.queue = append(e.queue, EmitterChips(cfg, e.rng)...)
	e.queue = append(e.queue, cfg.InitialChips...)
	e.graph = graph.BuildFlow(e.board, e.inFlight())

	e.record("Start")
	if e.checkCritical() {
		return e.res, finalizeTrace(e.trace)
	}
	for !e.res.CriticalFailure && e.step < cfg.MaxSteps {
		if len(e.queue) > 0 {
			e.step++
			chip := e.queue[0]
			e.queue = e.queue[1:]
			e.resolve(chip)
			if e.res.CriticalFailure {
				break
			}
			if len(e.queue) > 0 {
				e.maybeVoluntaryFire()
			}
			continue
		}
		if !e.maybeVoluntaryFire() {
			break
		}
	}
	if len(e.queue) > 0 && e.step >= cfg.MaxSteps && !e.res.CriticalFailure {
		e.abortStepLimit()
	}
	if !e.res.CriticalFailure && !e.res.StepLimitExceeded {
		e.record("Ende")
	}

	return e.res, finalizeTrace(e.trace)
}

func finalizeTrace(snaps []Snapshot) []Snapshot {
	for i, s := range snaps {
		if s.Event == "verloren" || s.Event == "Schrittlimit" || s.Event == "Ende" {
			return snaps[:i+1]
		}
	}
	return snaps
}

func (e *engine) record(event string) {
	e.recordWithActive(event, nil)
}

func (e *engine) recordWithActive(event string, active *Chip) {
	if !e.cfg.Trace {
		return
	}
	if e.res.CriticalFailure && event != "verloren" {
		return
	}
	if e.res.StepLimitExceeded && event != "Schrittlimit" {
		return
	}
	queue := append([]Chip{}, e.queue...)
	var activeCopy *Chip
	if active != nil {
		c := *active
		activeCopy = &c
	}
	var narrative string
	if event == "Start" || event == "verloren" || event == "Schrittlimit" || event == "Ende" {
		narrative = narrate(event, nil, e.board, queue)
	} else if active != nil {
		narrative = narrate(event, active, e.board, queue)
	} else if e.hasResolved {
		c := e.lastResolved
		narrative = narrate(event, &c, e.board, queue)
	} else {
		narrative = narrate(event, nil, e.board, queue)
	}
	e.trace = append(e.trace, Snapshot{
		Step:      e.step,
		Event:     event,
		Narrative: narrative,
		Board:     e.board.Clone(),
		Graph:     graph.BuildFlow(e.board, e.inFlight()),
		Queue:     queue,
		Active:    activeCopy,
		QueueSize: len(e.queue),
	})
}

func (e *engine) recordStep(event string, active *Chip) {
	if e.res.CriticalFailure {
		return
	}
	if active != nil {
		if e.checkCritical(*active) {
			return
		}
	} else if e.checkCritical() {
		return
	}
	e.recordWithActive(event, active)
}

func (e *engine) loseGame() {
	if e.lostRecorded {
		e.res.CriticalFailure = true
		e.queue = nil
		return
	}
	e.res.CriticalFailure = true
	e.queue = nil
	e.lostRecorded = true
	e.record("verloren")
}

func (e *engine) abortStepLimit() {
	e.res.StepLimitExceeded = true
	e.queue = nil
	e.record("Schrittlimit")
}

func (e *engine) checkCritical(inFlight ...Chip) bool {
	if e.res.CriticalFailure {
		e.queue = nil
		return true
	}
	if !criticalExceeded(e.board, e.queue, inFlight, e.cfg.CriticalLimit) {
		return false
	}
	e.loseGame()
	return true
}

func (e *engine) inFlight() []graph.InFlight {
	out := make([]graph.InFlight, len(e.queue))
	for i, c := range e.queue {
		out[i] = graph.InFlight{
			Particle: chipParticle(c.Type),
			Pos:      c.Pos,
			Dir:      c.Dir,
		}
	}
	return out
}

func chipParticle(t ChipType) graph.ParticleType {
	switch t {
	case ChipNeutron:
		return graph.Neutron
	case ChipVoltage:
		return graph.Voltage
	default:
		return graph.Heat
	}
}

func (e *engine) resolve(chip Chip) {
	if e.res.CriticalFailure {
		return
	}
	e.lastResolved = chip
	e.hasResolved = true
	if e.checkCritical(chip) {
		return
	}

	nextPos := chip.Pos.Neighbor(chip.Dir)

	if nextPos.IsEmitter() {
		e.queue = append(e.queue, e.deflectEmitter(chip)...)
		e.recordStep("Zuender-Abpraller", nil)
		return
	}

	if blocked := hex.BlockedBoundary(chip.Pos, nextPos, chip.Dir); blocked != hex.BoundaryNone {
		e.handleBlocked(chip, blocked)
		return
	}

	if nextPos.IsTurbine() {
		event, active := e.processTurbine(chip)
		e.recordStep(event, active)
		return
	}

	tile := &e.board.Tiles[nextPos.Q][nextPos.R]
	if tile.Type == field.Empty {
		e.queue = append(e.queue, Chip{Type: chip.Type, Pos: nextPos, Dir: chip.Dir})
		e.recordStep("Leeres Feld", nil)
		return
	}
	if tile.BurnedOut {
		e.recordStep("Ausgebrannt", &chip)
		return
	}

	incoming := (chip.Dir + 3) % 6

	destroyedEmergencyGen := tile.Type == field.EmergencyGenerator && chip.Type == ChipVoltage
	newChips, graphChanged := react(e.board, e.graph, nextPos, tile, chip, incoming, e.rng)
	if destroyedEmergencyGen {
		e.purgeChipsAt(nextPos)
		newChips = nil
		graphChanged = true
	} else {
		e.queue = append(e.queue, newChips...)
	}
	if graphChanged {
		graph.Rebuild(e.graph, e.board, e.inFlight())
	}
	event := "Feldreaktion"
	if destroyedEmergencyGen {
		event = "Notgenerator zerstoert"
	}
	e.recordStep(event, &chip)
}

func (e *engine) handleBlocked(chip Chip, kind hex.BoundaryKind) {
	switch kind {
	case hex.BoundaryInternalWall:
		e.recordStep("Innere Wand", &chip)
		return
	case hex.BoundaryOuter:
		e.handleOuterBoundary(chip)
	}
}

func (e *engine) handleOuterBoundary(chip Chip) {
	switch chip.Type {
	case ChipHeat:
		if chip.Pos.IsPlayer1() {
			reflected := Chip{
				Type: ChipHeat,
				Pos:  chip.Pos,
				Dir:  hex.ReflectOffOuterWall(chip.Dir),
			}
			e.queue = append(e.queue, reflected)
			e.recordStep("Waerme-Reflektion", &reflected)
		} else {
			e.recordStep("Waerme verpufft", &chip)
		}
	case ChipNeutron:
		e.recordStep("Neutron verpufft", &chip)
	case ChipVoltage:
		if !chip.Pos.IsPlayer2() {
			e.recordStep("Spannung verpufft", &chip)
			return
		}
		if z, ok := e.tryVoltageWallDelivery(chip.Pos, chip.Dir); ok {
			e.recordStep(board.BorderDemandEvent(z), &chip)
			return
		}
		reflected := Chip{
			Type: ChipVoltage,
			Pos:  chip.Pos,
			Dir:  hex.ReflectOffOuterWall(chip.Dir),
		}
		e.queue = append(e.queue, reflected)
		e.lastResolved = reflected
		e.hasResolved = true
		e.recordStep("Spannungs-Spike", nil)
	}
}

func (e *engine) tryVoltageWallDelivery(from hex.Coord, travelDir int) (board.Zone, bool) {
	z, ok := e.board.TryConsumeWallDemand(from, travelDir, e.rng)
	if !ok {
		return 0, false
	}
	e.res.ZoneDeliveries[z]++
	return z, true
}

func (e *engine) tryPlantDemand() bool {
	if e.board.TryConsumeZone(board.ZonePlant, e.rng) {
		e.res.ZoneDeliveries[board.ZonePlant]++
		return true
	}
	return false
}

func (e *engine) deflectEmitter(chip Chip) []Chip {
	return []Chip{{
		Type: chip.Type,
		Pos:  hex.Coord{Q: hex.EmitterCol, R: hex.EmitterRow},
		Dir:  hex.RandomShootDir(e.rng),
	}}
}

func shootDir(cfg Config, rng *rand.Rand) int {
	if cfg.StartDir >= 0 {
		return cfg.StartDir % 6
	}
	return hex.RandomShootDir(rng)
}

// EmitterChips returns chips fired from the igniter at shift start.
func EmitterChips(cfg Config, rng *rand.Rand) []Chip {
	pos := hex.Coord{Q: hex.EmitterCol, R: hex.EmitterRow}
	if cfg.MixedEmitterTrigger {
		ct := ChipHeat
		if rng.Intn(2) == 1 {
			ct = ChipNeutron
		}
		return []Chip{{Type: ct, Pos: pos, Dir: shootDir(cfg, rng)}}
	}
	chips := make([]Chip, 0, cfg.InitialHeat+cfg.InitialNeutron)
	for i := 0; i < cfg.InitialHeat; i++ {
		chips = append(chips, Chip{Type: ChipHeat, Pos: pos, Dir: shootDir(cfg, rng)})
	}
	for i := 0; i < cfg.InitialNeutron; i++ {
		chips = append(chips, Chip{Type: ChipNeutron, Pos: pos, Dir: shootDir(cfg, rng)})
	}
	return chips
}

func (e *engine) purgeChipsAt(c hex.Coord) {
	n := 0
	for _, chip := range e.queue {
		if chip.Pos != c {
			e.queue[n] = chip
			n++
		}
	}
	e.queue = e.queue[:n]
}

func (e *engine) maybeVoluntaryFire() bool {
	if e.res.CriticalFailure {
		return false
	}
	type source struct {
		pos  hex.Coord
		kind field.Type
	}
	sources := make([]source, 0)
	for _, c := range hex.AllBoardCoords {
		t := &e.board.Tiles[c.Q][c.R]
		switch t.Type {
		case field.CapacitorBank, field.PumpedStorage, field.LeadAccumulator:
			if t.StoredVoltage > 0 {
				sources = append(sources, source{c, t.Type})
			}
		case field.EmergencyGenerator:
			if !t.BurnedOut && t.Charge > 0 {
				sources = append(sources, source{c, t.Type})
			}
		}
	}
	if len(sources) == 0 {
		return false
	}
	if len(e.queue) > 0 && e.rng.Float64() > 0.35 {
		return false
	}
	src := sources[e.rng.Intn(len(sources))]
	tile := &e.board.Tiles[src.pos.Q][src.pos.R]
	switch src.kind {
	case field.EmergencyGenerator:
		tile.Charge--
	default:
		tile.StoredVoltage--
	}
	e.queue = append(e.queue, Chip{
		Type: ChipVoltage,
		Pos:  src.pos,
		Dir:  hex.RandomTravelDir(e.rng),
	})
	e.lastResolved = e.queue[len(e.queue)-1]
	e.hasResolved = true
	e.recordStep("Freiwilliger Schuss", nil)
	return true
}

func looseChipsOnSide(b *board.State, chips []Chip, player1 bool) int {
	count := 0
	for _, chip := range chips {
		if chip.Pos.IsPlayer1() == player1 {
			count++
		}
	}
	if !player1 {
		for _, c := range hex.AllBoardCoords {
			if !c.IsPlayer2() {
				continue
			}
			count += b.Tiles[c.Q][c.R].StoredVoltage
		}
	}
	return count
}

func criticalExceeded(b *board.State, queue []Chip, inFlight []Chip, limit int) bool {
	chips := append(append([]Chip{}, queue...), inFlight...)
	return looseChipsOnSide(b, chips, true) > limit ||
		looseChipsOnSide(b, chips, false) > limit
}

func (e *engine) processTurbine(chip Chip) (event string, active *Chip) {
	tCoord := hex.Coord{Q: hex.TurbineCol, R: hex.TurbineRow}
	switch chip.Type {
	case ChipHeat:
		e.res.HeatAtTurbine++
		e.queue = append(e.queue, Chip{
			Type: ChipVoltage,
			Pos:  tCoord,
			Dir:  hex.RandomShootDir(e.rng),
		})
		return "Turbine", &chip
	case ChipVoltage:
		if e.tryPlantDemand() {
			return board.BorderDemandEvent(board.ZonePlant), &chip
		}
		spiked := Chip{
			Type: ChipVoltage,
			Pos:  tCoord,
			Dir:  hex.RandomShootDir(e.rng),
		}
		e.queue = append(e.queue, spiked)
		e.lastResolved = spiked
		e.hasResolved = true
		return "Spannungs-Spike", nil
	default:
		return "Turbine", &chip
	}
}

func passThrough(chip Chip, pos hex.Coord) []Chip {
	return []Chip{{Type: chip.Type, Pos: pos, Dir: chip.Dir}}
}

func react(b *board.State, g *graph.Graph, pos hex.Coord, tile *field.Tile, chip Chip, incoming int, rng *rand.Rand) ([]Chip, bool) {
	graphChanged := false

	switch tile.Type {
	case field.Empty:
		return passThrough(chip, pos), false

	case field.Mirror:
		if chip.Type == ChipVoltage {
			return passThrough(chip, pos), false
		}
		out := tile.Orientation.WireOutgoing(incoming)
		return []Chip{{Type: chip.Type, Pos: pos, Dir: out}}, false

	case field.Relay:
		if chip.Type != ChipVoltage {
			return passThrough(chip, pos), false
		}
		out := tile.Orientation.WireOutgoing(incoming)
		return []Chip{{Type: chip.Type, Pos: pos, Dir: out}}, false

	case field.CoolingTower:
		if chip.Type == ChipHeat {
			return nil, false
		}
		return passThrough(chip, pos), false

	case field.AbsorberRod:
		if chip.Type == ChipNeutron {
			return nil, false
		}
		return passThrough(chip, pos), false

	case field.CoalChamber:
		return handleFuel(tile, chip, pos, rng, 1, 2, ChipHeat, func(t *field.Tile) {
			t.Charge--
			if t.Charge <= 0 {
				t.BurnedOut = true
			}
		})

	case field.GasBoiler:
		return handleFuel(tile, chip, pos, rng, 3, 4, ChipHeat, func(t *field.Tile) {
			t.Charge -= 3
			if t.Charge <= 0 {
				t.BurnedOut = true
			}
		})

	case field.UraniumPlate:
		if tile.BurnedOut {
			return nil, false
		}
		if chip.Type == ChipHeat {
			return emitRandom(pos, rng, ChipHeat, 1), false
		}
		if chip.Type != ChipNeutron || tile.Charge <= 0 {
			return passThrough(chip, pos), false
		}
		tile.Charge--
		if tile.Charge <= 0 {
			tile.BurnedOut = true
			graphChanged = true
		}
		chips := emitRandom(pos, rng, ChipNeutron, 2)
		chips = append(chips, emitRandom(pos, rng, ChipHeat, 1)...)
		return chips, graphChanged

	case field.Tokamak:
		if chip.Type != ChipNeutron {
			return passThrough(chip, pos), false
		}
		tile.TokamakCounter++
		if tile.TokamakCounter >= 4 {
			tile.TokamakCounter = 0
			return emitRandom(pos, rng, ChipHeat, 8), false
		}
		return nil, false

	case field.Transformer:
		return handleFuel(tile, chip, pos, rng, 1, 2, ChipVoltage, func(t *field.Tile) {
			t.Charge--
			if t.Charge <= 0 {
				t.BurnedOut = true
			}
		})

	case field.Ground:
		if chip.Type != ChipVoltage || tile.Charge <= 0 {
			if tile.BurnedOut {
				return nil, false
			}
			return passThrough(chip, pos), false
		}
		tile.Charge--
		if tile.Charge <= 0 {
			tile.BurnedOut = true
			graphChanged = true
		}
		return nil, false

	case field.HVCascade:
		return handleFuel(tile, chip, pos, rng, 3, 4, ChipVoltage, func(t *field.Tile) {
			t.Charge -= 3
			if t.Charge <= 0 {
				t.BurnedOut = true
			}
		})

	case field.CapacitorBank, field.PumpedStorage, field.LeadAccumulator:
		if chip.Type != ChipVoltage {
			return passThrough(chip, pos), false
		}
		max := field.Catalog[tile.Type].MaxCharge
		if tile.StoredVoltage < max {
			tile.StoredVoltage++
			return nil, false
		}
		return emitRandom(pos, rng, ChipVoltage, 1), false

	case field.Superconductor:
		if chip.Type != ChipVoltage {
			return passThrough(chip, pos), false
		}
		target := pos
		for step := 0; step < hex.Cols+hex.Rows; step++ {
			n := target.Neighbor(tile.SuperTarget.TravelDir())
			if !hex.CanEnter(target, n) || !n.IsPlayer2() {
				break
			}
			target = n
		}
		return []Chip{{Type: ChipVoltage, Pos: target, Dir: tile.SuperTarget.TravelDir()}}, false

	case field.EmergencyGenerator:
		if chip.Type == ChipVoltage {
			*tile = field.Tile{Type: field.Empty}
			return nil, true
		}
		return passThrough(chip, pos), false
	}

	return passThrough(chip, pos), false
}

func handleFuel(tile *field.Tile, chip Chip, pos hex.Coord, rng *rand.Rand, cost, emit int, required ChipType, consume func(*field.Tile)) ([]Chip, bool) {
	changed := false
	if tile.BurnedOut {
		return nil, false
	}
	if chip.Type != required {
		return passThrough(chip, pos), false
	}
	if tile.Charge < cost {
		return passThrough(chip, pos), false
	}
	consume(tile)
	if tile.BurnedOut {
		changed = true
	}
	return emitRandom(pos, rng, required, emit), changed
}

func emitRandom(pos hex.Coord, rng *rand.Rand, ct ChipType, n int) []Chip {
	out := make([]Chip, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, Chip{Type: ct, Pos: pos, Dir: hex.RandomTravelDir(rng)})
	}
	return out
}

// RunMonteCarlo runs many simulations and returns individual results.
func RunMonteCarlo(state *board.State, runs int, rng *rand.Rand, cfg Config) []Result {
	results := make([]Result, runs)
	for i := 0; i < runs; i++ {
		results[i] = Run(state, rng, cfg)
	}
	return results
}

func (cfg Config) demandsForRun(rng *rand.Rand) board.ShiftDemands {
	if cfg.EnergyCard.ID == "" {
		return cfg.ShiftDemands
	}
	shift := cfg.Shift
	if cfg.RandomShift && rng != nil {
		shift = 1 + rng.Intn(5)
	}
	if shift < 1 {
		shift = 1
	}
	if shift > 5 {
		shift = 5
	}
	return cfg.EnergyCard.ShiftDemands(shift)
}
