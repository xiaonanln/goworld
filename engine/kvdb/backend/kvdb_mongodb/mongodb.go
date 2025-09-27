package kvdbmongo

import (
	"context"

	"io"

	"github.com/xiaonanln/goworld/engine/gwlog"
	kvdbtypes "github.com/xiaonanln/goworld/engine/kvdb/types"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	_DEFAULT_DB_NAME = "goworld"
	_VAL_KEY         = "_"
)

type mongoKVDB struct {
	s *mongo.Client
	c *mongo.Collection
}

// OpenMongoKVDB opens mongodb as KVDB engine
func OpenMongoKVDB(url string, dbname string, collectionName string) (kvdbtypes.KVDBEngine, error) {
	gwlog.Debugf("Connecting MongoDB ...")
	session, err := mongo.Connect(options.Client().ApplyURI(url))
	if err != nil {
		return nil, err
	}

	if dbname == "" {
		// if db is not specified, use default
		dbname = _DEFAULT_DB_NAME
	}
	db := session.Database(dbname)
	c := db.Collection(collectionName)
	return &mongoKVDB{
		s: session,
		c: c,
	}, nil
}

func (kvdb *mongoKVDB) Put(key string, val string) error {
	opts := options.UpdateOne().SetUpsert(true)
	_, err := kvdb.c.UpdateOne(context.TODO(), bson.M{"_id": key}, bson.M{"$set": bson.M{
		_VAL_KEY: val,
	}}, opts)
	return err
}

func (kvdb *mongoKVDB) Get(key string) (val string, err error) {
	var doc map[string]string
	err = kvdb.c.FindOne(context.TODO(), bson.M{"_id": key}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err = nil
		}
		return
	}
	val = doc[_VAL_KEY]
	return
}

type mongoKVIterator struct {
	it *mongo.Cursor
}

func (it *mongoKVIterator) Next() (kvdbtypes.KVItem, error) {
	ok := it.it.Next(context.Background())
	if ok {
		var doc map[string]string
		err := it.it.Decode(&doc)
		if err != nil {
			return kvdbtypes.KVItem{}, err
		}
		return kvdbtypes.KVItem{
			Key: doc["_id"],
			Val: doc["_"],
		}, nil
	}

	err := it.it.Close(context.Background())
	if err != nil {
		return kvdbtypes.KVItem{}, err
	}
	return kvdbtypes.KVItem{}, io.EOF
}

func (kvdb *mongoKVDB) Find(beginKey string, endKey string) (kvdbtypes.Iterator, error) {
	cur, err := kvdb.c.Find(context.Background(), bson.M{"_id": bson.M{"$gte": beginKey, "$lt": endKey}})
	if err != nil {
		return nil, err
	}
	return &mongoKVIterator{
		it: cur,
	}, nil
}

func (kvdb *mongoKVDB) Close() {
	kvdb.s.Disconnect(context.Background())
}

func (kvdb *mongoKVDB) IsConnectionError(err error) bool {
	return err == io.EOF || err == io.ErrUnexpectedEOF
}
