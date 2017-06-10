package string_storage_mongodb

import (
	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/storage"
)

const (
	DEFAULT_DB_NAME = "goworld"
)

var (
	db *mgo.Database
)

type MongoDBEntityStorge struct {
	db *mgo.Database
}

func OpenMongoDB(url string, dbname string) (storage.EntityStorage, error) {
	gwlog.Debug("Connecting MongoDB ...")
	session, err := mgo.Dial(url)
	if err != nil {
		panic(err)
	}

	session.SetMode(mgo.Monotonic, true)
	if dbname == "" {
		// if db is not specified, use default
		dbname = DEFAULT_DB_NAME
	}
	db = session.DB(dbname)
	return &MongoDBEntityStorge{
		db: db,
	}, nil
}

func collectionName(name string) string {
	return fmt.Sprintf("S_%s", name)
}

func (ss *MongoDBEntityStorge) Write(name string, entityID common.EntityID, data interface{}) error {
	col := ss.db.C(collectionName(name))
	_, err := col.UpsertId(entityID, bson.M{
		"data": data,
	})
	return err
}

func (ss *MongoDBEntityStorge) Read(name string, entityID common.EntityID) (interface{}, error) {
	col := ss.db.C(collectionName(name))
	q := col.FindId(entityID)
	var doc bson.M
	err := q.One(&doc)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}(doc["data"].(bson.M)), nil
}
