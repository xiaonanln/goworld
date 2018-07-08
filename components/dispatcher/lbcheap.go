package main

import (
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/proto"
)

type lbcheapentry struct {
	gameid  uint16
	lbcinfo proto.GameLBCInfo
	heapidx int // index of this entry in the heap
}

type lbcheap []*lbcheapentry

func (h lbcheap) Len() int {
	return len(h)
}

func (h lbcheap) Less(i, j int) bool {
	return h[i].lbcinfo.CPUPercent < h[j].lbcinfo.CPUPercent
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
	gameids := []uint16{}
	for i := 0; i < len(h); i++ {
		if h[i].heapidx != i {
			gwlog.Fatalf("lbcheap elem at index %d but has heapidx=%d", i, h[i].heapidx)
		}
		gameids = append(gameids, h[i].gameid)
	}
	//gwlog.Infof("lbcheap: gameids: %v", gameids)
}
