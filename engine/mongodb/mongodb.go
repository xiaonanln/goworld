package mongodb

import (
	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/engine/async"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	MONGODB_ASYNC_JOB_GROUP = "mongodb"
)

var (
	errNoSession = errors.Errorf("no session, please dail")
)

type MongoDB struct {
	session *mgo.Session
	db      *mgo.Database
}

func (mdb *MongoDB) checkConnected() bool {
	return mdb.session != nil && mdb.db != nil
}

func Dial(url string, dbname string, ac async.AsyncCallback) {
	async.AppendAsyncJob(MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		gwlog.Infof("Dailing MongoDB: %s ...", url)
		session, err := mgo.Dial(url)
		if err != nil {
			return nil, err
		}

		session.SetMode(mgo.Monotonic, true)
		db := session.DB(dbname)
		return &MongoDB{session, db}, nil
	}, ac)
}

func (mdb *MongoDB) Close(ac async.AsyncCallback) {
	async.AppendAsyncJob(MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		mdb.session.Close()
		mdb.session = nil
		mdb.db = nil
		return nil, nil
	}, ac)
}

func (mdb *MongoDB) UseDB(dbname string, ac async.AsyncCallback) {
	async.AppendAsyncJob(MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		mdb.db = mdb.session.DB(dbname)
		return nil, nil
	}, ac)
}

func (mdb *MongoDB) SetMode(consistency mgo.Mode, ac async.AsyncCallback) {
	async.AppendAsyncJob(MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		mdb.session.SetMode(consistency, true)
		return nil, nil
	}, ac)
}

func (mdb *MongoDB) FindId(collectionName string, id interface{}, setupQuery func(query *mgo.Query), ac async.AsyncCallback) {
	async.AppendAsyncJob(MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
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

func (mdb *MongoDB) FindOne(collectionName string, query bson.M, setupQuery func(query *mgo.Query), ac async.AsyncCallback) {
	async.AppendAsyncJob(MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
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

func (mdb *MongoDB) FindAll(collectionName string, query bson.M, setupQuery func(query *mgo.Query), ac async.AsyncCallback) {
	async.AppendAsyncJob(MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
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

func (mdb *MongoDB) Count(collectionName string, query bson.M, setupQuery func(query *mgo.Query), ac async.AsyncCallback) {
	async.AppendAsyncJob(MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
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

func (mdb *MongoDB) Insert(collectionName string, doc bson.M, ac async.AsyncCallback) {
	async.AppendAsyncJob(MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}
		err := mdb.db.C(collectionName).Insert(doc)
		return nil, err
	}, ac)
}

func (mdb *MongoDB) InsertMany(collectionName string, docs []bson.M, ac async.AsyncCallback) {
	async.AppendAsyncJob(MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
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

func (mdb *MongoDB) UpdateId(collectionName string, id interface{}, update bson.M, ac async.AsyncCallback) {
	async.AppendAsyncJob(MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		err := mdb.db.C(collectionName).UpdateId(id, update)
		return nil, err
	}, ac)
}

func (mdb *MongoDB) Update(collectionName string, query bson.M, update bson.M, ac async.AsyncCallback) {
	async.AppendAsyncJob(MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		err := mdb.db.C(collectionName).Update(query, update)
		return nil, err
	}, ac)
}

func (mdb *MongoDB) UpdateAll(collectionName string, query bson.M, update bson.M, ac async.AsyncCallback) {
	async.AppendAsyncJob(MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
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

func (mdb *MongoDB) UpsertId(collectionName string, id interface{}, update bson.M, ac async.AsyncCallback) {
	async.AppendAsyncJob(MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
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

func (mdb *MongoDB) Upsert(collectionName string, query bson.M, update bson.M, ac async.AsyncCallback) {
	async.AppendAsyncJob(MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
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

func (mdb *MongoDB) RemoveId(collectionName string, id interface{}, ac async.AsyncCallback) {
	async.AppendAsyncJob(MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		err := mdb.db.C(collectionName).RemoveId(id)
		return nil, err
	}, ac)
}

func (mdb *MongoDB) Remove(collectionName string, query bson.M, ac async.AsyncCallback) {
	async.AppendAsyncJob(MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		err := mdb.db.C(collectionName).Remove(query)
		return nil, err
	}, ac)
}

func (mdb *MongoDB) RemoveAll(collectionName string, query bson.M, ac async.AsyncCallback) {
	async.AppendAsyncJob(MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
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

func (mdb *MongoDB) EnsureIndex(collectionName string, index mgo.Index, ac async.AsyncCallback) {
	async.AppendAsyncJob(MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		err := mdb.db.C(collectionName).EnsureIndex(index)
		return nil, err
	}, ac)
}

func (mdb *MongoDB) EnsureIndexKey(collectionName string, keys []string, ac async.AsyncCallback) {
	async.AppendAsyncJob(MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		err := mdb.db.C(collectionName).EnsureIndexKey(keys...)
		return nil, err
	}, ac)
}

func (mdb *MongoDB) DropIndex(collectionName string, keys []string, ac async.AsyncCallback) {
	async.AppendAsyncJob(MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		err := mdb.db.C(collectionName).DropIndex(keys...)
		return nil, err
	}, ac)
}

//func (mdb *MongoDB) DropIndexName(collectionName string, indexName string, ac async.AsyncCallback) {
//	async.AppendAsyncJob(MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
//		if !mdb.checkConnected() {
//			return nil, errNoSession
//			return
//		}
//
//		err := mdb.db.C(collectionName).DropIndexName(indexName)
//		return nil, err
//	}
//}

func (mdb *MongoDB) DropCollection(collectionName string, ac async.AsyncCallback) {
	async.AppendAsyncJob(MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		err := mdb.db.C(collectionName).DropCollection()
		return nil, err
	}, ac)
}

func (mdb *MongoDB) DropDatabase(ac async.AsyncCallback) {
	async.AppendAsyncJob(MONGODB_ASYNC_JOB_GROUP, func() (interface{}, error) {
		if !mdb.checkConnected() {
			return nil, errNoSession
		}

		err := mdb.db.DropDatabase()
		return nil, err
	}, ac)
}
