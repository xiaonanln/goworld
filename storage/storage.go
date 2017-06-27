package storage

import (
	"time"

	"github.com/xiaonanln/goSyncQueue"
	"github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/storage/backend/filesystem"
)

type EntityStorage interface {
	List(typeName string) ([]common.EntityID, error)
	Write(typeName string, entityID common.EntityID, data interface{}) error
	Read(typeName string, entityID common.EntityID) (interface{}, error)
	Exists(typeName string, entityID common.EntityID) (bool, error)
}

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
}

func Load(typeName string, entityID common.EntityID, callback LoadCallbackFunc) {
	operationQueue.Push(loadRequest{
		TypeName: typeName,
		EntityID: entityID,
		Callback: callback,
	})
}

func Exists(typeName string, entityID common.EntityID, callback ExistsCallbackFunc) {
	operationQueue.Push(existsRequest{
		TypeName: typeName,
		EntityID: entityID,
		Callback: callback,
	})
}

func ListEntityIDs(typeName string, callback ListCallbackFunc) {
	operationQueue.Push(listEntityIDsRequest{
		TypeName: typeName,
		Callback: callback,
	})
}

func Initialize() {
	var err error
	cfg := config.GetStorage()
	if cfg.Type == "filesystem" {
		storageEngine, err = entity_storage_filesystem.OpenDirectory(cfg.Directory)
		if err != nil {
			gwlog.Panic(err)
		}
	} else {
		gwlog.Panicf("unknown storage type: %s", cfg.Type)
	}

	go storageRoutine()
}

func storageRoutine() {
	defer func() {
		err := recover()
		gwlog.TraceError("storage routine paniced: %s, restarting ...", err)
		go storageRoutine() // restart the storage routine
	}()

	for {
		op := operationQueue.Pop()
		if saveReq, ok := op.(saveRequest); ok {
			// handle save request
			for {
				if consts.DEBUG_SAVE_LOAD {
					gwlog.Debug("storage: SAVING %s %s ...", saveReq.TypeName, saveReq.EntityID)
				}
				err := storageEngine.Write(saveReq.TypeName, saveReq.EntityID, saveReq.Data)
				if err != nil {
					// save failed ?
					gwlog.Error("storage: save failed: %s\nData: %v", err, saveReq.Data)
					time.Sleep(time.Second) // wait for 1 second to retry
					continue                // always retry if fail
				} else {
					if saveReq.Callback != nil {
						timer.AddCallback(0, func() {
							saveReq.Callback()
						})
					}
					break
				}
			}
		} else if loadReq, ok := op.(loadRequest); ok {
			// handle load request
			gwlog.Debug("storage: LOADING %s %s ...", loadReq.TypeName, loadReq.EntityID)
			data, err := storageEngine.Read(loadReq.TypeName, loadReq.EntityID)
			if err != nil {
				// save failed ?
				gwlog.TraceError("storage: load %s %s failed: %s", loadReq.TypeName, loadReq.EntityID, err)
				data = nil
			}

			if loadReq.Callback != nil {
				timer.AddCallback(0, func() {
					loadReq.Callback(data, err)
				})
			}
		} else if existsReq, ok := op.(existsRequest); ok {
			exists, err := storageEngine.Exists(existsReq.TypeName, existsReq.EntityID)
			if existsReq.Callback != nil {
				existsReq.Callback(exists, err)
			}
		} else if listReq, ok := op.(listEntityIDsRequest); ok {
			eids, err := storageEngine.List(listReq.TypeName)
			if err != nil {
				gwlog.TraceError("ListEntityIDs %s failed: %s", listReq.TypeName, err)
			}
			if listReq.Callback != nil {
				listReq.Callback(eids, err)
			}
		} else {
			gwlog.Panicf("storage: unknown operation: %v", op)
		}
	}
}
