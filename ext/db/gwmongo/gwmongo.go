package gwmongo

import (
	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/engine/async"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	_MONGODB_ASYNC_JOB_GROUP = "_mongodb"
)

var (
	errNoSession = errors.Errorf("no session, please dail")
)

// MongoDB is a MongoDB instance can be used to manipulate Mongo DBs
type DB struct {
	session *mgo.Session
	db      *mgo.Database
}

func (mdb *DB) checkConnected() bool {
	return mdb.session != nil && mdb.db != nil
}

// Dial connects the a MongoDB
// returns *MongoDB
func Dial(url string, dbname string, ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		gwlog.Infof("Dailing MongoDB: %s ...", url)
		session, err := mgo.Dial(url)
		if err != nil {
			return nil, err
		}

		session.SetMode(mgo.Monotonic, true)
		db := session.DB(dbname)
		return &DB{session, db}, nil
	}, ac)
}

// Close closes MongoDB
func (mdb *DB) Close(ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		mdb.session.Close()
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

		mdb.db = mdb.session.DB(dbname)
		return nil, nil
	}, ac)
}

// SetMode sets the consistency mode
func (mdb *DB) SetMode(consistency mgo.Mode, ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		mdb.session.SetMode(consistency, true)
		return nil, nil
	}, ac)
}

// FindId finds document in collection by Id
func (mdb *DB) FindId(collectionName string, id interface{}, setupQuery func(query *mgo.Query), ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}
		q := mdb.db.C(collectionName).FindId(id)
		if setupQuery != nil {
			setupQuery(q)
		}
		var res bson.M
		err := q.One(&res)
		return res, err
	}, ac)
}

// FindOne finds one document with specified query
func (mdb *DB) FindOne(collectionName string, query bson.M, setupQuery func(query *mgo.Query), ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		q := mdb.db.C(collectionName).Find(query)
		if setupQuery != nil {
			setupQuery(q)
		}
		var res bson.M
		err := q.One(&res)
		return res, err
	}, ac)
}

// FindAll finds all documents with specified query
func (mdb *DB) FindAll(collectionName string, query bson.M, setupQuery func(query *mgo.Query), ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}
		q := mdb.db.C(collectionName).Find(query)
		if setupQuery != nil {
			setupQuery(q)
		}
		var res []bson.M
		err := q.All(&res)
		return res, err
	}, ac)
}

// Count counts the number of documents by query
func (mdb *DB) Count(collectionName string, query bson.M, setupQuery func(query *mgo.Query), ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}
		q := mdb.db.C(collectionName).Find(query)
		if setupQuery != nil {
			setupQuery(q)
		}
		n, err := q.Count()
		return n, err
	}, ac)
}

// Insert inserts a document
func (mdb *DB) Insert(collectionName string, doc bson.M, ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}
		err := mdb.db.C(collectionName).Insert(doc)
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
		err := mdb.db.C(collectionName).Insert(insertDocs...)
		return nil, err
	}, ac)
}

// UpdateId updates a document by id
func (mdb *DB) UpdateId(collectionName string, id interface{}, update bson.M, ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		err := mdb.db.C(collectionName).UpdateId(id, update)
		return nil, err
	}, ac)
}

// Update updates a document by query
func (mdb *DB) Update(collectionName string, query bson.M, update bson.M, ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		err := mdb.db.C(collectionName).Update(query, update)
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
		info, err := mdb.db.C(collectionName).UpdateAll(query, update)
		if info != nil {
			updated = info.Updated
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

		var upsertId interface{}
		info, err := mdb.db.C(collectionName).UpsertId(id, update)
		if info != nil {
			upsertId = info.UpsertedId
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

		var upsertId interface{}
		info, err := mdb.db.C(collectionName).Upsert(query, update)
		if info != nil {
			upsertId = info.UpsertedId
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

		err := mdb.db.C(collectionName).RemoveId(id)
		return nil, err
	}, ac)
}

// Remove removes a document by query
func (mdb *DB) Remove(collectionName string, query bson.M, ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		err := mdb.db.C(collectionName).Remove(query)
		return nil, err
	}, ac)
}

// Remove removes all documents by query
func (mdb *DB) RemoveAll(collectionName string, query bson.M, ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return 0, errNoSession
		}

		var n int
		info, err := mdb.db.C(collectionName).RemoveAll(query)
		if info != nil {
			n = info.Removed
		}
		return n, err
	}, ac)
}

// EnsureIndex creates an index
func (mdb *DB) EnsureIndex(collectionName string, index mgo.Index, ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		err := mdb.db.C(collectionName).EnsureIndex(index)
		return nil, err
	}, ac)
}

// EnsureIndexKey creates an index by keys
func (mdb *DB) EnsureIndexKey(collectionName string, keys []string, ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		err := mdb.db.C(collectionName).EnsureIndexKey(keys...)
		return nil, err
	}, ac)
}

// DropIndex drops an index by keys
func (mdb *DB) DropIndex(collectionName string, keys []string, ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		err := mdb.db.C(collectionName).DropIndex(keys...)
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
//		err := mdb.db.C(collectionName).DropIndexName(indexName)
//		return nil, err
//	}
//}

// DropCollection drops c collection
func (mdb *DB) DropCollection(collectionName string, ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		err := mdb.db.C(collectionName).DropCollection()
		return nil, err
	}, ac)
}

// DropDatabase drops the database
func (mdb *DB) DropDatabase(ac async.AsyncCallback) {
	async.AppendAsyncJob(_MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		err := mdb.db.DropDatabase()
		return nil, err
	}, ac)
}
