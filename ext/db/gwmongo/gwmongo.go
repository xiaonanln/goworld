package gwmongo

import (
	"context"

	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/engine/async"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	_MONGODB_ASYNC_JOB_GROUP = "_mongodb"
)

var (
	errNoSession = errors.Errorf("no session, please dail")
)

// MongoDB is a MongoDB instance can be used to manipulate Mongo DBs
type DB struct {
	session *mongo.Client
	db      *mongo.Database
}

func (mdb *DB) checkConnected() bool {
	return mdb.session != nil && mdb.db != nil
}

// Dial connects the a MongoDB
// returns *MongoDB
func Dial(url string, dbname string, ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		gwlog.Infof("Dailing MongoDB: %s ...", url)
		session, err := mongo.Connect(options.Client().ApplyURI(url))
		if err != nil {
			return nil, err
		}
		db := session.Database(dbname)
		return &DB{session, db}, nil
	}, ac)
}

// Close closes MongoDB
func (mdb *DB) Close(ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		mdb.session.Disconnect(context.Background())
		mdb.session = nil
		mdb.db = nil
		return nil, nil
	}, ac)
}

// UseDB uses the specified DB
func (mdb *DB) UseDB(dbname string, ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		mdb.db = mdb.session.Database(dbname)
		return nil, nil
	}, ac)
}

// // SetMode sets the consistency mode
// func (mdb *DB) SetMode(consistency mgo.Mode, ac async.AsyncCallback) {
// 	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
// 		if !mdb.checkConnected() {
// 			return nil, errNoSession
// 		}
// 		return nil, nil
// 	}, ac)
// }

// FindId finds document in collection by Id
func (mdb *DB) FindId(collectionName string, id interface{}, setupQuery func(query *mongo.SingleResult), ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}
		q := mdb.db.Collection(collectionName).FindOne(context.Background(), bson.M{"_id": id})
		if setupQuery != nil {
			setupQuery(q)
		}
		var res bson.M
		err := q.Decode(&res)
		return res, err
	}, ac)
}

// FindOne finds one document with specified query
func (mdb *DB) FindOne(collectionName string, query bson.M, setupQuery func(query *options.FindOneOptionsBuilder), ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		opts := options.FindOne()
		if setupQuery != nil {
			setupQuery(opts)
		}
		q := mdb.db.Collection(collectionName).FindOne(context.Background(), query, opts)
		var res bson.M
		err := q.Decode(&res)
		return res, err
	}, ac)
}

// FindAll finds all documents with specified query
func (mdb *DB) FindAll(collectionName string, query bson.M, setupQuery func(query *options.FindOptionsBuilder), ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}
		opts := options.Find()
		if setupQuery != nil {
			setupQuery(opts)
		}
		cur, err := mdb.db.Collection(collectionName).Find(context.Background(), query, opts)
		var res []bson.M
		err = cur.All(context.Background(), &res)
		return res, err
	}, ac)
}

// Count counts the number of documents by query
func (mdb *DB) Count(collectionName string, query bson.M, setupQuery func(query *mongo.Cursor), ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}
		return mdb.db.Collection(collectionName).CountDocuments(context.TODO(), query)
	}, ac)
}

// Insert inserts a document
func (mdb *DB) Insert(collectionName string, doc bson.M, ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}
		_, err := mdb.db.Collection(collectionName).InsertOne(context.TODO(), doc)
		return nil, err
	}, ac)
}

// InsertMany inserts multiple documents
func (mdb *DB) InsertMany(collectionName string, docs []bson.M, ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}
		insertDocs := make([]interface{}, len(docs))
		for i := 0; i < len(docs); i++ {
			insertDocs[i] = docs[i]
		}
		_, err := mdb.db.Collection(collectionName).InsertMany(context.TODO(), insertDocs)
		return nil, err
	}, ac)
}

// UpdateId updates a document by id
func (mdb *DB) UpdateId(collectionName string, id interface{}, update bson.M, ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		ret, err := mdb.db.Collection(collectionName).UpdateByID(context.TODO(), id, update)
		if err != nil {
			return nil, err
		}
		return ret.MatchedCount, err
	}, ac)
}

