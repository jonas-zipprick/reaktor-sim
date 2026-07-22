package seedsearch

import (
	"container/heap"
	"sort"
)

func selectParentsFromShiftResult(sr ShiftResult, keep int) ([]parentBoard, error) {
	if len(sr.Outcomes) > 0 {
		return selectParents(sr.Outcomes, keep), nil
	}
	if sr.spillPath == "" {
		return nil, nil
	}
	return selectParentsStream(sr.spillPath, keep)
}

func selectParentsStream(path string, keep int) ([]parentBoard, error) {
	sel := newParentSelector(keep)
	err := readOutcomeBatches(path, func(batch []Outcome) error {
		for _, o := range batch {
			sel.consider(o)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return sel.parents(), nil
}

type parentSelector struct {
	keep       int
	candidateN int
	loopHeap   *topHeap
	tableHeaps []*topHeap
}

func newParentSelector(keep int) *parentSelector {
	if keep < 1 {
		keep = 1
	}
	sel := &parentSelector{
		keep:       keep,
		candidateN: keep * 4,
		loopHeap:   &topHeap{less: lessLoops},
		tableHeaps: make([]*topHeap, len(tableLessFns)),
	}
	for i, less := range tableLessFns {
		sel.tableHeaps[i] = &topHeap{less: less}
	}
	return sel
}

var tableLessFns = []func(a, b Outcome) bool{
	lessWins,
	lessAllDemandsNoDamage,
	lessMax1DemandNoDamage,
	lessMax1DemandMax1Damage,
}

func (s *parentSelector) consider(o Outcome) {
	if o.Loops > 0 {
		tryHeapAdd(s.loopHeap, o, s.keep)
	}
	for _, h := range s.tableHeaps {
		tryHeapAdd(h, o, s.candidateN)
	}
}

func tryHeapAdd(h *topHeap, o Outcome, maxSize int) {
	if maxSize <= 0 {
		return
	}
	if len(h.items) < maxSize {
		heap.Push(h, o)
		return
	}
	if h.less(o, h.items[0]) {
		h.items[0] = o
		heap.Fix(h, 0)
	}
}

func (s *parentSelector) parents() []parentBoard {
	exclude := make(map[string]struct{})
	for _, o := range s.loopHeap.items {
		exclude[parentBoardKey(o)] = struct{}{}
	}
	seen := make(map[string]struct{})
	var parents []parentBoard
	for _, h := range s.tableHeaps {
		ranked := sortedHeapItems(h)
		picked := 0
		for _, o := range ranked {
			if picked >= s.keep {
				break
			}
			key := parentBoardKey(o)
			if _, ok := seen[key]; ok {
				continue
			}
			if _, skip := exclude[key]; skip {
				continue
			}
			seen[key] = struct{}{}
			parents = append(parents, parentBoard{
				carryFP:       o.CarryBoardFingerprint,
				prevFP:        o.BoardFingerprint,
				demand:        o.MedianEndDemand,
				damage:        o.MedianEndDamage,
				emitterDamage: o.MedianEndEmitterDamage,
				leftover:      o.EndLeftover,
			})
			picked++
		}
	}
	return parents
}

func sortedHeapItems(h *topHeap) []Outcome {
	result := append([]Outcome(nil), h.items...)
	sort.Slice(result, func(i, j int) bool { return h.less(result[i], result[j]) })
	return result
}
