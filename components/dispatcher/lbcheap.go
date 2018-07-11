package main

import (
	"container/heap"

	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/proto"
)

type lbcheapentry struct {
	gameid         uint16
	heapidx        int // index of this entry in the heap
	CPUPercent     float64
	origCPUPercent float64
}

func (e *lbcheapentry) update(info proto.GameLBCInfo) {
	e.origCPUPercent = info.CPUPercent
	e.CPUPercent = info.CPUPercent
}

type lbcheap []*lbcheapentry

func (h lbcheap) Len() int {
	return len(h)
}

func (h lbcheap) Less(i, j int) bool {
	return h[i].CPUPercent < h[j].CPUPercent
}

func (h lbcheap) Swap(i, j int) {
	// need to swap heapidx
	h[i].heapidx, h[j].heapidx = h[j].heapidx, h[i].heapidx
	h[i], h[j] = h[j], h[i]
}

func (h *lbcheap) Push(x interface{}) {
	entry := x.(*lbcheapentry)
	entry.heapidx = len(*h)
	*h = append(*h, entry)
}

func (h *lbcheap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h lbcheap) validateHeapIndexes() {
	if !config.Debug() {
		return
	}

	gameids := []uint16{}
	for i := 0; i < len(h); i++ {
		if h[i].heapidx != i {
			gwlog.Fatalf("lbcheap elem at index %d but has heapidx=%d", i, h[i].heapidx)
		}
		if i > 0 {
			if h.Less(i, 0) {
				gwlog.Fatalf("lbcheap elem at index 0 is not min")
			}
		}
		gameids = append(gameids, h[i].gameid)
	}
	//gwlog.Infof("lbcheap: gameids: %v", gameids)
}
func (h *lbcheap) chosen(idx int) {
	entry := (*h)[idx]
	if entry.CPUPercent < entry.origCPUPercent+10 {
		entry.CPUPercent += 0.1
		heap.Fix(h, idx)
	}
}
