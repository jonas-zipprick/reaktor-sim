// Package sim runs Monte-Carlo simulations of one reactor shift.
package sim

import (
	"math"
	"math/rand"
	"sort"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/energy"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/finance"
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
const MaxStepsPerRun = 500

// Result holds metrics from one simulation run.
type Result struct {
	HeatAtTurbine     int
	ZoneDeliveries    [4]int
	CriticalFailure   bool
	CriticalP1        bool // critical mass on player 1 (reactor) side
	CriticalP2        bool // critical mass on player 2 (grid) side
	StepLimitExceeded bool
	Steps             int // chip resolutions executed in this run
	AllDemandsMet     bool
	EndDemands        [4]int // remaining demand per board.Zone at run end
	EndDamage         [4]int // remaining damage per board.Zone at run end
	EndEmitterDamage  int    // remaining igniter damage at run end
	Shift             int    // energy-card shift used for this run (1-5)
	ReactorRepairSpent int   // money spent on igniter repair at shift start
	RepairSpent       int    // money spent on grid damage repair at shift start
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
	StartDir      int // 0-5 emitter shoot override per run, or -1 = random once per run
	ShiftDemands  board.ShiftDemands
	EnergyCard    energy.Card // when set, overrides ShiftDemands per run
	FinanceCard   finance.Card
	Shift         int         // 1-5 for EnergyCard
	RandomShift          bool // pick shift 1-5 per Monte-Carlo run
	ReactorRepairBudget  int  // max money for igniter repair per run (1 per chip); 0 = none
	RepairBudget         int  // max money for grid damage repair per run (1 per chip); 0 = none
	Trace                bool
	TraceStep            func(Snapshot) error // if set, snapshots are streamed and not retained
	InitialChips  []Chip
}

func DefaultConfig() Config {
	card := energy.DefaultCard()
	fin := finance.DefaultCard()
	return Config{
		MaxSteps:      MaxStepsPerRun,
		CriticalLimit: fin.CriticalLimit(),
		EnergyCard:    card,
		FinanceCard:   fin,
		Shift:         1,
		ShiftDemands:  card.ShiftDemands(1),
	}
}

type engine struct {
	board           *board.State
	graph           *graph.Graph
	queue           []Chip
	res             Result
	rng             *rand.Rand
	cfg             Config
	emitterShootDir int // fixed for entire run (Zünder)
	turbineShootDir int // fixed for entire run (Turbine voltage)
	trace           []Snapshot
	step         int
	lastResolved Chip
	hasResolved  bool
	lostRecorded bool
	lastEmitted  []Chip // chips released by the latest fuel/cascade reaction
}

// Run executes one full shift simulation on a board clone.
func Run(state *board.State, rng *rand.Rand, cfg Config) Result {
	res, _, _ := run(state, rng, cfg)
	return res
}

// RunTrace executes one shift and records snapshots after each chip resolution.
func RunTrace(state *board.State, rng *rand.Rand, cfg Config) (Result, []Snapshot) {
	cfg.Trace = true
	res, _, snaps := run(state, rng, cfg)
	return res, snaps
}

// MedianRunIndex picks the run whose step count is closest to the median.
func MedianRunIndex(results []Result) int {
	if len(results) == 0 {
		return 0
	}
	steps := make([]int, len(results))
	for i, r := range results {
		steps[i] = r.Steps
	}
	med := medianInt(steps)
	best := 0
	bestDist := absInt(results[0].Steps - med)
	for i, r := range results {
		d := absInt(r.Steps - med)
		if d < bestDist || (d == bestDist && i < best) {
			best = i
			bestDist = d
		}
	}
	return best
}

// ApplyShiftCarry re-runs one Monte-Carlo run and copies its post-shift-end
// placeable tiles onto dst for multi-shift board continuity.
func ApplyShiftCarry(dst *board.State, seed int64, runOneBased int, cfg Config) {
	runRNG := rand.New(rand.NewSource(seed + int64(runOneBased)))
	_, endBoard, _ := run(dst, runRNG, cfg)
	board.CopyPlaceableTiles(dst, endBoard)
}

func medianInt(vals []int) int {
	if len(vals) == 0 {
		return 0
	}
	s := append([]int(nil), vals...)
	sort.Ints(s)
	n := len(s)
	if n%2 == 1 {
		return s[n/2]
	}
	return int(math.Round(float64(s[n/2-1]+s[n/2]) / 2))
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func run(state *board.State, rng *rand.Rand, cfg Config) (Result, *board.State, []Snapshot) {
	emitterDir, turbineDir := pickRunShootDirs(cfg, rng)
	e := &engine{
		board:           state.Clone(),
		rng:             rng,
		cfg:             cfg,
		emitterShootDir: emitterDir,
		turbineShootDir: turbineDir,
	}
	if cfg.ReactorRepairBudget > 0 {
		e.res.ReactorRepairSpent = e.board.RepairEmitterDamage(cfg.ReactorRepairBudget)
	}
	if cfg.RepairBudget > 0 {
		e.res.RepairSpent = e.board.RepairRandomDamage(e.rng, cfg.RepairBudget)
	}
	shift, demands := cfg.runShiftAndDemands(e.rng)
	e.res.Shift = shift
	e.board.Demands = make(map[hex.Coord][4]int)
	e.board.ApplyDemands(demands)
	e.queue = make([]Chip, 0, 64)
	if cfg.InitialChips != nil {
		e.queue = append(e.queue, cfg.InitialChips...)
	} else if hasOutstandingDemand(e.board) {
		e.queue = append(e.queue, emitterChips(e.board, cfg, e.rng, e.emitterShootDir)...)
	}
	e.queue = append(e.queue, shiftStartChips(e.board, e.rng)...)
	e.graph = graph.BuildFlow(e.board, e.inFlight())

	e.record("Start")
	if e.checkCritical() {
		e.captureEndState()
		return e.res, e.board, finalizeTrace(e.trace)
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
		shiftEndCleanup(e.board)
		e.record("Ende")
	}
	e.res.AllDemandsMet = AllDemandsMet(e.board)
	e.captureEndState()

	return e.res, e.board, finalizeTrace(e.trace)
}

func (e *engine) captureEndState() {
	e.res.Steps = e.step
	for z := board.ZoneIndustry; z <= board.ZonePlant; z++ {
		e.res.EndDemands[z] = e.board.TotalDemand(z)
		e.res.EndDamage[z] = e.board.TotalDamage(z)
	}
	e.res.EndEmitterDamage = e.board.EmitterDamage
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
		narrative = narrate(event, active, e.board, queue, nil)
	} else if active != nil {
		narrative = narrate(event, active, e.board, queue, e.lastEmitted)
	} else if e.hasResolved {
		c := e.lastResolved
		narrative = narrate(event, &c, e.board, queue, e.lastEmitted)
	} else {
		narrative = narrate(event, nil, e.board, queue, nil)
	}
	e.emitTraceSnapshot(event, narrative, queue, activeCopy)
}

func (e *engine) emitTraceSnapshot(event, narrative string, queue []Chip, activeCopy *Chip) {
	snap := Snapshot{
		Step:      e.step,
		Event:     event,
		Narrative: narrative,
		Board:     e.board.Clone(),
		Graph:     graph.BuildFlow(e.board, e.inFlight()),
		Queue:     queue,
		Active:    activeCopy,
		QueueSize: len(e.queue),
	}
	if e.cfg.TraceStep != nil {
		_ = e.cfg.TraceStep(snap)
		return
	}
	e.trace = append(e.trace, snap)
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

func (e *engine) loseGame(p1, p2 bool) {
	if e.lostRecorded {
		e.res.CriticalFailure = true
		e.queue = nil
		return
	}
	e.res.CriticalFailure = true
	e.res.CriticalP1 = p1
	e.res.CriticalP2 = p2
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
	p1, p2 := criticalSidesExceeded(e.board, e.queue, inFlight, e.cfg.CriticalLimit)
	if !p1 && !p2 {
		return false
	}
	e.loseGame(p1, p2)
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

	if chip.Type == ChipVoltage && board.TurbinePlantWallLeave(chip.Pos, chip.Dir) {
		if e.handlePlantWallLeave(chip) {
			return
		}
	}

	nextPos := chip.Pos.StepTarget(chip.Dir)

	if nextPos.IsEmitter() {
		e.board.AddEmitterDamage()
		e.recordStep("Zuender-Treffer", &chip)
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
	if tile.BurnedOut && tryBurnedRedirect(e, nextPos, tile, chip) {
		return
	}
	if tile.BurnedOut {
		e.recordStep("Ausgebrannt", &chip)
		return
	}

	chargeBefore := tile.Charge
	incoming := (chip.Dir + 3) % 6
	tileTypeBefore := tile.Type

	newChips, graphChanged := react(e.board, e.graph, nextPos, tile, chip, incoming, e.cfg.EnergyCard.ID, e.rng)
	if chargeBefore > tile.Charge && isFuelEmitter(tile.Type) && chip.Type == chipTypeForFuel(tile.Type) {
		e.lastEmitted = newChips
	} else {
		e.lastEmitted = nil
	}
	e.queue = append(e.queue, newChips...)
	if graphChanged {
		graph.Rebuild(e.graph, e.board, e.inFlight())
	}
	event := "Feldreaktion"
	if tileTypeBefore == field.CapacitorBank && tile.Type == field.Empty && len(newChips) > 0 {
		event = "Kondensator explodiert"
	}
	e.recordStep(event, &chip)
}

func tryBurnedRedirect(e *engine, pos hex.Coord, tile *field.Tile, chip Chip) bool {
	if !tile.BurnedOut {
		return false
	}
	switch tile.Type {
	case field.CoalChamber:
		if chip.Type != ChipHeat {
			return false
		}
	case field.Transformer:
		// any chip type
	case field.EmergencyGenerator:
		if chip.Type != ChipVoltage {
			return false
		}
	default:
		return false
	}
	out := emitRandom(pos, e.rng, chip.Type, 1)
	e.queue = append(e.queue, out...)
	e.recordStep("Weiterleitung", &chip)
	return true
}

// shiftStartChips releases one stored voltage from each Blei-Akkumulator at shift start.
func shiftStartChips(b *board.State, rng *rand.Rand) []Chip {
	var out []Chip
	for _, c := range hex.AllBoardCoords {
		t := &b.Tiles[c.Q][c.R]
		if t.Type == field.LeadAccumulator && t.StoredVoltage > 0 {
			t.StoredVoltage--
			out = append(out, Chip{
				Type: ChipVoltage,
				Pos:  c,
				Dir:  hex.RandomTravelDir(rng),
			})
		}
	}
	return out
}

// shiftEndCleanup applies end-of-shift board cleanup per game rules: capacitor
// storage is cleared and depleted placeable fields are removed from the board.
func shiftEndCleanup(b *board.State) {
	for _, c := range hex.AllBoardCoords {
		t := &b.Tiles[c.Q][c.R]
		if t.Type == field.CapacitorBank {
			t.StoredVoltage = 0
		}
	}
	for _, c := range board.PlaceableSlots() {
		t := &b.Tiles[c.Q][c.R]
		if t.BurnedOut {
			*t = field.Tile{Type: field.Empty}
		}
	}
}

// reflectOffWall bounces a heat chip off the outer wall. If the chip sits on a
// mirror, the bounce re-enters the mirror from the wall edge and is deflected
// again (chips fly in straight lines, walls reflect at the same angle), so a
// chip a mirror sends into an adjacent wall is routed back the way it came.
func (e *engine) reflectOffWall(pos hex.Coord, dir int) int {
	reflected := hex.ReflectOffOuterWall(dir)
	tile := &e.board.Tiles[pos.Q][pos.R]
	if tile.Type == field.Mirror && !tile.BurnedOut {
		incoming := (reflected + 3) % 6
		return tile.Orientation.WireOutgoing(incoming)
	}
	return reflected
}

func (e *engine) handleBlocked(chip Chip, kind hex.BoundaryKind) {
	switch kind {
	case hex.BoundaryInternalWall:
		if chip.Type == ChipHeat {
			event, active := e.processTurbine(chip)
			e.recordStep(event, active)
			return
		}
		if chip.Type == ChipVoltage {
			if _, ok := board.PlantWallZone(chip.Pos, chip.Dir); ok {
				if e.handlePlantWallLeave(chip) {
					return
				}
			}
			reflected := Chip{
				Type: ChipVoltage,
				Pos:  chip.Pos,
				Dir:  e.reflectOffWall(chip.Pos, chip.Dir),
			}
			e.queue = append(e.queue, reflected)
			e.recordStep("Spannungs-Reflektion", &reflected)
			return
		}
		e.recordStep("Innere Wand", &chip)
		return
	case hex.BoundaryOuter:
		e.handleOuterBoundary(chip)
	}
}

func (e *engine) handleOuterBoundary(chip Chip) {
	switch chip.Type {
	case ChipHeat:
		if hex.HeatReflectsAtOuterWall(chip.Pos, chip.Dir) {
			reflected := Chip{
				Type: ChipHeat,
				Pos:  chip.Pos,
				Dir:  e.reflectOffWall(chip.Pos, chip.Dir),
			}
			e.queue = append(e.queue, reflected)
			e.recordStep("Waerme-Reflektion", &reflected)
		} else {
			e.recordStep("Waerme verpufft", &chip)
		}
	case ChipNeutron:
		e.recordStep("Neutron verpufft", &chip)
	case ChipVoltage:
		if hex.VoltageReflectsAtOuterWall(chip.Pos, chip.Dir) {
			reflected := Chip{
				Type: ChipVoltage,
				Pos:  chip.Pos,
				Dir:  hex.ReflectOffOuterWall(chip.Dir),
			}
			e.queue = append(e.queue, reflected)
			e.recordStep("Spannungs-Reflektion", &reflected)
			return
		}
		if z, ok := e.tryVoltageWallDelivery(chip.Pos, chip.Dir); ok {
			e.recordStep(board.BorderDemandEvent(z), &chip)
			return
		}
		if z, ok := e.board.AddWallDamage(chip.Pos, chip.Dir, e.rng); ok {
			e.recordStep(board.BorderDamageEvent(z), &chip)
			return
		}
		if !chip.Pos.IsPlayer2() {
			e.recordStep("Spannung verpufft", &chip)
			return
		}
		e.recordStep("Spannung verpufft", &chip)
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

func (e *engine) handlePlantWallLeave(chip Chip) bool {
	if z, ok := e.tryVoltageWallDelivery(chip.Pos, chip.Dir); ok {
		e.recordStep(board.BorderDemandEvent(z), &chip)
		return true
	}
	if z, ok := e.board.AddWallDamage(chip.Pos, chip.Dir, e.rng); ok {
		e.recordStep(board.BorderDamageEvent(z), &chip)
		return true
	}
	e.recordStep("Spannung verpufft", &chip)
	return true
}

func (e *engine) tryPlantDemand() bool {
	if e.board.TryConsumeZone(board.ZonePlant, e.rng) {
		e.res.ZoneDeliveries[board.ZonePlant]++
		return true
	}
	return false
}

func shootDir(cfg Config, rng *rand.Rand) int {
	if cfg.StartDir >= 0 {
		return cfg.StartDir % 6
	}
	return hex.RandomShootDir(rng)
}

func pickRunShootDirs(cfg Config, rng *rand.Rand) (emitter, turbine int) {
	return shootDir(cfg, rng), hex.RandomShootDir(rng)
}

// EmitterChips returns the chip fired from the igniter at shift start when demand
// remains on the board. All chips share one shoot direction (picked once per call).
func EmitterChips(state *board.State, cfg Config, rng *rand.Rand) []Chip {
	if !hasOutstandingDemand(state) {
		return nil
	}
	return emitterChips(state, cfg, rng, shootDir(cfg, rng))
}

func emitterChips(state *board.State, cfg Config, rng *rand.Rand, dir int) []Chip {
	pos := hex.Coord{Q: hex.EmitterCol, R: hex.EmitterRow}
	count := 1
	if cfg.EnergyCard.ID == "schturmowschtschina" {
		count = 2
	}
	chips := make([]Chip, 0, count)
	for i := 0; i < count; i++ {
		ct := ChipHeat
		if count == 1 && hasUraniumField(state) && rng.Intn(2) == 1 {
			ct = ChipNeutron
		}
		chips = append(chips, Chip{Type: ct, Pos: pos, Dir: dir})
	}
	return chips
}

func hasUraniumField(state *board.State) bool {
	for _, c := range hex.AllBoardCoords {
		if state.Tiles[c.Q][c.R].Type == field.UraniumPlate {
			return true
		}
	}
	return false
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
			if !t.BurnedOut && t.Charge > 0 && emergencyGeneratorMayFire(e) {
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
		if tile.Charge <= 0 {
			tile.BurnedOut = true
		}
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

// emergencyGeneratorMayFire applies player-2 heuristics for voluntary Notgenerator shots.
func emergencyGeneratorMayFire(e *engine) bool {
	if !hasOutstandingDemand(e.board) {
		return false
	}
	if player1HeatLoose(e.queue) {
		return false
	}
	return !otherVoltageOnBoard(e)
}

func otherVoltageOnBoard(e *engine) bool {
	for _, c := range e.queue {
		if c.Type == ChipVoltage {
			return true
		}
	}
	for _, c := range hex.AllBoardCoords {
		t := &e.board.Tiles[c.Q][c.R]
		switch t.Type {
		case field.CapacitorBank, field.PumpedStorage, field.LeadAccumulator:
			if t.StoredVoltage > 0 {
				return true
			}
		}
	}
	return false
}

func hasOutstandingDemand(s *board.State) bool {
	return !AllDemandsMet(s)
}

func player1HeatLoose(chips []Chip) bool {
	for _, c := range chips {
		if c.Type != ChipHeat {
			continue
		}
		if c.Pos.IsPlayer1() || c.Pos.IsEmitter() {
			return true
		}
	}
	return false
}

func looseChipsOnSide(b *board.State, chips []Chip, player1 bool) int {
	count := 0
	for _, chip := range chips {
		if chip.Pos.IsPlayer1() == player1 {
			count++
		}
	}
	if player1 {
		count += b.EmitterDamage
	}
	if !player1 {
		// Schadens-Chips are geloest; StoredVoltage in Speichern is gebunden until fired.
		count += b.TotalPlayer2Damage()
	}
	return count
}

func criticalSidesExceeded(b *board.State, queue []Chip, inFlight []Chip, limit int) (p1, p2 bool) {
	chips := append(append([]Chip{}, queue...), inFlight...)
	return looseChipsOnSide(b, chips, true) > limit,
		looseChipsOnSide(b, chips, false) > limit
}

func criticalExceeded(b *board.State, queue []Chip, inFlight []Chip, limit int) bool {
	p1, p2 := criticalSidesExceeded(b, queue, inFlight, limit)
	return p1 || p2
}

func (e *engine) processTurbine(chip Chip) (event string, active *Chip) {
	tCoord := hex.Coord{Q: hex.TurbineCol, R: hex.TurbineRow}
	switch chip.Type {
	case ChipHeat:
		e.res.HeatAtTurbine++
		e.queue = append(e.queue, Chip{
			Type: ChipVoltage,
			Pos:  tCoord,
			Dir:  e.turbineShootDir,
		})
		return "Turbine", &chip
	case ChipVoltage:
		if e.tryPlantDemand() {
			return board.BorderDemandEvent(board.ZonePlant), &chip
		}
		e.board.AddZoneDamage(board.ZonePlant)
		return board.BorderDamageEvent(board.ZonePlant), &chip
	default:
		return "Turbine", &chip
	}
}

func passThrough(chip Chip, pos hex.Coord) []Chip {
	return []Chip{{Type: chip.Type, Pos: pos, Dir: chip.Dir}}
}

func react(b *board.State, g *graph.Graph, pos hex.Coord, tile *field.Tile, chip Chip, incoming int, energyCardID string, rng *rand.Rand) ([]Chip, bool) {
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
		if energyCardID == "testlauf-volllast" {
			return passThrough(chip, pos), false
		}
		if chip.Type == ChipHeat {
			if tile.Orientation.ParallelToAxis(incoming) {
				out := hex.PassThroughDir(incoming)
				return []Chip{{Type: chip.Type, Pos: pos, Dir: out}}, false
			}
			return nil, false
		}
		return passThrough(chip, pos), false

	case field.AbsorberRod:
		if energyCardID == "testlauf-volllast" {
			return passThrough(chip, pos), false
		}
		if chip.Type == ChipNeutron {
			return nil, false
		}
		return passThrough(chip, pos), false

	case field.CoalChamber:
		return handleFuel(tile, chip, pos, rng, 1, 2, ChipHeat)

	case field.GasBoiler:
		return handleFuel(tile, chip, pos, rng, 3, 4, ChipHeat)

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
		return handleFuel(tile, chip, pos, rng, 1, 2, ChipVoltage)

	case field.Ground:
		if chip.Type != ChipVoltage {
			return passThrough(chip, pos), false
		}
		if tile.Orientation.ParallelToAxis(incoming) {
			out := hex.PassThroughDir(incoming)
			return []Chip{{Type: chip.Type, Pos: pos, Dir: out}}, false
		}
		return nil, false

	case field.HVCascade:
		return handleFuel(tile, chip, pos, rng, 3, 4, ChipVoltage)

	case field.CapacitorBank:
		if chip.Type != ChipVoltage {
			return passThrough(chip, pos), false
		}
		max := field.Catalog[field.CapacitorBank].MaxCharge
		if tile.StoredVoltage < max {
			tile.StoredVoltage++
			return nil, false
		}
		count := tile.StoredVoltage + 1
		*tile = field.Tile{Type: field.Empty}
		return emitRandom(pos, rng, ChipVoltage, count), true

	case field.PumpedStorage, field.LeadAccumulator:
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
		if chip.Type != ChipVoltage {
			return passThrough(chip, pos), false
		}
		if !tile.BurnedOut && tile.Charge > 0 {
			return nil, false
		}
		return emitRandom(pos, rng, ChipVoltage, 1), false
	}

	return passThrough(chip, pos), false
}

func handleFuel(tile *field.Tile, chip Chip, pos hex.Coord, rng *rand.Rand, cost, emit int, required ChipType) ([]Chip, bool) {
	if tile.BurnedOut {
		return nil, false
	}
	if chip.Type != required {
		return passThrough(chip, pos), false
	}
	fromCharge := cost
	if tile.Charge < fromCharge {
		fromCharge = tile.Charge
	}
	if fromCharge == 0 {
		return passThrough(chip, pos), false
	}
	totalEmit := 1 + fromCharge
	if totalEmit > emit {
		totalEmit = emit
	}
	tile.Charge -= fromCharge
	changed := false
	if tile.Charge <= 0 {
		tile.BurnedOut = true
		changed = true
	}
	out := make([]Chip, 0, totalEmit)
	for i := 0; i < totalEmit; i++ {
		out = append(out, Chip{Type: required, Pos: pos, Dir: hex.RandomTravelDir(rng)})
	}
	return out, changed
}

func isFuelEmitter(t field.Type) bool {
	switch t {
	case field.CoalChamber, field.GasBoiler, field.Transformer, field.HVCascade:
		return true
	default:
		return false
	}
}

func chipTypeForFuel(t field.Type) ChipType {
	switch t {
	case field.Transformer, field.HVCascade:
		return ChipVoltage
	default:
		return ChipHeat
	}
}

func emitRandom(pos hex.Coord, rng *rand.Rand, ct ChipType, n int) []Chip {
	out := make([]Chip, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, Chip{Type: ct, Pos: pos, Dir: hex.RandomTravelDir(rng)})
	}
	return out
}

// RunMonteCarlo runs many simulations and returns individual results.
// Each run uses an independent RNG seeded with baseSeed+runIndex (1-based),
// matching -trace and -trace-loop reproducibility.
func RunMonteCarlo(state *board.State, runs int, baseSeed int64, cfg Config) []Result {
	results := make([]Result, runs)
	for i := 0; i < runs; i++ {
		runRNG := rand.New(rand.NewSource(baseSeed + int64(i+1)))
		results[i] = Run(state, runRNG, cfg)
	}
	return results
}

// AllDemandsMet reports whether no demand chips remain on the board.
func AllDemandsMet(s *board.State) bool {
	for _, z := range []board.Zone{
		board.ZoneIndustry,
		board.ZoneResidential,
		board.ZoneRail,
		board.ZonePlant,
	} {
		if s.TotalDemand(z) > 0 {
			return false
		}
	}
	return true
}

// LoopTraceRunIndices returns 1-based Monte-Carlo run numbers for results that
// hit the step limit, at most max traces.
func LoopTraceRunIndices(results []Result, max int) []int {
	return traceRunIndices(results, max, func(r Result) bool { return r.StepLimitExceeded })
}

// WinTraceRunIndices returns 1-based Monte-Carlo run numbers for results where
// all demands were fulfilled, at most max traces.
func WinTraceRunIndices(results []Result, max int) []int {
	return traceRunIndices(results, max, func(r Result) bool { return r.AllDemandsMet })
}

func traceRunIndices(results []Result, max int, match func(Result) bool) []int {
	if max <= 0 || len(results) == 0 {
		return nil
	}
	out := make([]int, 0, max)
	for i, res := range results {
		if !match(res) {
			continue
		}
		out = append(out, i+1)
		if len(out) >= max {
			break
		}
	}
	return out
}

func (cfg Config) runShiftAndDemands(rng *rand.Rand) (int, board.ShiftDemands) {
	if cfg.EnergyCard.ID == "" {
		shift := cfg.Shift
		if shift < 1 {
			shift = 1
		}
		if shift > 5 {
			shift = 5
		}
		return shift, cfg.ShiftDemands
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
	return shift, cfg.EnergyCard.ShiftDemands(shift)
}
