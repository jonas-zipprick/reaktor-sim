package graph

import (
	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
)

// InFlight is a chip currently moving on the board.
type InFlight struct {
	Particle ParticleType
	Pos      hex.Coord
	Dir      int
}

// BuildFlow creates a graph from chips in flight and optional player-fired storage.
// Only active sources emit edges; passive fields stay at 0 % until hit.
func BuildFlow(state *board.State, chips []InFlight) *Graph {
	g := &Graph{Nodes: make(map[hex.Coord]Node)}

	for _, chip := range chips {
		addFlowChip(g, chip)
	}

	enrichEmissionFans(g, state, chips)
	ensureNodes(g, state)
	normalizeNodeEdges(g)
	return g
}

// Build is an alias for BuildPotential (static topology when every field is stimulated).
// Prefer BuildFlow for trace snapshots and shift-start views.
func Build(state *board.State) *Graph {
	return BuildPotential(state)
}

// BuildPotential averages reactive transitions over all incoming directions.
// Useful for offline analysis, not for per-step simulation views.
func BuildPotential(state *board.State) *Graph {
	g := &Graph{Nodes: make(map[hex.Coord]Node)}
	for _, c := range hex.AllBoardCoords {
		tile := state.Tiles[c.Q][c.R]
		hasDemand := state.DemandLabel(c) != ""
		switch {
		case c.IsEmitter(), c.IsTurbine():
		case tile.Type != field.Empty:
		case hasDemand:
		default:
			continue
		}
		node := Node{Coord: c}
		switch {
		case c.IsEmitter():
			addEmitterEdges(&node)
		case c.IsTurbine():
			addTurbineEdges(&node)
		default:
			for dir := 0; dir < 6; dir++ {
				outs := outgoingTransitions(c, &tile, dir)
				for _, o := range outs {
					mergeEdge(&node, o)
				}
			}
		}
		if len(node.Edges) > 0 || c.IsTurbine() || c.IsEmitter() {
			g.Nodes[c] = node
		}
	}
	normalizeNodeEdges(g)
	return g
}

func addFlowChip(g *Graph, chip InFlight) {
	from := chip.Pos
	to, ok := flowTarget(from, chip.Dir, chip.Particle)
	if !ok {
		return
	}
	node := g.Nodes[from]
	node.Coord = from
	heat, neutron, voltage := 0.0, 0.0, 0.0
	switch chip.Particle {
	case Heat:
		heat = 1
	case Neutron:
		neutron = 1
	case Voltage:
		voltage = 1
	}
	mergeEdge(&node, rawOut{to: to, heat: heat, neutron: neutron, voltage: voltage})
	g.Nodes[from] = node
}

func flowTarget(from hex.Coord, dir int, particle ParticleType) (hex.Coord, bool) {
	next := from.Neighbor(dir)
	if hex.CanEnter(from, next) {
		return next, true
	}
	switch hex.BlockedBoundary(from, next, dir) {
	case hex.BoundaryOuter:
		switch particle {
		case Heat:
			if !from.IsPlayer1() {
				return hex.Coord{}, false
			}
		case Neutron:
			return hex.Coord{}, false
		case Voltage:
			if !from.IsPlayer2() {
				return hex.Coord{}, false
			}
		default:
			return hex.Coord{}, false
		}
		refDir := hex.ReflectOffOuterWall(dir)
		refNext := from.Neighbor(refDir)
		if hex.CanEnter(from, refNext) {
			return refNext, true
		}
		return hex.Coord{}, false
	default:
		return hex.Coord{}, false
	}
}

func ensureNodes(g *Graph, state *board.State) {
	for _, c := range hex.AllBoardCoords {
		tile := state.Tiles[c.Q][c.R]
		hasDemand := state.DemandLabel(c) != ""
		switch {
		case c.IsEmitter(), c.IsTurbine():
		case tile.Type != field.Empty:
		case hasDemand:
		default:
			continue
		}
		if _, ok := g.Nodes[c]; !ok {
			g.Nodes[c] = Node{Coord: c}
		}
	}
}

// enrichEmissionFans replaces per-chip edges at reactive fields with all six
// Richtungswuerfel outcomes (je 1/6), including outer-wall reflections.
func enrichEmissionFans(g *Graph, state *board.State, chips []InFlight) {
	byPos := map[hex.Coord]bool{}
	for _, chip := range chips {
		byPos[chip.Pos] = true
	}
	for pos := range byPos {
		tile := &state.Tiles[pos.Q][pos.R]
		pt, ok := emissionFanParticle(tile)
		if !ok {
			continue
		}
		node := g.Nodes[pos]
		node.Coord = pos
		node.Edges = nil
		addDirectionDiceEdges(&node, pos, pt)
		g.Nodes[pos] = node
	}
}

func emissionFanParticle(tile *field.Tile) (ParticleType, bool) {
	if tile.BurnedOut {
		return 0, false
	}
	switch tile.Type {
	case field.CoalChamber, field.GasBoiler:
		return Heat, true
	case field.Transformer, field.HVCascade:
		return Voltage, true
	default:
		return 0, false
	}
}

func addDirectionDiceEdges(node *Node, from hex.Coord, pt ParticleType) {
	p := 1.0 / 6.0
	for dir := 0; dir < 6; dir++ {
		to, ok := flowTarget(from, dir, pt)
		if !ok {
			continue
		}
		ro := rawOut{to: to}
		switch pt {
		case Heat:
			ro.heat = p
		case Neutron:
			ro.neutron = p
		case Voltage:
			ro.voltage = p
		}
		mergeEdge(node, ro)
	}
}
