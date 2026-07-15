// Package graph builds a probabilistic flow graph from a board state.
package graph

import (
	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
)

// ParticleType is a flowing chip category.
type ParticleType int

const (
	Heat ParticleType = iota
	Neutron
	Voltage
)

// Edge holds transition probabilities for the three particle types on one directed edge.
type Edge struct {
	To      hex.Coord
	Heat    float64
	Neutron float64
	Voltage float64
}

// Node is one hex cell with outgoing edges.
type Node struct {
	Coord hex.Coord
	Edges []Edge
}

// Graph is the full board graph derived from a game state.
type Graph struct {
	Nodes map[hex.Coord]Node
}

func addEmitterEdges(node *Node) {
	addShootEdges(node, 1, 0, 0)
}

func addTurbineEdges(node *Node) {
	addShootEdges(node, 0, 0, 1)
}

func addShootEdges(node *Node, heat, neutron, voltage float64) {
	c := node.Coord
	p := 1.0 / float64(len(hex.ShootRotations))
	for _, rot := range hex.ShootRotations {
		next := c.StepTarget(rot.TravelDir())
		if hex.CanEnter(c, next) {
			node.Edges = append(node.Edges, Edge{
				To:      next,
				Heat:    heat * p,
				Neutron: neutron * p,
				Voltage: voltage * p,
			})
		}
	}
}

type rawOut struct {
	to      hex.Coord
	heat    float64
	neutron float64
	voltage float64
}

func mergeEdge(node *Node, o rawOut) {
	for i := range node.Edges {
		if node.Edges[i].To == o.to {
			node.Edges[i].Heat += o.heat
			node.Edges[i].Neutron += o.neutron
			node.Edges[i].Voltage += o.voltage
			return
		}
	}
	node.Edges = append(node.Edges, Edge{
		To: o.to, Heat: o.heat, Neutron: o.neutron, Voltage: o.voltage,
	})
}

func normalizeNodeEdges(g *Graph) {
	for c, node := range g.Nodes {
		var hSum, nSum, vSum float64
		for _, e := range node.Edges {
			hSum += e.Heat
			nSum += e.Neutron
			vSum += e.Voltage
		}
		for i, e := range node.Edges {
			if hSum > 0 {
				node.Edges[i].Heat = e.Heat / hSum
			}
			if nSum > 0 {
				node.Edges[i].Neutron = e.Neutron / nSum
			}
			if vSum > 0 {
				node.Edges[i].Voltage = e.Voltage / vSum
			}
		}
		g.Nodes[c] = node
	}
}

func outgoingTransitions(c hex.Coord, tile *field.Tile, incomingDir int) []rawOut {
	if tile.BurnedOut {
		switch tile.Type {
		case field.CoalChamber:
			return emitterOut(c, 1, 1, 0, true)
		case field.Transformer:
			return burnedRedirectOut(c)
		}
		return nil
	}

	if c.IsTurbine() {
		return nil
	}

	switch tile.Type {
	case field.Empty:
		return forwardOrReflect(c, incomingDir, 1, 1, 1)
	case field.Mirror:
		outDir := tile.Orientation.WireOutgoing(incomingDir)
		next := c.Neighbor(outDir)
		if hex.CanEnter(c, next) {
			return []rawOut{{to: next, heat: 1, neutron: 1}}
		}
	case field.Relay:
		outDir := tile.Orientation.WireOutgoing(incomingDir)
		next := c.Neighbor(outDir)
		if hex.CanEnter(c, next) {
			return []rawOut{{to: next, voltage: 1}}
		}
	case field.CoalChamber:
		return emitterOut(c, 2, 1, 0, tile.Charge > 0)
	case field.GasBoiler:
		return emitterOut(c, 4, 1, 0, tile.Charge > 0)
	case field.CoolingTower:
		if tile.Orientation.ParallelToAxis(incomingDir) {
			outDir := hex.PassThroughDir(incomingDir)
			next := c.Neighbor(outDir)
			if hex.CanEnter(c, next) {
				return []rawOut{{to: next, heat: 1}}
			}
		}
		return absorb()
	case field.AbsorberRod:
		return absorbNeutron()
	case field.UraniumPlate:
		if tile.Charge <= 0 {
			return nil
		}
		outs := emitterOut(c, 1, 0, 1, true) // 1 heat
		neutronOuts := emitterOut(c, 2, 0, 0, true)
		for i := range outs {
			outs[i].neutron = neutronOuts[i].neutron
		}
		return outs
	case field.Tokamak:
		return nil // handled dynamically in simulation
	case field.Transformer:
		return emitterOut(c, 0, 0, 2, tile.Charge > 0)
	case field.Ground:
		if tile.Orientation.ParallelToAxis(incomingDir) {
			outDir := hex.PassThroughDir(incomingDir)
			next := c.Neighbor(outDir)
			if hex.CanEnter(c, next) {
				return []rawOut{{to: next, voltage: 1}}
			}
		}
		return absorbVoltage()
	case field.HVCascade:
		return emitterOut(c, 0, 0, 4, tile.Charge > 0)
	case field.Superconductor:
		// Teleport to border in superTarget direction — approximate as neighbor chain.
		next := c
		for step := 0; step < hex.Cols+hex.Rows; step++ {
			n := next.Neighbor(tile.SuperTarget.TravelDir())
			if !hex.CanEnter(next, n) || !n.IsPlayer2() {
				break
			}
			next = n
		}
		if next != c {
			return []rawOut{{to: next, voltage: 1}}
		}
	}
	return nil
}

