package seedsearch

import (
	"container/heap"
	"sort"
)

type topHeap struct {
	items []Outcome
	less  func(a, b Outcome) bool
}

func (h topHeap) Len() int { return len(h.items) }

func (h topHeap) Less(i, j int) bool {
	return !h.less(h.items[i], h.items[j])
}

func (h topHeap) Swap(i, j int) { h.items[i], h.items[j] = h.items[j], h.items[i] }

func (h *topHeap) Push(x any) {
	h.items = append(h.items, x.(Outcome))
}

func (h *topHeap) Pop() any {
	old := h.items[len(h.items)-1]
	h.items = h.items[:len(h.items)-1]
	return old
}

func topN(outcomes []Outcome, n int, less func(a, b Outcome) bool) []Outcome {
	if n <= 0 || len(outcomes) == 0 {
		return nil
	}
	if n >= len(outcomes) {
		cp := append([]Outcome(nil), outcomes...)
		sort.Slice(cp, func(i, j int) bool { return less(cp[i], cp[j]) })
		return cp
	}
	h := &topHeap{
		items: append([]Outcome(nil), outcomes[:n]...),
		less:  less,
	}
	heap.Init(h)
	for _, o := range outcomes[n:] {
		if less(o, h.items[0]) {
			h.items[0] = o
			heap.Fix(h, 0)
		}
	}
	result := append([]Outcome(nil), h.items...)
	sort.Slice(result, func(i, j int) bool { return less(result[i], result[j]) })
	return result
}
