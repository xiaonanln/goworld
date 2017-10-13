package mongodb

import (
	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/engine/async"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/netutil"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	opQueue = make(chan func(), consts.MONGODB_OP_QUEUE_MAXLEN)

	errNoSession = errors.Errorf("no session, please dail")
)

func init() {
	go netutil.ServeForever(func() {
		for op := range opQueue {
			op()
		}
	})
}

type MongoDB struct {
	session *mgo.Session
	db      *mgo.Database
}

func (mdb *MongoDB) checkSessionValid() bool {
	return mdb.session != nil && mdb.db != nil
}

func Dial(url string, dbname string, ac async.AsyncCallback) {
	opQueue <- func() {
		gwlog.Infof("Dailing MongoDB: %s ...", url)
		session, err := mgo.Dial(url)
		if err != nil {
			ac.Callback(nil, err)
			return
		}

		session.SetMode(mgo.Monotonic, true)
		db := session.DB(dbname)
		ac.Callback(&MongoDB{session, db}, nil)
	}
}

func (mdb *MongoDB) Close(ac async.AsyncCallback) {
	opQueue <- func() {
		if !mdb.checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		mdb.session.Close()
		mdb.session = nil
		mdb.db = nil
		ac.Callback(nil, nil)
	}
}

func (mdb *MongoDB) UseDB(dbname string, ac async.AsyncCallback) {
	opQueue <- func() {
		if !mdb.checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		mdb.db = mdb.session.DB(dbname)
		ac.Callback(nil, nil)
	}
}

func (mdb *MongoDB) SetMode(consistency mgo.Mode, ac async.AsyncCallback) {
	opQueue <- func() {
		if !mdb.checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		mdb.session.SetMode(consistency, true)
		ac.Callback(nil, nil)
	}
}

func (mdb *MongoDB) FindId(collectionName string, id interface{}, setupQuery func(query *mgo.Query), ac async.AsyncCallback) {
	opQueue <- func() {
		if !mdb.checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}
		q := mdb.db.C(collectionName).FindId(id)
		if setupQuery != nil {
			setupQuery(q)
		}
		var res bson.M
		err := q.One(&res)
		ac.Callback(res, err)
	}
}

func (mdb *MongoDB) FindOne(collectionName string, query bson.M, setupQuery func(query *mgo.Query), ac async.AsyncCallback) {
	opQueue <- func() {
		if !mdb.checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		q := mdb.db.C(collectionName).Find(query)
		if setupQuery != nil {
			setupQuery(q)
		}
		var res bson.M
		err := q.One(&res)
		ac.Callback(res, err)
	}
}

func (mdb *MongoDB) FindAll(collectionName string, query bson.M, setupQuery func(query *mgo.Query), ac async.AsyncCallback) {
	opQueue <- func() {
		if !mdb.checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}
		q := mdb.db.C(collectionName).Find(query)
		if setupQuery != nil {
			setupQuery(q)
		}
		var res []bson.M
		err := q.All(&res)
		ac.Callback(res, err)
	}
}

func (mdb *MongoDB) Count(collectionName string, query bson.M, setupQuery func(query *mgo.Query), ac async.AsyncCallback) {
	opQueue <- func() {
		if !mdb.checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}
		q := mdb.db.C(collectionName).Find(query)
		if setupQuery != nil {
			setupQuery(q)
		}
		n, err := q.Count()
		ac.Callback(n, err)
	}
}

func (mdb *MongoDB) Insert(collectionName string, doc bson.M, ac async.AsyncCallback) {
	opQueue <- func() {
		if !mdb.checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}
		err := mdb.db.C(collectionName).Insert(doc)
		ac.Callback(nil, err)
	}
}

func (mdb *MongoDB) InsertMany(collectionName string, docs []bson.M, ac async.AsyncCallback) {
	opQueue <- func() {
		if !mdb.checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}
		insertDocs := make([]interface{}, len(docs))
		for i := 0; i < len(docs); i++ {
			insertDocs[i] = docs[i]
		}
		err := mdb.db.C(collectionName).Insert(insertDocs...)
		ac.Callback(nil, err)
	}
}

func (mdb *MongoDB) UpdateId(collectionName string, id interface{}, update bson.M, ac async.AsyncCallback) {
	opQueue <- func() {
		if !mdb.checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		err := mdb.db.C(collectionName).UpdateId(id, update)
		ac.Callback(nil, err)
	}
}

