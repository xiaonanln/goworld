package kvdb_redis

import (
	"github.com/garyburd/redigo/redis"
	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/kvdb/types"
)

const (
	keyPrefix = "_KV_"
)

type redisKVDB struct {
	c redis.Conn
}

func OpenRedisKVDB(host string) (*redisKVDB, error) {
	c, err := redis.Dial("tcp", host)
	if err != nil {
		return nil, errors.Wrap(err, "redis dail failed")
	}

	db := &redisKVDB{
		c: c,
	}
	return db, nil
}

func (db *redisKVDB) Get(key string) (val string, err error) {
	r, err := db.c.Do("GET", keyPrefix+key)
	if err != nil {
		return "", err
	}
	return r.(string), err
}

func (db *redisKVDB) Put(key string, val string) error {
	_, err := db.c.Do("PUT", keyPrefix+key, val)
	return err
}
func (db *redisKVDB) Find(key string) kvdb_types.Iterator {
	db.c.Do("GETRANGE", keyPrefix+key)
	return nil
}
