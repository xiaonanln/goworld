package opmon

import (
	"sync"
	"time"

	"github.com/xiaonanln/goworld/gwlog"
)

var (
	operationAllocPool = sync.Pool{
		New: func() interface{} {
			return &operation{}
		},
	}

	monitor = newMonitor()
)

type Monitor struct {
	sync.Mutex
	opCounter   map[string]uint64
	opTotalTime map[string]time.Duration
}

func newMonitor() *Monitor {
	m := &Monitor{
		opCounter:   map[string]uint64{},
		opTotalTime: map[string]time.Duration{},
	}
	return m
}

func (monitor *Monitor) record(opname string, duration time.Duration) {
	monitor.Lock()
	monitor.opCounter[opname] += 1
	monitor.opTotalTime[opname] += duration
	monitor.Unlock()
}

type operation struct {
	name      string
	startTime time.Time
}

func StartOperation(operationName string) *operation {
	op := operationAllocPool.Get().(*operation)
	op.name = operationName
	op.startTime = time.Now()
	return op
}

func (op *operation) Finish(warnThreshold time.Duration) {
	takeTime := time.Now().Sub(op.startTime)
	monitor.record(op.name, takeTime)
	if takeTime >= warnThreshold {
		gwlog.Warn("opmon: operation %s takes %s > %s", op.name, takeTime, warnThreshold)
	}
}
