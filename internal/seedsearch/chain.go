package seedsearch

import (
	"fmt"

	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/energy"
)

// TraceChain walks from a final-shift outcome back to shift 1 via PrevBoardFingerprint
// and matching carried demand/damage, returning outcomes in ascending shift order.
func TraceChain(scan ScanResult, final Outcome, card energy.Card) ([]Outcome, error) {
	if final.Shift < 1 || final.Shift > len(scan.Shifts) {
		return nil, fmt.Errorf("outcome shift %d outside scan", final.Shift)
	}
	chain := []Outcome{final}
	current := final
	for k := current.Shift; k > 1; k-- {
		prevOutcomes, err := scan.Shifts[k-2].LoadOutcomes()
		if err != nil {
			return nil, err
		}
		parent, err := findParent(prevOutcomes, current, k, card)
		if err != nil {
			return nil, err
		}
		chain = append([]Outcome{parent}, chain...)
		current = parent
	}
	return chain, nil
}

func findParent(candidates []Outcome, child Outcome, childShift int, card energy.Card) (Outcome, error) {
	cardDemand := card.ShiftDemands(childShift)
	wantDemand := carryDemand(child.StartDemands, cardDemand)
	for _, o := range candidates {
		if o.BoardFingerprint != child.PrevBoardFingerprint {
			continue
		}
		if o.MedianEndDamage != child.StartDamage {
			continue
		}
		if o.MedianEndEmitterDamage != child.StartEmitterDamage {
			continue
		}
		if o.MedianEndDemand != wantDemand {
			continue
		}
		if o.EndLeftover != child.StartLeftover {
			continue
		}
		return o, nil
	}
	return Outcome{}, fmt.Errorf("kein Parent fuer Schicht %d (board %s)", childShift, child.PrevBoardFingerprint)
}

func carryDemand(start board.ShiftDemands, card board.ShiftDemands) [4]int {
	return [4]int{
		start.Industry - card.Industry,
		start.Residential - card.Residential,
		start.Rail - card.Rail,
		start.Plant - card.Plant,
	}
}

// ShiftDirName formats a top-sim subdirectory for one shift outcome.
func ShiftDirName(o Outcome) string {
	return fmt.Sprintf("Schicht %d (seed%d, %s)", o.Shift, o.Seed, o.BoardFingerprint)
}
