package kvdb

import (
	"time"

	"io"

	"strconv"

	"github.com/xiaonanln/go-xnsyncutil/xnsyncutil"
	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/kvdbne/kvdb/backend/kvdb_mongodb"
	"github.com/xiaonanln/goworld/engine/kvdbne/kvdb/backend/kvdb_redis"
	. "github.com/xiaonanln/goworld/engine/kvdbne/kvdb/types"
	"github.com/xiaonanln/goworld/engine/opmon"
	"github.com/xiaonanln/goworld/engine/post"
)

var (
	kvdbEngine     KVDBEngine
	kvdbOpQueue    *xnsyncutil.SyncQueue
	kvdbTerminated *xnsyncutil.OneTimeCond
)

type KVDBGetCallback func(val string, err error)
type KVDBPutCallback func(err error)
type KVDBGetRangeCallback func(items []KVItem, err error)

// Initialize the KVDB
//
// Called by game server engine
func Initialize() {
	kvdbCfg := config.GetKVDB()
	if kvdbCfg.Type == "" {
		return
	}

	gwlog.Info("KVDB initializing, config:\n%s", config.DumpPretty(kvdbCfg))
	kvdbOpQueue = xnsyncutil.NewSyncQueue()
	kvdbTerminated = xnsyncutil.NewOneTimeCond()

	assureKVDBEngineReady()

	go kvdbRoutine()
}

func assureKVDBEngineReady() (err error) {
	if kvdbEngine != nil { // connection is valid
		return
	}

	kvdbCfg := config.GetKVDB()

	if kvdbCfg.Type == "mongodb" {
		kvdbEngine, err = kvdb_mongo.OpenMongoKVDB(kvdbCfg.Url, kvdbCfg.DB, kvdbCfg.Collection)
	} else if kvdbCfg.Type == "redis" {
		dbindex, err := strconv.Atoi(kvdbCfg.DB)
		if err == nil {
			kvdbEngine, err = kvdb_redis.OpenRedisKVDB(kvdbCfg.Host, dbindex)
		}
	} else {
		gwlog.Fatal("KVDB type %s is not implemented", kvdbCfg.Type)
	}
	return
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

func Close() {
	kvdbOpQueue.Close()
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
		err := assureKVDBEngineReady()
		if err != nil {
			gwlog.Error("KVDB engine is not ready: %s", err)
			time.Sleep(time.Second)
			continue
		}

		req := kvdbOpQueue.Pop()
		if req == nil { // queue is closed, returning nil
			kvdbEngine.Close()
			break
		}

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

	kvdbTerminated.Signal()
}

func WaitTerminated() {
	kvdbTerminated.Wait()
}

func handleGetReq(getReq *getReq) {
	val, err := kvdbEngine.Get(getReq.key)
	if getReq.callback != nil {
		post.Post(func() {
			getReq.callback(val, err)
		})
	}

	if err != nil && kvdbEngine.IsEOF(err) {
		kvdbEngine.Close()
		kvdbEngine = nil
	}
}

func handlePutReq(putReq *putReq) {
	err := kvdbEngine.Put(putReq.key, putReq.val)
	if putReq.callback != nil {
		post.Post(func() {
			putReq.callback(err)
		})
	}

	if err != nil && kvdbEngine.IsEOF(err) {
		kvdbEngine.Close()
		kvdbEngine = nil
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
