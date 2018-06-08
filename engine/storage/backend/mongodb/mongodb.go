package entitystoragemongodb

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"io"

	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/storage/storage_common"
)

const (
	_DEFAULT_DB_NAME = "goworld"
)

var (
	db *mgo.Database
)

type mongoDBEntityStorge struct {
	db *mgo.Database
}

// OpenMongoDB opens mongodb as entity storage
func OpenMongoDB(url string, dbname string) (storagecommon.EntityStorage, error) {
	gwlog.Debugf("Connecting MongoDB ...")
	session, err := mgo.Dial(url)
	if err != nil {
		return nil, err
	}

	session.SetMode(mgo.Monotonic, true)
	if dbname == "" {
		// if db is not specified, use default
		dbname = _DEFAULT_DB_NAME
	}
	db = session.DB(dbname)
	return &mongoDBEntityStorge{
		db: db,
	}, nil
}

func (es *mongoDBEntityStorge) Write(typeName string, entityID common.EntityID, data interface{}) error {
	col := es.getCollection(typeName)
	_, err := col.UpsertId(entityID, bson.M{
		"data": data,
	})
	return err
}

func (es *mongoDBEntityStorge) Read(typeName string, entityID common.EntityID) (interface{}, error) {
	col := es.getCollection(typeName)
	q := col.FindId(entityID)
	var doc bson.M
	err := q.One(&doc)
	if err != nil {
		return nil, err
	}
	return es.convertM2Map(doc["data"].(bson.M)), nil
}

func (es *mongoDBEntityStorge) convertM2Map(m bson.M) map[string]interface{} {
	ma := map[string]interface{}(m)
	es.convertM2MapInMap(ma)
	return ma
}

func (es *mongoDBEntityStorge) convertM2MapInMap(m map[string]interface{}) {
	for k, v := range m {
		switch im := v.(type) {
		case bson.M:
			m[k] = es.convertM2Map(im)
		case map[string]interface{}:
			es.convertM2MapInMap(im)
		case []interface{}:
			es.convertM2MapInList(im)
		}
	}
}

func (es *mongoDBEntityStorge) convertM2MapInList(l []interface{}) {
	for i, v := range l {
		switch im := v.(type) {
		case bson.M:
			l[i] = es.convertM2Map(im)
		case map[string]interface{}:
			es.convertM2MapInMap(im)
		case []interface{}:
			es.convertM2MapInList(im)
		}
	}
}

func (es *mongoDBEntityStorge) getCollection(typeName string) *mgo.Collection {
	return es.db.C(typeName)
}

func (es *mongoDBEntityStorge) List(typeName string) ([]common.EntityID, error) {
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

func (es *mongoDBEntityStorge) Exists(typeName string, entityID common.EntityID) (bool, error) {
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

func (es *mongoDBEntityStorge) Close() {
	es.db.Session.Close()
}

func (es *mongoDBEntityStorge) IsEOF(err error) bool {
	return err == io.EOF || err == io.ErrUnexpectedEOF
}
