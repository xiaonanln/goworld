package opmon

import (
	"sync"
	"time"

	"sort"

	"fmt"
	"os"

	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

var (
	operationAllocPool = sync.Pool{
		New: func() interface{} {
			return &Operation{}
		},
	}

	monitor = newMonitor()
)

func init() {
	if consts.OPMON_DUMP_INTERVAL > 0 {
		go func() {
			for {
				time.Sleep(consts.OPMON_DUMP_INTERVAL)
				monitor.Dump()
			}
		}()
	}
}

type _OpInfo struct {
	count         uint64
	totalDuration time.Duration
	maxDuration   time.Duration
}

type _Monitor struct {
	sync.Mutex
	opInfos map[string]*_OpInfo
}

func newMonitor() *_Monitor {
	m := &_Monitor{
		opInfos: map[string]*_OpInfo{},
	}
	return m
}

func (monitor *_Monitor) record(opname string, duration time.Duration) {
	monitor.Lock()
	info := monitor.opInfos[opname]
	if info == nil {
		info = &_OpInfo{}
		monitor.opInfos[opname] = info
	}
	info.count += 1
	info.totalDuration += duration
	if duration > info.maxDuration {
		info.maxDuration = duration
	}
	monitor.Unlock()
}

func (monitor *_Monitor) Dump() {
	type _T struct {
		name string
		info *_OpInfo
	}
	var opInfos map[string]*_OpInfo
	monitor.Lock()
	opInfos = monitor.opInfos
	monitor.opInfos = map[string]*_OpInfo{} // clear to be empty
	monitor.Unlock()

	var copyOpInfos []_T
	for name, opinfo := range opInfos {
		copyOpInfos = append(copyOpInfos, _T{name, opinfo})
	}
	sort.Slice(copyOpInfos, func(i, j int) bool {
		_t1 := copyOpInfos[i]
		_t2 := copyOpInfos[j]
		return _t1.name < _t2.name
	})
	fmt.Fprint(os.Stderr, "=====================================================================================\n")
	for _, _t := range copyOpInfos {
		opname, opinfo := _t.name, _t.info
		fmt.Fprintf(os.Stderr, "%-30sx%-10d AVG %-10s MAX %-10s\n", opname, opinfo.count, opinfo.totalDuration/time.Duration(opinfo.count), opinfo.maxDuration)
	}
}

// Operation is the type of operation to be monitored
type Operation struct {
	name      string
	startTime time.Time
}

// StartOperation creates a new operation
func StartOperation(operationName string) *Operation {
	op := operationAllocPool.Get().(*Operation)
	op.name = operationName
	op.startTime = time.Now()
	return op
}

// Finish finishes the operation and records the duration of operation
func (op *Operation) Finish(warnThreshold time.Duration) {
	takeTime := time.Now().Sub(op.startTime)
	monitor.record(op.name, takeTime)
	if takeTime >= warnThreshold {
		gwlog.Warnf("opmon: operation %s takes %s > %s", op.name, takeTime, warnThreshold)
	}
}
