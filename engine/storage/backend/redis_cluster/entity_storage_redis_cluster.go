package entitystoragerediscluster

import (
	"io"

	"time"

	rediscluster "github.com/chasex/redis-go-cluster"
	"github.com/garyburd/redigo/redis"
	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/netutil"
	"github.com/xiaonanln/goworld/engine/storage/storage_common"
)

var (
	dataPacker = netutil.MessagePackMsgPacker{}
)

type redisClusterEntityStorage struct {
	c *rediscluster.Cluster
}

// OpenRedis opens redis as entity storage
func OpenRedisCluster(startNodes []string) (storagecommon.EntityStorage, error) {
	c, err := rediscluster.NewCluster(&rediscluster.Options{
		StartNodes:   startNodes,
		ConnTimeout:  10 * time.Second, // Connection timeout
		ReadTimeout:  60 * time.Second, // Read timeout
		WriteTimeout: 60 * time.Second, // Write timeout
		KeepAlive:    1,                // Maximum keep alive connecion in each node
		AliveTime:    10 * time.Minute, // Keep alive timeout
	})

	if err != nil {
		return nil, errors.Wrap(err, "connect redis cluster failed")
	}

	es := &redisClusterEntityStorage{
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

func (es *redisClusterEntityStorage) List(typeName string) ([]common.EntityID, error) {
	keyMatch := typeName + "$*"
	r, err := redis.Values(es.c.Do("SCAN", "0", "MATCH", keyMatch, "COUNT", 10000))
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
			println("key", key)
			eids = append(eids, common.EntityID(key[prefixLen:]))
		}

		if isZeroCursor(nextCursor) {
			break
		}
		r, err = redis.Values(es.c.Do("SCAN", nextCursor, "MATCH", keyMatch, "COUNT", 10000))
		if err != nil {
			return nil, err
		}
	}
	return eids, nil
}

func isZeroCursor(c interface{}) bool {
	return string(c.([]byte)) == "0"
}

func (es *redisClusterEntityStorage) Write(typeName string, entityID common.EntityID, data interface{}) error {
	b, err := packData(data)
	if err != nil {
		return err
	}

	_, err = es.c.Do("SET", entityKey(typeName, entityID), b)
	return err
}

func (es *redisClusterEntityStorage) Read(typeName string, entityID common.EntityID) (interface{}, error) {
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

func (es *redisClusterEntityStorage) Exists(typeName string, entityID common.EntityID) (bool, error) {
	key := entityKey(typeName, entityID)
	exists, err := redis.Bool(es.c.Do("EXISTS", key))
	return exists, err
}

func (es *redisClusterEntityStorage) Close() {
	es.c.Close()
}

func (es *redisClusterEntityStorage) IsEOF(err error) bool {
	return err == io.EOF || err == io.ErrUnexpectedEOF
}
