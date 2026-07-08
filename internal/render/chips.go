package render

import (
	"fmt"
	"strings"

	"github.com/jonas/reaktor-sim/internal/field"
	"github.com/jonas/reaktor-sim/internal/hex"
	"github.com/jonas/reaktor-sim/internal/sim"
)

// ChipView holds loose chips for cell annotations.
type ChipView struct {
	Queue  []sim.Chip
	Active *sim.Chip
}

// LooseCounts tallies unbound chips on and incoming to a cell.
type LooseCounts struct {
	OnHeat, OnNeutron, OnVoltage int
	InHeat, InNeutron, InVoltage int
}

// LooseCountsAt counts queue chips on c and heading into c.
func LooseCountsAt(view ChipView, c hex.Coord) LooseCounts {
	var r LooseCounts
	for _, chip := range view.Queue {
		if chip.Pos == c {
			addOn(&r, chip.Type)
		}
		if chip.Pos.Neighbor(chip.Dir) == c {
			addIn(&r, chip.Type)
		}
	}
	if view.Active != nil && !chipInQueue(view.Queue, *view.Active) {
		if view.Active.Pos.Neighbor(view.Active.Dir) == c {
			addIn(&r, view.Active.Type)
		}
	}
	return r
}

func chipInQueue(queue []sim.Chip, chip sim.Chip) bool {
	for _, q := range queue {
		if q.Type == chip.Type && q.Pos == chip.Pos && q.Dir == chip.Dir {
			return true
		}
	}
	return false
}

// LooseLabel formats unbound (+) and incoming (>) chips, e.g. "+2W>1N".
func LooseLabel(counts LooseCounts) string {
	on := formatLoose(counts.OnHeat, counts.OnNeutron, counts.OnVoltage, "+")
	in := formatLoose(counts.InHeat, counts.InNeutron, counts.InVoltage, ">")
	if on == "" {
		return in
	}
	if in == "" {
		return on
	}
	return on + in
}

func formatLoose(h, n, v int, prefix string) string {
	parts := make([]string, 0, 3)
	if h > 0 {
		parts = append(parts, fmt.Sprintf("%s%dW", prefix, h))
	}
	if n > 0 {
		parts = append(parts, fmt.Sprintf("%s%dN", prefix, n))
	}
	if v > 0 {
		parts = append(parts, fmt.Sprintf("%s%dS", prefix, v))
	}
	return strings.Join(parts, "")
}

func addOn(r *LooseCounts, t sim.ChipType) {
	switch t {
	case sim.ChipNeutron:
		r.OnNeutron++
	case sim.ChipVoltage:
		r.OnVoltage++
	default:
		r.OnHeat++
	}
}

func addIn(r *LooseCounts, t sim.ChipType) {
	switch t {
	case sim.ChipNeutron:
		r.InNeutron++
	case sim.ChipVoltage:
		r.InVoltage++
	default:
		r.InHeat++
	}
}

// isEmptySlot reports an unoccupied playable cell (no field placed).
func isEmptySlot(c hex.Coord, tile field.Tile) bool {
	if c.IsEmitter() || c.IsTurbine() || c.HasWallRight() {
		return false
	}
	if tile.BurnedOut {
		return false
	}
	return tile.Type == field.Empty
}

// CellHasContent reports whether a cell gets fill, symbol and dark border.
func CellHasContent(c hex.Coord, tile field.Tile, view ChipView) bool {
	if !c.Valid() {
		return false
	}
	if c.IsEmitter() || c.IsTurbine() || c.HasWallRight() {
		return true
	}
	if tile.BurnedOut || tile.Type != field.Empty {
		return true
	}
	return LooseLabel(LooseCountsAt(view, c)) != ""
}
