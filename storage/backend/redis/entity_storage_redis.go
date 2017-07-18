package entity_storage_redis

import (
	"github.com/garyburd/redigo/redis"
	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/common"
)

type redisEntityStorage struct {
	c redis.Conn
}

func OpenRedis(host string, dbindex int) (*redisEntityStorage, error) {
	c, err := redis.Dial("tcp", host)
	if err != nil {
		return nil, errors.Wrap(err, "redis dail failed")
	}

	es := &redisEntityStorage{
		c: c,
	}

	return es, nil
}

func (es *redisEntityStorage) List(typeName string) ([]common.EntityID, error) {
	return nil, nil
}

func (es *redisEntityStorage) Write(typeName string, entityID common.EntityID, data interface{}) error {
	return nil
}

func (es *redisEntityStorage) Read(typeName string, entityID common.EntityID) (interface{}, error) {
	return nil, nil
}

func (es *redisEntityStorage) Exists(typeName string, entityID common.EntityID) (bool, error) {
	return false, nil
}
