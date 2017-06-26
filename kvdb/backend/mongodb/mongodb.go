package kvdb_mongo

import (
	"gopkg.in/mgo.v2"

	"github.com/xiaonanln/goworld/gwlog"
)

const (
	DEFAULT_DB_NAME = "goworld"
	VAL_KEY         = "_"
)

type MongoKVDB struct {
	c *mgo.Collection
}

func OpenMongoKVDB(url string, dbname string, collectionName string) (*MongoKVDB, error) {
	gwlog.Debug("Connecting MongoDB ...")
	session, err := mgo.Dial(url)
	if err != nil {
		return nil, err
	}

	session.SetMode(mgo.Monotonic, true)
	if dbname == "" {
		// if db is not specified, use default
		dbname = DEFAULT_DB_NAME
	}
	db := session.DB(dbname)
	c := db.C(collectionName)
	return &MongoKVDB{
		c: c,
	}, nil
}

func (kvdb *MongoKVDB) Put(key string, val string) error {
	_, err := kvdb.c.UpsertId(key, map[string]string{
		VAL_KEY: val,
	})
	return err
}

func (kvdb *MongoKVDB) Get(key string) (val string, err error) {
	q := kvdb.c.FindId(key)
	var doc map[string]string
	err = q.One(&doc)
	if err != nil {
		if err == mgo.ErrNotFound {
			err = nil
		}
		return
	}
	val = doc[VAL_KEY]
	return
}
