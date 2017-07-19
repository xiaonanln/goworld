package entity_storage_redis

import (
	"github.com/garyburd/redigo/redis"
	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/netutil"
)

var (
	dataPacker = netutil.MessagePackMsgPacker{}
)

type redisEntityStorage struct {
	c redis.Conn
}

func OpenRedis(host string, dbindex int) (*redisEntityStorage, error) {
	c, err := redis.Dial("tcp", host)
	if err != nil {
		return nil, errors.Wrap(err, "redis dail failed")
	}

	if _, err := c.Do("SELECT", dbindex); err != nil {
		return nil, errors.Wrap(err, "redis select db failed")
	}

	es := &redisEntityStorage{
		c: c,
	}

	return es, nil
}

func entityKey(typeName string, eid common.EntityID) string {
	return typeName + "$" + string(eid)
}

func packData(data interface{}) (b []byte, err error) {
	b, err = dataPacker.PackMsg(data, b)
	return
}

func (es *redisEntityStorage) List(typeName string) ([]common.EntityID, error) {
	r, err := redis.Values(es.c.Do("SCAN", "0", "MATCH", typeName+"$*", "COUNT", 10000))
	if err != nil {
		return nil, err
	}
	var eids []common.EntityID
	prefixLen := len(typeName) + 1
	for {
		nextCursor := r[0]
		keys, err := redis.Strings(r[1], nil)
		if err != nil {
			return nil, err
		}

		for _, key := range keys {
			eids = append(eids, common.EntityID(key[prefixLen:]))
		}

		if isZeroCursor(nextCursor) {
			break
		}
		r, err = redis.Values(es.c.Do("SCAN", nextCursor))
	}
	return eids, nil
}

func isZeroCursor(c interface{}) bool {
	return string(c.([]byte)) == "0"
}

func (es *redisEntityStorage) Write(typeName string, entityID common.EntityID, data interface{}) error {
	b, err := packData(data)
	if err != nil {
		return err
	}

	_, err = es.c.Do("SET", entityKey(typeName, entityID), b)
	return err
}

func (es *redisEntityStorage) Read(typeName string, entityID common.EntityID) (interface{}, error) {
	b, err := redis.Bytes(es.c.Do("GET", entityKey(typeName, entityID)))
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	if err = dataPacker.UnpackMsg(b, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func (es *redisEntityStorage) Exists(typeName string, entityID common.EntityID) (bool, error) {
	key := entityKey(typeName, entityID)
	exists, err := redis.Bool(es.c.Do("EXISTS", key))
	return exists, err
}
