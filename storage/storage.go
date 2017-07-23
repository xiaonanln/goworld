package storage

import (
	"time"

	"os"

	"strconv"

	"github.com/xiaonanln/goSyncQueue"
	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/opmon"
	"github.com/xiaonanln/goworld/post"
	"github.com/xiaonanln/goworld/storage/backend/filesystem"
	"github.com/xiaonanln/goworld/storage/backend/mongodb"
	"github.com/xiaonanln/goworld/storage/backend/redis"
	. "github.com/xiaonanln/goworld/storage/storage_common"
)

var (
	storageEngine  EntityStorage
	operationQueue = sync_queue.NewSyncQueue()
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

type SaveCallbackFunc func()
type LoadCallbackFunc func(data interface{}, err error)
type ExistsCallbackFunc func(exists bool, err error)
type ListCallbackFunc func([]common.EntityID, error)

func Save(typeName string, entityID common.EntityID, data interface{}, callback SaveCallbackFunc) {
	operationQueue.Push(saveRequest{
		TypeName: typeName,
		EntityID: entityID,
		Data:     data,
		Callback: callback,
	})
	checkOperationQueueLen()
}

func Load(typeName string, entityID common.EntityID, callback LoadCallbackFunc) {
	operationQueue.Push(loadRequest{
		TypeName: typeName,
		EntityID: entityID,
		Callback: callback,
	})
	checkOperationQueueLen()
}

func Exists(typeName string, entityID common.EntityID, callback ExistsCallbackFunc) {
	operationQueue.Push(existsRequest{
		TypeName: typeName,
		EntityID: entityID,
		Callback: callback,
	})
	checkOperationQueueLen()
}

func ListEntityIDs(typeName string, callback ListCallbackFunc) {
	operationQueue.Push(listEntityIDsRequest{
		TypeName: typeName,
		Callback: callback,
	})
	checkOperationQueueLen()
}

func GetQueueLen() int {
	return operationQueue.Len()
}

var recentWarnedQueueLen = 0

func checkOperationQueueLen() {
	qlen := operationQueue.Len()
	if qlen > 100 && qlen%100 == 0 && recentWarnedQueueLen != qlen {
		gwlog.Warn("Storage operation queue length = %d", qlen)
		recentWarnedQueueLen = qlen
	}
}

func Initialize() {
	err := assureStorageEngineReady()
	if err != nil {
		gwlog.Fatal("Storage engine is not ready: %s", err)
	}
	go storageRoutine()
}

func assureStorageEngineReady() (err error) {
	if storageEngine != nil {
		return
	}

	cfg := config.GetStorage()
	if cfg.Type == "filesystem" {
		storageEngine, err = entity_storage_filesystem.OpenDirectory(cfg.Directory)
	} else if cfg.Type == "mongodb" {
		storageEngine, err = entity_storage_mongodb.OpenMongoDB(cfg.Url, cfg.DB)
	} else if cfg.Type == "redis" {
		var dbindex int
		if dbindex, err = strconv.Atoi(cfg.DB); err == nil {
			storageEngine, err = entity_storage_redis.OpenRedis(cfg.Host, dbindex)
		}
	} else {
		gwlog.Panicf("unknown storage type: %s", cfg.Type)
		if consts.DEBUG_MODE {
			os.Exit(2)
		}
	}

	return
}

func storageRoutine() {
	defer func() {
		err := recover()
		gwlog.TraceError("storage routine paniced: %s, restarting ...", err)
		if consts.DEBUG_MODE {
			os.Exit(2)
		}
		go storageRoutine() // restart the storage routine
	}()

	for {
		err := assureStorageEngineReady()
		if err != nil {
			gwlog.Error("Storage engine is not ready: %s", err)
			time.Sleep(time.Second)
			continue
		}

		op := operationQueue.Pop()
		var monop *opmon.Operation
		if saveReq, ok := op.(saveRequest); ok {
			// handle save request
			monop = opmon.StartOperation("storage.save")
			for {
				if consts.DEBUG_SAVE_LOAD {
					gwlog.Debug("storage: SAVING %s %s ...", saveReq.TypeName, saveReq.EntityID)
				}
				err := assureStorageEngineReady()
				if err != nil {
					gwlog.Error("Storage engine is not ready: %s", err)
					time.Sleep(time.Second) // wait for 1 second to retry
					continue
				}

				if storageEngine == nil {
					gwlog.Fatal("storage engine is nil")
				}

				err = storageEngine.Write(saveReq.TypeName, saveReq.EntityID, saveReq.Data)
				if err != nil {
					// save failed ?
					gwlog.Error("storage: save failed: %s", err)

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
			gwlog.Debug("storage: LOADING %s %s ...", loadReq.TypeName, loadReq.EntityID)
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
