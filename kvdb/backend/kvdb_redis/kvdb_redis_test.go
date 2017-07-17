package kvdb_redis

import (
	"math/rand"
	"strconv"
	"testing"

	"io"
)

func TestRedisKVDB_Set(t *testing.T) {
	kvdb, err := OpenRedisKVDB("127.0.0.1:6379")
	if err != nil {
		t.Fatal(err)
	}
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

func TestRedisKVDB_Find(t *testing.T) {
	kvdb, err := OpenRedisKVDB("127.0.0.1:6379")
	if err != nil {
		t.Fatal(err)
	}

	beginKey := "1000"
	endKey := "9999"
	it := kvdb.Find(beginKey)
	oldKey := ""
	for {
		item, err := it.Next()
		if item.Key >= endKey {
			break
		}

		if err != nil {
			if err != io.EOF {
				t.Error(err)
			}

			break
		}

		if item.Key <= oldKey { // the keys should be increasing
			t.Errorf("old key is %s, new key is %s, should be increasing", oldKey, item.Key)
		}

		println(item.Key, item.Val)
		oldKey = item.Key
	}
}
