package kvdb

import (
	"math/rand"
	"strconv"
	"testing"

	"fmt"
	"github.com/xiaonanln/goworld/kvdb/backend/kvdb_mongodb"
	"github.com/xiaonanln/goworld/kvdb/backend/kvdb_redis"
	"io"
)

func TestMongoBackend_Set(t *testing.T) {
	testKVDBBackend_Set(t, openTestMongoKVDB(t))
}

func TestRedisBackend_Set(t *testing.T) {
	testKVDBBackend_Set(t, openTestRedisKVDB(t))
}

func testKVDBBackend_Set(t *testing.T, kvdb KVDBEngine) {
	val, err := kvdb.Get("__key_not_exists__")
	if err != nil || val != "" {
		t.Fatal(err)
	}

	for i := 0; i < 10000; i++ {
		key := strconv.Itoa(rand.Intn(10000))
		val := strconv.Itoa(rand.Intn(10000))
		err = kvdb.Put(key, val)
		if err != nil {
			t.Fatal(err)
		}
		var verifyVal string
		verifyVal, err = kvdb.Get(key)
		if err != nil {
			t.Fatal(err)
		}

		if verifyVal != val {
			t.Errorf("%s != %s", val, verifyVal)
		}
	}

}

func TestMongoBackend_Find(t *testing.T) {
	testBackend_Find(t, openTestMongoKVDB(t))
}

func TestRedisBackend_Find(t *testing.T) {
	testBackend_Find(t, openTestRedisKVDB(t))
}

func testBackend_Find(t *testing.T, kvdb KVDBEngine) {
	beginKey := strconv.Itoa(1000 + rand.Intn(2000-1000))
	if len(beginKey) != 4 {
		t.Fatal("wrong begin key: %s", beginKey)
	}

	endKey := strconv.Itoa(5000 + rand.Intn(5000))

	if len(endKey) != 4 {
		t.Fatal("wrong end key: %s", endKey)
	}
	kvdb.Put(beginKey, beginKey)
	kvdb.Put(endKey, endKey)

	it := kvdb.Find(beginKey, endKey)
	oldKey := ""
	beginKeyFound, endKeyFound := false, false
	//println("testBackend_Find", beginKey, endKey)
	for {
		item, err := it.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			t.Error(err)
			break
		}

		if item.Key <= oldKey { // the keys should be increasing
			t.Errorf("old key is %s, new key is %s, should be increasing", oldKey, item.Key)
		}

		//println("visit", item.Key)
		if item.Key == beginKey {
			beginKeyFound = true
		} else if item.Key == endKey {
			endKeyFound = true
		}

		//println(item.Key, item.Val)
		oldKey = item.Key
	}
	if !beginKeyFound {
		t.Errorf("begin key is not found")
	}
	if endKeyFound {
		t.Errorf("end key is found")
	}
}

func BenchmarkMongoBackend_GetSet(b *testing.B) {
	benchmarkBackend_GetSet(b, openTestMongoKVDB(b))
}

func BenchmarkRedisBackend_GetSet(b *testing.B) {
	benchmarkBackend_GetSet(b, openTestRedisKVDB(b))
}

func benchmarkBackend_GetSet(b *testing.B, kvdb KVDBEngine) {
	key := "testkey"

	for i := 0; i < b.N; i++ {
		val := strconv.Itoa(rand.Intn(1000))
		kvdb.Put(key, val)
		getval, err := kvdb.Get(key)
		if err != nil {
			b.Error(err)
		}

		if getval != val {
			b.Errorf("put %s but get %s", val, getval)
		}
	}
}

func BenchmarkMongoBackend_Find(b *testing.B) {
	benchmarkBackend_Find(b, openTestMongoKVDB(b))
}

func BenchmarkRedisBackend_Find(b *testing.B) {
	benchmarkBackend_Find(b, openTestRedisKVDB(b))
}

func benchmarkBackend_Find(b *testing.B, kvdb KVDBEngine) {
	var keys []string
	for i := 1; i <= 10; i++ {
		keys = append(keys, fmt.Sprintf("%03d", i))
	}
	for _, key := range keys {
		kvdb.Put(key, key)
	}

	//fmt.Printf("keys %v\n", keys)
	beginKey, endKey := keys[0], keys[len(keys)-1]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		it := kvdb.Find(beginKey, endKey)
		for {
			_, err := it.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				b.Error(err)
			}
			//println(item.Key, item.Val)
		}
	}
}

type _Fataler interface {
	Fatal(args ...interface{})
}

func openTestMongoKVDB(f _Fataler) KVDBEngine {
	kvdb, err := kvdb_mongo.OpenMongoKVDB("mongodb://127.0.0.1:27017/goworld", "goworld", "__kv__")
	if err != nil {
		f.Fatal(err)
	}
	return kvdb
}

func openTestRedisKVDB(f _Fataler) KVDBEngine {
	kvdb, err := kvdb_redis.OpenRedisKVDB("127.0.0.1:6379")
	if err != nil {
		f.Fatal(err)
	}
	return kvdb
}
