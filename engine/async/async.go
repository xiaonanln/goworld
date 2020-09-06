package async

import (
	"sync"

	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/gwutils"
	"github.com/xiaonanln/goworld/engine/post"
)

var (
	numAsyncJobWorkersRunning sync.WaitGroup
)

// AsyncCallback is a function which will be called after async job is finished with result and error
type AsyncCallback func(res interface{}, err error)

func (ac AsyncCallback) callback(res interface{}, err error) {
	if ac != nil {
		post.Post(func() {
			ac(res, err)
		})
	}
}

// AsyncRoutine is a function that will be executed in the async goroutine and its result and error will be passed to AsyncCallback
type AsyncRoutine func() (res interface{}, err error)

type asyncJobWorker struct {
	jobQueue chan asyncJobItem
}

type asyncJobItem struct {
	routine  AsyncRoutine
	callback AsyncCallback
}

func newAsyncJobWorker() *asyncJobWorker {
	ajw := &asyncJobWorker{
		jobQueue: make(chan asyncJobItem, consts.ASYNC_JOB_QUEUE_MAXLEN),
	}
	numAsyncJobWorkersRunning.Add(1)
	go ajw.loop()
	return ajw
}

func (ajw *asyncJobWorker) appendJob(routine AsyncRoutine, callback AsyncCallback) {
	ajw.jobQueue <- asyncJobItem{routine, callback}
}

func (ajw *asyncJobWorker) loop() {
	defer numAsyncJobWorkersRunning.Done()

	gwutils.RepeatUntilPanicless(func() {
		for item := range ajw.jobQueue {
			res, err := item.routine()
			item.callback.callback(res, err)
		}
	})
}

var (
	asyncJobWorkersLock sync.RWMutex
	asyncJobWorkers     = map[string]*asyncJobWorker{}
)

func getAsyncJobWorker(group string) (ajw *asyncJobWorker) {
	asyncJobWorkersLock.RLock()
	ajw = asyncJobWorkers[group]
	asyncJobWorkersLock.RUnlock()

	if ajw == nil {
		asyncJobWorkersLock.Lock()
		ajw = asyncJobWorkers[group]
		if ajw == nil {
			ajw = newAsyncJobWorker()
			asyncJobWorkers[group] = ajw
		}
		asyncJobWorkersLock.Unlock()
	}
	return
}

// AppendAsyncJob append an async job to be executed asyncly (not in the game goroutine)
func AppendAsyncJob(group string, routine AsyncRoutine, callback AsyncCallback) {
	ajw := getAsyncJobWorker(group)
	ajw.appendJob(routine, callback)
}

// WaitClear wait for all async job workers to finish (should only be called in the game goroutine)
func WaitClear() bool {
	var cleared bool
	// Close all job queue workers
	gwlog.Infof("Waiting for all async job workers to be cleared ...")
	asyncJobWorkersLock.Lock()
	if len(asyncJobWorkers) > 0 {
		for group, alw := range asyncJobWorkers {
			close(alw.jobQueue)
			gwlog.Infof("\tclear %s", group)
		}
		asyncJobWorkers = map[string]*asyncJobWorker{}
		cleared = true
	}
	asyncJobWorkersLock.Unlock()

	// wait for all job workers to quit
	numAsyncJobWorkersRunning.Wait()
	return cleared
}
