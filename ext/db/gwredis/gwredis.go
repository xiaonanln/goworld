package gwredis

import (
	"github.com/garyburd/redigo/redis"
	"github.com/xiaonanln/goworld/engine/async"
)

const (
	_REDIS_ASYNC_JOB_GROUP = "_redis"
)

type DB struct {
	conn redis.Conn
}

func Dial(network, address string, options []redis.DialOption, ac async.AsyncCallback) {
	async.AppendAsyncJob(_REDIS_ASYNC_JOB_GROUP, func() (res interface{}, err error) {
		conn, err := redis.Dial(network, address, options...)
		if err == nil {
			return &DB{conn}, nil
		} else {
			return nil, err
		}
	}, ac)
}

func DialURL(rawurl string, options []redis.DialOption, ac async.AsyncCallback) {
	async.AppendAsyncJob(_REDIS_ASYNC_JOB_GROUP, func() (res interface{}, err error) {
		conn, err := redis.DialURL(rawurl, options...)
		if err == nil {
			return &DB{conn}, nil
		} else {
			return nil, err
		}
	}, ac)
}

func (db *DB) Do(commandName string, args []interface{}, ac async.AsyncCallback) {
	async.AppendAsyncJob(_REDIS_ASYNC_JOB_GROUP, func() (res interface{}, err error) {
		res, err = db.conn.Do(commandName, args...)
		return
	}, ac)
}