func emitterOut(c hex.Coord, count int, heatW, neutronW float64, active bool) []rawOut {
	if !active {
		return nil
	}
	pt := Heat
	if neutronW > 0 && heatW == 0 {
		pt = Neutron
	}
	outs := make([]rawOut, 0, 6)
	p := 1.0 / 6.0
	for dir := 0; dir < 6; dir++ {
		to, ok := flowTarget(c, dir, pt)
		if !ok {
			continue
		}
		outs = append(outs, rawOut{
			to:      to,
			heat:    float64(count) * heatW * p,
			neutron: float64(count) * neutronW * p,
		})
	}
	return outs
}

func burnedRedirectOut(c hex.Coord) []rawOut {
	outs := make([]rawOut, 0, 18)
	p := 1.0 / 6.0
	for dir := 0; dir < 6; dir++ {
		for _, pt := range []ParticleType{Heat, Neutron, Voltage} {
			to, ok := flowTarget(c, dir, pt)
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
			outs = append(outs, ro)
		}
	}
	return outs
}

func absorb() []rawOut { return nil }

func absorbNeutron() []rawOut { return nil }

func absorbVoltage() []rawOut { return nil }

func forwardOrReflect(c hex.Coord, incomingDir int, heat, neutron, voltage float64) []rawOut {
	next := c.Neighbor(incomingDir)
	if hex.CanEnter(c, next) {
		return []rawOut{{to: next, heat: heat, neutron: neutron, voltage: voltage}}
	}
	switch hex.BlockedBoundary(c, next, incomingDir) {
	case hex.BoundaryInternalWall:
		if heat > 0 {
			turbine := hex.Coord{Q: hex.TurbineCol, R: hex.TurbineRow}
			return []rawOut{{to: turbine, voltage: heat}}
		}
		return nil
	case hex.BoundaryOuter:
		h, n, v := 0.0, 0.0, 0.0
		if heat > 0 && c.IsPlayer1() {
			h = heat
		}
		if voltage > 0 && c.IsPlayer2() {
			return nil
		}
		if neutron > 0 {
			return nil
		}
		if h == 0 {
			return nil
		}
		refDir := hex.ReflectOffOuterWall(incomingDir)
		refNext := c.Neighbor(refDir)
		if hex.CanEnter(c, refNext) {
			return []rawOut{{to: refNext, heat: h, neutron: n, voltage: v}}
		}
		return nil
	default:
		return nil
	}
}

// Outgoing returns normalized edges for a coordinate from the graph.
func (g *Graph) Outgoing(c hex.Coord) []Edge {
	if n, ok := g.Nodes[c]; ok {
		return n.Edges
	}
	return nil
}

// Rebuild updates the graph after board state changes (e.g. burnout).
func Rebuild(g *Graph, state *board.State, chips []InFlight) {
	*g = *BuildFlow(state, chips)
}