func (mdb *MongoDB) Update(collectionName string, query bson.M, update bson.M, ac async.AsyncCallback) {
	opQueue <- func() {
		if !mdb.checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		err := mdb.db.C(collectionName).Update(query, update)
		ac.Callback(nil, err)
	}
}

func (mdb *MongoDB) UpdateAll(collectionName string, query bson.M, update bson.M, ac async.AsyncCallback) {
	opQueue <- func() {
		if !mdb.checkSessionValid() {
			ac.Callback(0, errNoSession)
			return
		}

		var updated int
		info, err := mdb.db.C(collectionName).UpdateAll(query, update)
		if info != nil {
			updated = info.Updated
		}
		ac.Callback(updated, err)
	}
}

func (mdb *MongoDB) UpsertId(collectionName string, id interface{}, update bson.M, ac async.AsyncCallback) {
	opQueue <- func() {
		if !mdb.checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		var upsertId interface{}
		info, err := mdb.db.C(collectionName).UpsertId(id, update)
		if info != nil {
			upsertId = info.UpsertedId
		}
		ac.Callback(upsertId, err)
	}
}

func (mdb *MongoDB) Upsert(collectionName string, query bson.M, update bson.M, ac async.AsyncCallback) {
	opQueue <- func() {
		if !mdb.checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		var upsertId interface{}
		info, err := mdb.db.C(collectionName).Upsert(query, update)
		if info != nil {
			upsertId = info.UpsertedId
		}
		ac.Callback(upsertId, err)
	}
}

func (mdb *MongoDB) RemoveId(collectionName string, id interface{}, ac async.AsyncCallback) {
	opQueue <- func() {
		if !mdb.checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		err := mdb.db.C(collectionName).RemoveId(id)
		ac.Callback(nil, err)
	}
}

func (mdb *MongoDB) Remove(collectionName string, query bson.M, ac async.AsyncCallback) {
	opQueue <- func() {
		if !mdb.checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		err := mdb.db.C(collectionName).Remove(query)
		ac.Callback(nil, err)
	}
}

func (mdb *MongoDB) RemoveAll(collectionName string, query bson.M, ac async.AsyncCallback) {
	opQueue <- func() {
		if !mdb.checkSessionValid() {
			ac.Callback(0, errNoSession)
			return
		}

		var n int
		info, err := mdb.db.C(collectionName).RemoveAll(query)
		if info != nil {
			n = info.Removed
		}
		ac.Callback(n, err)
	}
}

func (mdb *MongoDB) EnsureIndex(collectionName string, index mgo.Index, ac async.AsyncCallback) {
	opQueue <- func() {
		if !mdb.checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		err := mdb.db.C(collectionName).EnsureIndex(index)
		ac.Callback(nil, err)
	}
}

func (mdb *MongoDB) EnsureIndexKey(collectionName string, keys []string, ac async.AsyncCallback) {
	opQueue <- func() {
		if !mdb.checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		err := mdb.db.C(collectionName).EnsureIndexKey(keys...)
		ac.Callback(nil, err)
	}
}

func (mdb *MongoDB) DropIndex(collectionName string, keys []string, ac async.AsyncCallback) {
	opQueue <- func() {
		if !mdb.checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		err := mdb.db.C(collectionName).DropIndex(keys...)
		ac.Callback(nil, err)
	}
}

//func (mdb *MongoDB) DropIndexName(collectionName string, indexName string, ac async.AsyncCallback) {
//	opQueue <- func() {
//		if !mdb.checkSessionValid() {
//			ac.Callback(nil, errNoSession)
//			return
//		}
//
//		err := mdb.db.C(collectionName).DropIndexName(indexName)
//		ac.Callback(nil, err)
//	}
//}

func (mdb *MongoDB) DropCollection(collectionName string, ac async.AsyncCallback) {
	opQueue <- func() {
		if !mdb.checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		err := mdb.db.C(collectionName).DropCollection()
		ac.Callback(nil, err)
	}
}

func (mdb *MongoDB) DropDatabase(ac async.AsyncCallback) {
	opQueue <- func() {
		if !mdb.checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		err := mdb.db.DropDatabase()
		ac.Callback(nil, err)
	}
}
