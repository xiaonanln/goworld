package kvdb

import (
	"time"

	"io"

	"strconv"

	"github.com/xiaonanln/goworld/engine/async"
	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/kvdb/backend/kvdb_mongodb"
	"github.com/xiaonanln/goworld/engine/kvdb/backend/kvdbredis"
	"github.com/xiaonanln/goworld/engine/kvdb/backend/kvdbrediscluster"
	"github.com/xiaonanln/goworld/engine/kvdb/types"
)

const (
	_KVDB_ASYNC_JOB_GROUP = "_kvdb"
)

var (
	kvdbEngine kvdbtypes.KVDBEngine
)

// KVDBGetCallback is type of KVDB Get callback
type KVDBGetCallback func(val string, err error)

// KVDBPutCallback is type of KVDB Get callback
type KVDBPutCallback func(err error)

// KVDBGetRangeCallback is type of KVDB GetRange callback
type KVDBGetRangeCallback func(items []kvdbtypes.KVItem, err error)

// KVDBGetOrPutCallback is type of KVDB GetOrPut callback
type KVDBGetOrPutCallback func(oldVal string, err error)

// Initialize the KVDB
//
// Called by game server engine
func Initialize() {
	kvdbCfg := config.GetKVDB()
	if kvdbCfg.Type == "" {
		return
	}

	gwlog.Infof("KVDB initializing, config:\n%s", config.DumpPretty(kvdbCfg))
	assureKVDBEngineReady()
}

func assureKVDBEngineReady() (err error) {
	if kvdbEngine != nil { // connection is valid
		return
	}

	kvdbCfg := config.GetKVDB()

	if kvdbCfg.Type == "mongodb" {
		kvdbEngine, err = kvdbmongo.OpenMongoKVDB(kvdbCfg.Url, kvdbCfg.DB, kvdbCfg.Collection)
	} else if kvdbCfg.Type == "redis" {
		var dbindex = -1
		if kvdbCfg.DB != "" {
			dbindex, err = strconv.Atoi(kvdbCfg.DB)
			if err != nil {
				return err
			}
		}
		kvdbEngine, err = kvdbredis.OpenRedisKVDB(kvdbCfg.Url, dbindex)
	} else if kvdbCfg.Type == "redis_cluster" {
		kvdbEngine, err = kvdbrediscluster.OpenRedisKVDB(kvdbCfg.StartNodes.ToList())
	} else {
		gwlog.Fatalf("KVDB type %s is not implemented", kvdbCfg.Type)
	}
	return
}

// Get gets value of key from KVDB, returns in callback
func Get(key string, callback KVDBGetCallback) {
	var ac async.AsyncCallback
	if callback != nil {
		ac = func(res interface{}, err error) {
			if err != nil {
				callback("", err)
			} else {
				callback(res.(string), nil)
			}
		}
	}
	async.AppendAsyncJob(_KVDB_ASYNC_JOB_GROUP, kvdbRoutine(func() (res interface{}, err error) {
		res, err = kvdbEngine.Get(key)
		return
	}), ac)
}

func kvdbRoutine(r func() (res interface{}, err error)) func() (res interface{}, err error) {
	kvdbroutine := func() (res interface{}, err error) {
		for {
			err := assureKVDBEngineReady()
			if err == nil {
				break
			} else {
				gwlog.Errorf("KVDB engine is not ready: %s", err)
				time.Sleep(time.Second)
			}
		}

		res, err = r()

		if err != nil && kvdbEngine.IsConnectionError(err) {
			kvdbEngine.Close()
			kvdbEngine = nil
		}
		return
	}

	return kvdbroutine
}

// Put puts key-value item to KVDB, returns in callback
func Put(key string, val string, callback KVDBPutCallback) {
	var ac async.AsyncCallback
	if callback != nil {
		ac = func(res interface{}, err error) {
			callback(err)
		}
	}

	async.AppendAsyncJob(_KVDB_ASYNC_JOB_GROUP, kvdbRoutine(func() (res interface{}, err error) {
		err = kvdbEngine.Put(key, val)
		return
	}), ac)
}

// GetOrPut gets value of key from KVDB, if val not exists or is "", put key-value to KVDB.
func GetOrPut(key string, val string, callback KVDBGetOrPutCallback) {
	var ac async.AsyncCallback
	if callback != nil {
		ac = func(res interface{}, err error) {
			if err == nil {
				callback(res.(string), err)
			} else {
				callback("", err)
			}
		}
	}

	async.AppendAsyncJob(_KVDB_ASYNC_JOB_GROUP, kvdbRoutine(func() (res interface{}, err error) {
		oldVal, err := kvdbEngine.Get(key)
		if err == nil {
			if oldVal == "" {
				err = kvdbEngine.Put(key, val)
			}
		}

		return oldVal, err
	}), ac)
}

// GetRange retrives key-value items of specified key range, returns in callback
func GetRange(beginKey string, endKey string, callback KVDBGetRangeCallback) {
	var ac async.AsyncCallback
	if callback != nil {
		ac = func(res interface{}, err error) {
			if err == nil {
				callback(res.([]kvdbtypes.KVItem), nil)
			} else {
				callback(nil, err)
			}
		}
	}

	async.AppendAsyncJob(_KVDB_ASYNC_JOB_GROUP, kvdbRoutine(func() (res interface{}, err error) {
		it, err := kvdbEngine.Find(beginKey, endKey)
		if err != nil {
			return nil, err
		}

		var items []kvdbtypes.KVItem
		for {
			item, err := it.Next()
			if err == io.EOF {
				break
			}

			if err != nil {
				return nil, err
			}

			items = append(items, item)
		}
		return items, nil
	}), ac)
}

// NextLargerKey finds the next key that is larger than the specified key,
// but smaller than any other keys that is larger than the specified key
func NextLargerKey(key string) string {
	return key + "\x00" // the next string that is larger than key, but smaller than any other keys > key
}
