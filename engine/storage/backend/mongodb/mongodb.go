package entity_storage_mongodb

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"io"

	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/gwlog"

	. "github.com/xiaonanln/goworld/engine/storage/storage_common"
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

func OpenMongoDB(url string, dbname string) (EntityStorage, error) {
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
	db = session.DB(dbname)
	return &MongoDBEntityStorge{
		db: db,
	}, nil
}

func (es *MongoDBEntityStorge) Write(typeName string, entityID common.EntityID, data interface{}) error {
	col := es.getCollection(typeName)
	_, err := col.UpsertId(entityID, bson.M{
		"data": data,
	})
	return err
}

func (es *MongoDBEntityStorge) Read(typeName string, entityID common.EntityID) (interface{}, error) {
	col := es.getCollection(typeName)
	q := col.FindId(entityID)
	var doc bson.M
	err := q.One(&doc)
	if err != nil {
		return nil, err
	}
	return es.convertM2Map(doc["data"].(bson.M)), nil
}

func (es *MongoDBEntityStorge) convertM2Map(m bson.M) map[string]interface{} {
	ma := map[string]interface{}(m)
	for k, v := range ma {
		if m, ok := v.(bson.M); ok {
			ma[k] = es.convertM2Map(m)
		}
	}
	return ma
}

func (es *MongoDBEntityStorge) getCollection(typeName string) *mgo.Collection {
	return es.db.C(typeName)
}

func (es *MongoDBEntityStorge) List(typeName string) ([]common.EntityID, error) {
	col := es.getCollection(typeName)
	var docs []bson.M
	err := col.Find(nil).Select(bson.M{"_id": 1}).All(&docs)
	if err != nil {
		return nil, err
	}

	entityIDs := make([]common.EntityID, len(docs))
	for i, doc := range docs {
		entityIDs[i] = common.EntityID(doc["_id"].(string))
	}
	return entityIDs, nil
}

func (es *MongoDBEntityStorge) Exists(typeName string, entityID common.EntityID) (bool, error) {
	col := es.getCollection(typeName)
	query := col.FindId(entityID)
	var doc bson.M
	err := query.One(&doc)
	if err == nil {
		// doc found
		return true, nil
	} else if err == mgo.ErrNotFound {
		return false, nil
	} else {
		return false, err
	}
}

func (es *MongoDBEntityStorge) Close() {
	es.db.Session.Close()
}

func (es *MongoDBEntityStorge) IsEOF(err error) bool {
	return err == io.EOF || err == io.ErrUnexpectedEOF
}
