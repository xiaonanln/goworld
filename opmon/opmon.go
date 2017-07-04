package opmon

import (
	"sync"
	"time"

	"sort"

	"fmt"
	"os"

	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/gwlog"
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

type Monitor struct {
	sync.Mutex
	opInfos map[string]*_OpInfo
}

func newMonitor() *Monitor {
	m := &Monitor{
		opInfos: map[string]*_OpInfo{},
	}
	return m
}

func (monitor *Monitor) record(opname string, duration time.Duration) {
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

func (monitor *Monitor) Dump() {
	type _T struct {
		name string
		info _OpInfo
	}
	copyOpInfos := []_T{}
	monitor.Lock()

	for opname, opinfo := range monitor.opInfos {
		copyOpInfos = append(copyOpInfos, _T{opname, *opinfo})
	}

	monitor.Unlock()

	sort.Slice(copyOpInfos, func(i, j int) bool {
		_t1 := copyOpInfos[i]
		_t2 := copyOpInfos[j]
		return _t1.name < _t2.name
	})
	fmt.Fprintf(os.Stderr, "=====================================================================================\n")
	for _, _t := range copyOpInfos {
		opname, opinfo := _t.name, &_t.info
		fmt.Fprintf(os.Stderr, "%-30sx%-10d AVG %-10s MAX %-10s\n", opname, opinfo.count, opinfo.totalDuration/time.Duration(opinfo.count), opinfo.maxDuration)
	}
}

type Operation struct {
	name      string
	startTime time.Time
}

func StartOperation(operationName string) *Operation {
	op := operationAllocPool.Get().(*Operation)
	op.name = operationName
	op.startTime = time.Now()
	return op
}

func (op *Operation) Finish(warnThreshold time.Duration) {
	takeTime := time.Now().Sub(op.startTime)
	monitor.record(op.name, takeTime)
	if takeTime >= warnThreshold {
		gwlog.Warn("opmon: operation %s takes %s > %s", op.name, takeTime, warnThreshold)
	}
}
