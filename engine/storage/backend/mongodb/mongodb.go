package entitystoragemongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"io"

	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/gwlog"
	storagecommon "github.com/xiaonanln/goworld/engine/storage/storage_common"
)

const (
	_DEFAULT_DB_NAME = "goworld"
)

var (
	db *mongo.Database
)

type mongoDBEntityStorge struct {
	db *mongo.Database
}

// OpenMongoDB opens mongodb as entity storage
func OpenMongoDB(url string, dbname string) (storagecommon.EntityStorage, error) {
	gwlog.Debugf("Connecting MongoDB ...")
	session, err := mongo.Connect(options.Client().ApplyURI(url))
	if err != nil {
		return nil, err
	}

	if dbname == "" {
		// if db is not specified, use default
		dbname = _DEFAULT_DB_NAME
	}
	db = session.Database(dbname)
	return &mongoDBEntityStorge{
		db: db,
	}, nil
}

func (es *mongoDBEntityStorge) Write(typeName string, entityID common.EntityID, data interface{}) error {
	col := es.getCollection(typeName)
	opts := options.UpdateOne().SetUpsert(true)
	_, err := col.UpdateOne(context.TODO(), bson.M{"_id": entityID}, bson.M{"$set": bson.M{
		"data": data,
	}}, opts)
	return err
}

func (es *mongoDBEntityStorge) Read(typeName string, entityID common.EntityID) (interface{}, error) {
	col := es.getCollection(typeName)
	opts_find := options.FindOne().SetProjection(bson.M{"data": 1})
	var doc bson.M
	err := col.FindOne(context.TODO(), bson.M{"_id": entityID}, opts_find).Decode(&doc)
	if err != nil {
		return nil, err
	}
	if m, ok := doc["data"].(bson.M); ok {
		return es.convertM2Map(m), nil
	} else if m, ok := doc["data"].(bson.D); ok {
		ret := make(map[string]interface{}, len(m))
		for _, v := range m {
			ret[v.Key] = v.Value
		}
		return ret, nil
	}
	return nil, err
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

func (es *mongoDBEntityStorge) getCollection(typeName string) *mongo.Collection {
	return es.db.Collection(typeName)
}

func (es *mongoDBEntityStorge) List(typeName string) ([]common.EntityID, error) {
	col := es.getCollection(typeName)
	opts_finds := options.Find().SetProjection(bson.M{"_id": 1})
	cur, err := col.Find(context.Background(), bson.M{"_id": 1}, opts_finds)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return []common.EntityID{}, nil
		}
		return nil, err
	}
	defer cur.Close(context.Background())

	entityIDs := make([]common.EntityID, 0)
	for cur.Next(context.Background()) {
		if v, ok := cur.Current.Lookup("_id").StringValueOK(); ok {
			entityIDs = append(entityIDs, common.EntityID(v))
		}
	}
	return entityIDs, nil
}

func (es *mongoDBEntityStorge) Exists(typeName string, entityID common.EntityID) (bool, error) {
	col := es.getCollection(typeName)
	query := col.FindOne(context.Background(), bson.D{{"_id", entityID}})
	if err := query.Err(); err != nil {
		return false, err
	}
	return true, nil
}

func (es *mongoDBEntityStorge) Close() {
	es.db.Client().Disconnect(context.Background())
}

func (es *mongoDBEntityStorge) IsEOF(err error) bool {
	return err == io.EOF || err == io.ErrUnexpectedEOF
}