// Update updates a document by query
func (mdb *DB) Update(collectionName string, query bson.M, update bson.M, ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		_, err := mdb.db.Collection(collectionName).UpdateOne(context.TODO(), query, update)
		return nil, err
	}, ac)
}

// UpdateAll updates all documents by query
func (mdb *DB) UpdateAll(collectionName string, query bson.M, update bson.M, ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return 0, errNoSession
		}

		var updated int
		info, err := mdb.db.Collection(collectionName).UpdateMany(context.TODO(), query, update)
		if info != nil {
			updated = int(info.ModifiedCount)
		}
		return updated, err
	}, ac)
}

// UpsertId updates or inserts a document by id
func (mdb *DB) UpsertId(collectionName string, id interface{}, update bson.M, ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		opts := options.UpdateOne().SetUpsert(true)
		var upsertId interface{}
		info, err := mdb.db.Collection(collectionName).UpdateByID(context.TODO(), id, bson.M{"$set": update}, opts)
		if info != nil {
			upsertId = info.UpsertedID
		}
		return upsertId, err
	}, ac)
}

// Upsert updates or inserts a document by query
func (mdb *DB) Upsert(collectionName string, query bson.M, update bson.M, ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		opts := options.UpdateOne().SetUpsert(true)
		var upsertId interface{}
		info, err := mdb.db.Collection(collectionName).UpdateOne(context.TODO(), query, bson.M{"$set": update}, opts)
		if info != nil {
			upsertId = info.UpsertedID
		}
		return upsertId, err
	}, ac)
}

// RemoveId removes a document by id
func (mdb *DB) RemoveId(collectionName string, id interface{}, ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		ret, err := mdb.db.Collection(collectionName).DeleteOne(context.TODO(), bson.M{"_id": id})
		if err != nil {
			return nil, err
		}
		return ret.DeletedCount, err
	}, ac)
}

// Remove removes a document by query
func (mdb *DB) Remove(collectionName string, query bson.M, ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		ret, err := mdb.db.Collection(collectionName).DeleteOne(context.TODO(), query)
		if err != nil {
			return nil, err
		}
		return ret.DeletedCount, err
	}, ac)
}

// Remove removes all documents by query
func (mdb *DB) RemoveAll(collectionName string, query bson.M, ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return 0, errNoSession
		}

		var n int
		info, err := mdb.db.Collection(collectionName).DeleteMany(context.Background(), query)
		if info != nil {
			n = int(info.DeletedCount)
		}
		return n, err
	}, ac)
}

// EnsureIndex creates an index
func (mdb *DB) EnsureIndex(collectionName string, index mongo.IndexModel, ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		name, err := mdb.db.Collection(collectionName).Indexes().CreateMany(context.Background(), []mongo.IndexModel{index})
		return name, err
	}, ac)
}

// EnsureIndexKey creates an index by keys
func (mdb *DB) EnsureIndexKey(collectionName string, keys bson.D, ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		name, err := mdb.db.Collection(collectionName).Indexes().CreateOne(context.Background(), mongo.IndexModel{Keys: keys})
		return name, err
	}, ac)
}

// DropIndex drops an index by keys
func (mdb *DB) DropIndex(collectionName string, keys bson.D, ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		err := mdb.db.Collection(collectionName).Indexes().DropWithKey(context.Background(), keys)
		return nil, err
	}, ac)
}

//func (mdb *MongoDB) DropIndexName(collectionName string, indexName string, ac async.AsyncCallback) {
//	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
//		if !mdb.checkConnected() {
//			return nil, errNoSession
//			return
//		}
//
//		err := mdb.db.Collection(collectionName).DropIndexName(indexName)
//		return nil, err
//	}
//}

// DropCollection drops c collection
func (mdb *DB) DropCollection(collectionName string, ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		err := mdb.db.Collection(collectionName).Drop(context.Background())
		return nil, err
	}, ac)
}

// DropDatabase drops the database
func (mdb *DB) DropDatabase(ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		err := mdb.db.Drop(context.Background())
		return nil, err
	}, ac)
}
