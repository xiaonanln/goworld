package kvdb_mongo

import (
	"math/rand"
	"strconv"
	"testing"
)

func TestMongoKVDB_Set(t *testing.T) {
	kvdb, err := OpenMongoKVDB("mongodb://127.0.0.1:27017/goworld", "goworld", "__kv__")
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
