package kvdb

import (
	"time"

	"io"

	"strconv"

	"github.com/xiaonanln/goSyncQueue"
	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/kvdb/backend/kvdb_mongodb"
	"github.com/xiaonanln/goworld/kvdb/backend/kvdb_redis"
	. "github.com/xiaonanln/goworld/kvdb/types"
	"github.com/xiaonanln/goworld/netutil"
	"github.com/xiaonanln/goworld/opmon"
	"github.com/xiaonanln/goworld/post"
)

var (
	kvdbEngine  KVDBEngine
	kvdbOpQueue sync_queue.SyncQueue
)

type KVDBEngine interface {
	Get(key string) (val string, err error)
	Put(key string, val string) (err error)
	Find(beginKey string, endKey string) Iterator
}

type KVDBGetCallback func(val string, err error)
type KVDBPutCallback func(err error)
type KVDBGetRangeCallback func(items []KVItem, err error)

// Initialize the KVDB
//
// Called by game server engine
func Initialize() {
	var err error
	kvdbCfg := config.GetKVDB()
	if kvdbCfg.Type == "" {
		return // kvdb not enabled
	}

	gwlog.Info("KVDB initializing, config:\n%s", config.DumpPretty(kvdbCfg))

	if kvdbCfg.Type == "mongodb" {
		kvdbEngine, err = kvdb_mongo.OpenMongoKVDB(kvdbCfg.Url, kvdbCfg.DB, kvdbCfg.Collection)
		if err != nil {
			gwlog.Panic(err)
		}
	} else if kvdbCfg.Type == "redis" {
		dbindex, err := strconv.Atoi(kvdbCfg.DB)
		if err != nil {
			gwlog.Panic(err)
		}

		kvdbEngine, err = kvdb_redis.OpenRedisKVDB(kvdbCfg.Host, dbindex)
		if err != nil {
			gwlog.Panic(err)
		}
	} else {
		gwlog.Fatal("KVDB type %s is not implemented", kvdbCfg.Type)
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

type getRangeReq struct {
	beginKey string
	endKey   string
	callback KVDBGetRangeCallback
}

func Get(key string, callback KVDBGetCallback) {
	kvdbOpQueue.Push(&getReq{
		key, callback,
	})
	checkOperationQueueLen()
}

func Put(key string, val string, callback KVDBPutCallback) {
	kvdbOpQueue.Push(&putReq{
		key, val, callback,
	})
	checkOperationQueueLen()
}

func GetRange(beginKey string, endKey string, callback KVDBGetRangeCallback) {
	kvdbOpQueue.Push(&getRangeReq{
		beginKey, endKey, callback,
	})
	checkOperationQueueLen()
}

func NextLargerKey(key string) string {
	return key + "\x00" // the next string that is larger than key, but smaller than any other keys > key
}

var recentWarnedQueueLen = 0

func checkOperationQueueLen() {
	qlen := kvdbOpQueue.Len()
	if qlen > 100 && qlen%100 == 0 && recentWarnedQueueLen != qlen {
		gwlog.Warn("KVDB operation queue length = %d", qlen)
		recentWarnedQueueLen = qlen
	}
}

func kvdbRoutine() {
	for {
		req := kvdbOpQueue.Pop()
		var op *opmon.Operation
		if getReq, ok := req.(*getReq); ok {
			op = opmon.StartOperation("kvdb.get")
			handleGetReq(getReq)
		} else if putReq, ok := req.(*putReq); ok {
			op = opmon.StartOperation("kvdb.put")
			handlePutReq(putReq)
		} else if getRangeReq, ok := req.(*getRangeReq); ok {
			op = opmon.StartOperation("kvdb.getRange")
			handleGetRangeReq(getRangeReq)
		}
		op.Finish(time.Millisecond * 100)
	}
}

func handleGetReq(getReq *getReq) {
	val, err := kvdbEngine.Get(getReq.key)
	if getReq.callback != nil {
		post.Post(func() {
			getReq.callback(val, err)
		})
	}
}

func handlePutReq(putReq *putReq) {
	err := kvdbEngine.Put(putReq.key, putReq.val)
	if putReq.callback != nil {
		post.Post(func() {
			putReq.callback(err)
		})
	}
}

func handleGetRangeReq(getRangeReq *getRangeReq) {
	it := kvdbEngine.Find(getRangeReq.beginKey, getRangeReq.endKey)
	var items []KVItem
	for {
		item, err := it.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			if getRangeReq.callback != nil {
				post.Post(func() {
					getRangeReq.callback(nil, err)
				})
			}
			return
		}

		items = append(items, item)
	}

	if getRangeReq.callback != nil {
		post.Post(func() {
			getRangeReq.callback(items, nil)
		})
	}
}
