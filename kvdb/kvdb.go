package kvdb

import (
	"github.com/xiaonanln/goSyncQueue"
	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/kvdb/backend/mongodb"
	"github.com/xiaonanln/vacuum/netutil"
)

var (
	kvdbEngine  KVDBEngine
	kvdbOpQueue sync_queue.SyncQueue
)

type KVDBEngine interface {
	Get(key string) (val string, err error)
	Put(key string, val string) (err error)
}

type KVDBGetCallback func(val string, err error)
type KVDBPutCallback func(err error)

func Initialize() {
	var err error
	kvdbCfg := config.GetKVDB()
	if kvdbCfg.Type == "" {
		return // kvdb not enabled
	}

	if kvdbCfg.Type == "mongodb" {
		kvdbEngine, err = kvdb_mongo.OpenMongoKVDB(kvdbCfg.Url, kvdbCfg.DB, kvdbCfg.Collection)
		if err != nil {
			gwlog.Panic(err)
		}
	}

	kvdbOpQueue = sync_queue.NewSyncQueue()
	go netutil.ServeForever(kvdbRoutine)
}

type getReq struct {
	key      string
	callback KVDBGetCallback
}

type putReq struct {
	key      string
	val      string
	callback KVDBPutCallback
}

func Get(key string, callback KVDBGetCallback) {
	kvdbOpQueue.Push(getReq{
		key,
		callback,
	})
	return
}

func Put(key string, val string, callback KVDBPutCallback) {
	kvdbOpQueue.Push(putReq{
		key,
		val,
		callback,
	})
	return
}

func kvdbRoutine() {
	for {
		req := kvdbOpQueue.Pop()
		if getReq, ok := req.(getReq); ok {
			val, err := kvdbEngine.Get(getReq.key)
			if getReq.callback != nil {
				getReq.callback(val, err)
			}
		} else if putReq, ok := req.(putReq); ok {
			err := kvdbEngine.Put(putReq.key, putReq.val)
			if putReq.callback != nil {
				putReq.callback(err)
			}
		}
	}
}
