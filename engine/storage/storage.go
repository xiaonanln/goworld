package storage

import (
	"time"

	"github.com/xiaonanln/go-xnsyncutil/xnsyncutil"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/opmon"
	"github.com/xiaonanln/goworld/engine/post"
	"github.com/xiaonanln/goworld/engine/storage/backend/mongodb"
	"github.com/xiaonanln/goworld/engine/storage/storage_common"
)

var (
	storageEngine            storagecommon.EntityStorage
	operationQueue           = xnsyncutil.NewSyncQueue()
	storageRoutineTerminated = xnsyncutil.NewOneTimeCond()
)

type saveRequest struct {
	TypeName string
	EntityID common.EntityID
	Data     interface{}
	Callback SaveCallbackFunc
}

type loadRequest struct {
	TypeName string
	EntityID common.EntityID
	Callback LoadCallbackFunc
}

type existsRequest struct {
	TypeName string
	EntityID common.EntityID
	Callback ExistsCallbackFunc
}

type listEntityIDsRequest struct {
	TypeName string
	Callback ListCallbackFunc
}

// SaveCallbackFunc is the callback type of storage Save
type SaveCallbackFunc func()

// LoadCallbackFunc is the callback type of storage Load
type LoadCallbackFunc func(data interface{}, err error)

// ExistsCallbackFunc is the callback type of storage Exists
type ExistsCallbackFunc func(exists bool, err error)

// ListCallbackFunc is the callback type of storage List
type ListCallbackFunc func([]common.EntityID, error)

// Save saves entity data to storage
func Save(typeName string, entityID common.EntityID, data interface{}, callback SaveCallbackFunc) {
	operationQueue.Push(saveRequest{
		TypeName: typeName,
		EntityID: entityID,
		Data:     data,
		Callback: callback,
	})
	checkOperationQueueLen()
}

// Load loads entity data from storage
func Load(typeName string, entityID common.EntityID, callback LoadCallbackFunc) {
	operationQueue.Push(loadRequest{
		TypeName: typeName,
		EntityID: entityID,
		Callback: callback,
	})
	checkOperationQueueLen()
}

// Exists checks if entity of specified ID exists in storage
func Exists(typeName string, entityID common.EntityID, callback ExistsCallbackFunc) {
	operationQueue.Push(existsRequest{
		TypeName: typeName,
		EntityID: entityID,
		Callback: callback,
	})
	checkOperationQueueLen()
}

// ListEntityIDs returns all entity IDs in storage
//
// Return values can be large for common entity types
func ListEntityIDs(typeName string, callback ListCallbackFunc) {
	operationQueue.Push(listEntityIDsRequest{
		TypeName: typeName,
		Callback: callback,
	})
	checkOperationQueueLen()
}

var recentWarnedQueueLen = 0

func checkOperationQueueLen() {
	qlen := operationQueue.Len()
	if qlen > 100 && qlen%100 == 0 && recentWarnedQueueLen != qlen {
		gwlog.Warnf("Storage operation queue length = %d", qlen)
		recentWarnedQueueLen = qlen
	}
}

// Shutdown storage module
func Shutdown() {
	operationQueue.Close()
	storageRoutineTerminated.Wait()
}

// Initialize is called by engine to initialize storage module
func Initialize() {
	err := assureStorageEngineReady()
	if err != nil {
		gwlog.Fatalf("Storage engine is not ready: %s", err)
	}
	go storageRoutine()
}

func assureStorageEngineReady() (err error) {
	if storageEngine != nil {
		return
	}

	cfg := config.GetStorage()
	if cfg.Type == "mongodb" {
		storageEngine, err = entitystoragemongodb.OpenMongoDB(cfg.Url, cfg.DB)
	} else {
		gwlog.Panicf("unknown storage type: %s", cfg.Type)
	}

	return
}

func storageRoutine() {
	defer func() {
		err := recover()
		if err != nil {
			gwlog.TraceError("storage routine paniced: %s, restarting ...", err)
			go storageRoutine() // restart the storage routine
		} else {
			// normal quit
			storageEngine.Close()
			storageRoutineTerminated.Signal()
		}
	}()

	for {
		err := assureStorageEngineReady()
		if err != nil {
			gwlog.Errorf("Storage engine is not ready: %s", err)
			time.Sleep(time.Second)
			continue
		}

		op := operationQueue.Pop()
		if op == nil { // entity storage closed
			break
		}

		var monop *opmon.Operation
		if saveReq, ok := op.(saveRequest); ok {
			// handle save request
			monop = opmon.StartOperation("storage.save")
			for {
				if consts.DEBUG_SAVE_LOAD {
					gwlog.Debugf("storage: SAVING %s %s ...", saveReq.TypeName, saveReq.EntityID)
				}
				err := assureStorageEngineReady()
				if err != nil {
					gwlog.Errorf("Storage engine is not ready: %s", err)
					time.Sleep(time.Second) // wait for 1 second to retry
					continue
				}

				if storageEngine == nil {
					gwlog.Fatalf("storage engine is nil")
				}

				err = storageEngine.Write(saveReq.TypeName, saveReq.EntityID, saveReq.Data)
				if err != nil {
					// save failed ?
					gwlog.Errorf("storage: save failed: %s", err)

					if err != nil && storageEngine.IsEOF(err) {
						storageEngine.Close()
						storageEngine = nil
					}

					continue // always retry if fail
				} else {
					monop.Finish(time.Millisecond * 100)
					if saveReq.Callback != nil {
						post.Post(func() {
							saveReq.Callback()
						})
					}
					break
				}
			}
		} else if loadReq, ok := op.(loadRequest); ok {
			// handle load request
			gwlog.Debugf("storage: LOADING %s %s ...", loadReq.TypeName, loadReq.EntityID)
			monop = opmon.StartOperation("storage.load")
			data, err := storageEngine.Read(loadReq.TypeName, loadReq.EntityID)
			if err != nil {
				// save failed ?
				gwlog.TraceError("storage: load %s %s failed: %s", loadReq.TypeName, loadReq.EntityID, err)
				data = nil
			}

			monop.Finish(time.Millisecond * 100)
			if loadReq.Callback != nil {
				post.Post(func() {
					loadReq.Callback(data, err)
				})
			}

			if err != nil && storageEngine.IsEOF(err) {
				storageEngine.Close()
				storageEngine = nil
			}
		} else if existsReq, ok := op.(existsRequest); ok {
			monop = opmon.StartOperation("storage.exists")
			exists, err := storageEngine.Exists(existsReq.TypeName, existsReq.EntityID)
			monop.Finish(time.Millisecond * 100)
			if existsReq.Callback != nil {
				post.Post(func() {
					existsReq.Callback(exists, err)
				})
			}
			if err != nil && storageEngine.IsEOF(err) {
				storageEngine.Close()
				storageEngine = nil
			}
		} else if listReq, ok := op.(listEntityIDsRequest); ok {
			monop = opmon.StartOperation("storage.list")
			eids, err := storageEngine.List(listReq.TypeName)
			if err != nil {
				gwlog.TraceError("ListEntityIDs %s failed: %s", listReq.TypeName, err)
			}
			monop.Finish(time.Millisecond * 1000)
			if listReq.Callback != nil {
				post.Post(func() {
					listReq.Callback(eids, err)
				})
			}
			if err != nil && storageEngine.IsEOF(err) {
				storageEngine.Close()
				storageEngine = nil
			}
		} else {
			gwlog.Panicf("storage: unknown operation: %v", op)
		}
	}
}
