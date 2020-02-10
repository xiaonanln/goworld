package kvdb

import (
	"math/rand"
	"strconv"
	"testing"

	"fmt"
	"io"

	"github.com/xiaonanln/goworld/engine/kvdb/backend/kvdb_mongodb"
	"github.com/xiaonanln/goworld/engine/kvdb/backend/kvdbredis"
	. "github.com/xiaonanln/goworld/engine/kvdb/types"
)

func TestMongoBackendSet(t *testing.T) {
	testKVDBBackendSet(t, openTestMongoKVDB(t))
}

func TestRedisBackendSet(t *testing.T) {
	testKVDBBackendSet(t, openTestRedisKVDB(t))
}

func testKVDBBackendSet(t *testing.T, kvdb KVDBEngine) {
	val, err := kvdb.Get("__key_not_exists__")
	if err != nil || val != "" {
		t.Fatal(err)
	}

	for i := 0; i < 100; i++ {
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

func TestMongoBackendFind(t *testing.T) {
	testBackendFind(t, openTestMongoKVDB(t))
}

//func TestRedisBackendFind(t *testing.T) {
//	testBackendFind(t, openTestRedisKVDB(t))
//}

func testBackendFind(t *testing.T, kvdb KVDBEngine) {
	beginKey := strconv.Itoa(1000 + rand.Intn(2000-1000))
	if len(beginKey) != 4 {
		t.Fatalf("wrong begin key: %s", beginKey)
	}

	endKey := strconv.Itoa(5000 + rand.Intn(5000))

	if len(endKey) != 4 {
		t.Fatalf("wrong end key: %s", endKey)
	}
	if err := kvdb.Put(beginKey, beginKey); err != nil {
		t.Error(err)
	}
	if err := kvdb.Put(endKey, endKey); err != nil {
		t.Error(err)
	}

	it, err := kvdb.Find(beginKey, endKey)
	if err != nil {
		t.Error(err)
		return
	}

	oldKey := ""
	beginKeyFound, endKeyFound := false, false
	//println("testBackendFind", beginKey, endKey)
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

func BenchmarkMongoBackendGetSet(b *testing.B) {
	benchmarkBackendGetSet(b, openTestMongoKVDB(b))
}

func BenchmarkRedisBackendGetSet(b *testing.B) {
	benchmarkBackendGetSet(b, openTestRedisKVDB(b))
}

func benchmarkBackendGetSet(b *testing.B, kvdb KVDBEngine) {
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

func BenchmarkMongoBackendFind(b *testing.B) {
	benchmarkBackendFind(b, openTestMongoKVDB(b))
}

func BenchmarkRedisBackendFind(b *testing.B) {
	benchmarkBackendFind(b, openTestRedisKVDB(b))
}

func benchmarkBackendFind(b *testing.B, kvdb KVDBEngine) {
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
		it, err := kvdb.Find(beginKey, endKey)
		if err != nil {
			b.Error(err)
			continue
		}

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
	kvdb, err := kvdbmongo.OpenMongoKVDB("mongodb://127.0.0.1:27017/goworld", "goworld", "__kv__")
	if err != nil {
		f.Fatal(err)
	}
	return kvdb
}

func openTestRedisKVDB(f _Fataler) KVDBEngine {
	kvdb, err := kvdbredis.OpenRedisKVDB("redis://127.0.0.1:6379", 0)
	if err != nil {
		f.Fatal(err)
	}
	return kvdb
}
