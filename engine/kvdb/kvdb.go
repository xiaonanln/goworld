package kvdb

import (
	"time"

	"io"

	"strconv"

	"github.com/xiaonanln/go-xnsyncutil/xnsyncutil"
	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/kvdb/backend/kvdb_mongodb"
	"github.com/xiaonanln/goworld/engine/kvdb/backend/kvdbredis"
	"github.com/xiaonanln/goworld/engine/kvdb/types"
	"github.com/xiaonanln/goworld/engine/opmon"
	"github.com/xiaonanln/goworld/engine/post"
)

var (
	kvdbEngine     kvdbtypes.KVDBEngine
	kvdbOpQueue    *xnsyncutil.SyncQueue
	kvdbTerminated *xnsyncutil.OneTimeCond
)

// KVDBGetCallback is type of KVDB Get callback
type KVDBGetCallback func(val string, err error)

// KVDBPutCallback is type of KVDB Get callback
type KVDBPutCallback func(err error)

// KVDBGetRangeCallback is type of KVDB GetRange callback
type KVDBGetRangeCallback func(items []kvdbtypes.KVItem, err error)

// Initialize the KVDB
//
// Called by game server engine
func Initialize() {
	kvdbCfg := config.GetKVDB()
	if kvdbCfg.Type == "" {
		return
	}

	gwlog.Infof("KVDB initializing, config:\n%s", config.DumpPretty(kvdbCfg))
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
		kvdbEngine, err = kvdbmongo.OpenMongoKVDB(kvdbCfg.Url, kvdbCfg.DB, kvdbCfg.Collection)
	} else if kvdbCfg.Type == "redis" {
		var dbindex int = -1
		if kvdbCfg.DB != "" {
			dbindex, err = strconv.Atoi(kvdbCfg.DB)
			if err != nil {
				return err
			}
		}
		kvdbEngine, err = kvdbredis.OpenRedisKVDB(kvdbCfg.Url, dbindex)
	} else {
		gwlog.Fatalf("KVDB type %s is not implemented", kvdbCfg.Type)
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

// Get gets value of key from KVDB, returns in callback
func Get(key string, callback KVDBGetCallback) {
	kvdbOpQueue.Push(&getReq{
		key, callback,
	})
	checkOperationQueueLen()
}

// Put puts key-value item to KVDB, returns in callback
func Put(key string, val string, callback KVDBPutCallback) {
	kvdbOpQueue.Push(&putReq{
		key, val, callback,
	})
	checkOperationQueueLen()
}

// GetRange retrives key-value items of specified key range, returns in callback
func GetRange(beginKey string, endKey string, callback KVDBGetRangeCallback) {
	kvdbOpQueue.Push(&getRangeReq{
		beginKey, endKey, callback,
	})
	checkOperationQueueLen()
}

// NextLargerKey finds the next key that is larger than the specified key,
// but smaller than any other keys that is larger than the specified key
func NextLargerKey(key string) string {
	return key + "\x00" // the next string that is larger than key, but smaller than any other keys > key
}

// Close the KVDB
func Close() {
	kvdbOpQueue.Close()
}

var recentWarnedQueueLen = 0

func checkOperationQueueLen() {
	qlen := kvdbOpQueue.Len()
	if qlen > 100 && qlen%100 == 0 && recentWarnedQueueLen != qlen {
		gwlog.Warnf("KVDB operation queue length = %d", qlen)
		recentWarnedQueueLen = qlen
	}
}

func kvdbRoutine() {
	for {
		err := assureKVDBEngineReady()
		if err != nil {
			gwlog.Errorf("KVDB engine is not ready: %s", err)
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

// WaitTerminated waits for KVDB to terminate
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
	var items []kvdbtypes.KVItem
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
