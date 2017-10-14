package async

import (
	"sync"

	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/netutil"
	"github.com/xiaonanln/goworld/engine/post"
	"golang.org/x/net/context"
)

var (
	asyncRunning, asyncCancelRunning = context.WithCancel(context.Background())
	numAsyncJobWorkersRunning        sync.WaitGroup
)

type AsyncCallback func(res interface{}, err error)

func (ac AsyncCallback) Callback(res interface{}, err error) {
	if ac != nil {
		post.Post(func() {
			ac(res, err)
		})
	}
}

type AsyncRoutine func() (res interface{}, err error)

type AsyncJobWorker struct {
	jobQueue chan asyncJobItem
}

type asyncJobItem struct {
	routine  AsyncRoutine
	callback AsyncCallback
}

func newAsyncJobWorker() *AsyncJobWorker {
	ajw := &AsyncJobWorker{
		jobQueue: make(chan asyncJobItem, consts.ASYNC_JOB_QUEUE_MAXLEN),
	}
	numAsyncJobWorkersRunning.Add(1)
	go netutil.ServeForever(ajw.loop)
	return ajw
}

func (ajw *AsyncJobWorker) appendJob(routine AsyncRoutine, callback AsyncCallback) {
	ajw.jobQueue <- asyncJobItem{routine, callback}
}

func (ajw *AsyncJobWorker) loop() {
	for item := range ajw.jobQueue {
		res, err := item.routine()
		item.callback.Callback(res, err)
	}
	numAsyncJobWorkersRunning.Done()
}

var (
	asyncJobWorkersLock sync.RWMutex
	asyncJobWorkers     = map[string]*AsyncJobWorker{}
)

func getAsyncJobWorker(group string) (ajw *AsyncJobWorker) {
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

func AppendAsyncJob(group string, routine AsyncRoutine, callback AsyncCallback) {
	ajw := getAsyncJobWorker(group)
	ajw.appendJob(routine, callback)
}

func Shutdown() {
	// Close all job queue workers
	asyncJobWorkersLock.Lock()
	for _, alw := range asyncJobWorkers {
		close(alw.jobQueue)
	}
	asyncJobWorkers = map[string]*AsyncJobWorker{}
	asyncJobWorkersLock.Unlock()

	// wait for all job workers to quit
	numAsyncJobWorkersRunning.Wait()
}
