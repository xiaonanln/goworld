package kvdbredis

import (
	"io"

	"github.com/garyburd/redigo/redis"
	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/engine/kvdb/types"
)

const (
	keyPrefix = "_KV_"
)

type redisKVDB struct {
	c redis.Conn
}

// OpenRedisKVDB opens Redis for KVDB backend
func OpenRedisKVDB(url string, dbindex int) (kvdbtypes.KVDBEngine, error) {
	c, err := redis.DialURL(url)
	if err != nil {
		return nil, errors.Wrap(err, "redis dail failed")
	}

	db := &redisKVDB{
		c: c,
	}

	if err := db.initialize(dbindex); err != nil {
		panic(errors.Wrap(err, "redis kvdb initialize failed"))
	}

	return db, nil
}

func (db *redisKVDB) initialize(dbindex int) error {
	if dbindex >= 0 {
		if _, err := db.c.Do("SELECT", dbindex); err != nil {
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
	return err
}

type redisKVDBIterator struct {
	db       *redisKVDB
	leftKeys []string
}

func (it *redisKVDBIterator) Next() (kvdbtypes.KVItem, error) {
	if len(it.leftKeys) == 0 {
		return kvdbtypes.KVItem{}, io.EOF
	}

	key := it.leftKeys[0]
	it.leftKeys = it.leftKeys[1:]
	val, err := it.db.Get(key)
	if err != nil {
		return kvdbtypes.KVItem{}, err
	}

	return kvdbtypes.KVItem{key, val}, nil
}

func (db *redisKVDB) Find(beginKey string, endKey string) (kvdbtypes.Iterator, error) {
	return nil, errors.Errorf("operation not supported on redis")
}

func (db *redisKVDB) Close() {
	db.c.Close()
}

func (db *redisKVDB) IsConnectionError(err error) bool {
	return err == io.EOF || err == io.ErrUnexpectedEOF
}
