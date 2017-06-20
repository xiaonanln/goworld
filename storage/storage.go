package storage

import (
	"github.com/xiaonanln/goSyncQueue"
	"github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/storage/backend/filesystem"
	"github.com/xiaonanln/goworld/storage/common"
)

var (
	storageEngine  storage_common.EntityStorage
	operationQueue = sync_queue.NewSyncQueue()
)

const ( // storage request types
	SR_SAVE = iota
	SR_LOAD = iota
)

type saveRequest struct {
	TypeName string
	EntityID common.EntityID
	Data     interface{}
}

type loadRequest struct {
	TypeName string
	EntityID common.EntityID
	Callback LoadCallbackFunc
}

type LoadCallbackFunc func(data interface{}, err error)

func Save(typeName string, entityID common.EntityID, data interface{}) {
	operationQueue.Push(saveRequest{
		TypeName: typeName,
		EntityID: entityID,
		Data:     data,
	})
}

func Load(typeName string, entityID common.EntityID, callback LoadCallbackFunc) {
	operationQueue.Push(loadRequest{
		TypeName: typeName,
		EntityID: entityID,
		Callback: callback,
	})
}

func ListEntityIDs(typeName string) []common.EntityID {
	eids, err := storageEngine.List(typeName)
	if err != nil {
		gwlog.TraceError("ListEntityIDs %s failed: %s", typeName, err)
	}
	return eids
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
					continue // always retry if fail
				} else {
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

			timer.AddCallback(0, func() {
				loadReq.Callback(data, err)
			})
		} else {
			gwlog.Panicf("storage: unknown operation: %v", op)
		}
	}
}
