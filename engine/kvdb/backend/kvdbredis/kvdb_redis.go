package kvdbredis

import (
	"io"

	"github.com/garyburd/redigo/redis"
	"github.com/google/btree"
	"github.com/pkg/errors"
	//"github.com/xiaonanln/goworld/gwlog"
	. "github.com/xiaonanln/goworld/engine/kvdb/types"
)

const (
	keyPrefix = "_KV_"
)

type redisKVDB struct {
	c       redis.Conn
	keyTree *btree.BTree
}

type keyTreeItem struct {
	key string
}

func (ki keyTreeItem) Less(_other btree.Item) bool {
	return ki.key < _other.(keyTreeItem).key
}

// Open Redis for KVDB backend
func OpenRedisKVDB(host string, dbindex int) (KVDBEngine, error) {
	c, err := redis.Dial("tcp", host)
	if err != nil {
		return nil, errors.Wrap(err, "redis dail failed")
	}

	db := &redisKVDB{
		c:       c,
		keyTree: btree.New(2),
	}
	if err := db.initialize(dbindex); err != nil {
		panic(errors.Wrap(err, "redis kvdb initialize failed"))
	}

	return db, nil
}

func (db *redisKVDB) initialize(dbindex int) error {
	if _, err := db.c.Do("SELECT", dbindex); err != nil {
		return err
	}

	keyMatch := keyPrefix + "*"
	r, err := redis.Values(db.c.Do("SCAN", "0", "MATCH", keyMatch, "COUNT", 10000))
	if err != nil {
		return err
	}
	for {
		nextCursor := r[0]
		keys, err := redis.Strings(r[1], nil)
		if err != nil {
			return err
		}
		//gwlog.Info("SCAN: %v, nextcursor=%s", keys, string(nextCursor.([]byte)))
		for _, key := range keys {
			key := key[len(keyPrefix):]
			db.keyTree.ReplaceOrInsert(keyTreeItem{key})
		}

		if db.isZeroCursor(nextCursor) {
			break
		}
		r, err = redis.Values(db.c.Do("SCAN", nextCursor, "MATCH", keyMatch, "COUNT", 10000))
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *redisKVDB) isZeroCursor(c interface{}) bool {
	return string(c.([]byte)) == "0"
}

func (db *redisKVDB) Get(key string) (val string, err error) {
	r, err := db.c.Do("GET", keyPrefix+key)
	if err != nil {
		return "", err
	}
	if r == nil {
		return "", nil
	}
	return string(r.([]byte)), err
}

func (db *redisKVDB) Put(key string, val string) error {
	_, err := db.c.Do("SET", keyPrefix+key, val)
	if err != nil {
		db.keyTree.ReplaceOrInsert(keyTreeItem{key})
	}
	return err
}

type redisKVDBIterator struct {
	db       *redisKVDB
	leftKeys []string
}

func (it *redisKVDBIterator) Next() (KVItem, error) {
	if len(it.leftKeys) == 0 {
		return KVItem{}, io.EOF
	}

	key := it.leftKeys[0]
	it.leftKeys = it.leftKeys[1:]
	val, err := it.db.Get(key)
	if err != nil {
		return KVItem{}, err
	}

	return KVItem{key, val}, nil
}

func (db *redisKVDB) Find(beginKey string, endKey string) Iterator {
	keys := []string{} // retrive all keys in the range, ordered
	db.keyTree.AscendRange(keyTreeItem{beginKey}, keyTreeItem{endKey}, func(it btree.Item) bool {
		keys = append(keys, it.(keyTreeItem).key)
		return true
	})

	return &redisKVDBIterator{
		db:       db,
		leftKeys: keys,
	}
}

func (db *redisKVDB) Close() {
	db.c.Close()
}

func (db *redisKVDB) IsEOF(err error) bool {
	return err == io.EOF || err == io.ErrUnexpectedEOF
}
