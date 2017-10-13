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
	session *mgo.Session
	db      *mgo.Database
	opQueue = make(chan func(), consts.MONGODB_OP_QUEUE_MAXLEN)

	errNoSession = errors.Errorf("no session, please dail")
)

func init() {
	go netutil.ServeForever(func() {
		for {
			op := <-opQueue
			op()
		}
	})
}

func checkSessionValid() bool {
	return session != nil
}

func Dial(url string, dbname string, ac async.AsyncCallback) {
	opQueue <- func() {
		if checkSessionValid() {
			ac.Callback(nil, errors.Errorf("multiple dail"))
			return
		}

		var err error
		gwlog.Infof("Dailing MongoDB: %s ...", url)
		session, err = mgo.Dial(url)
		if err != nil {
			ac.Callback(nil, err)
			return
		}

		session.SetMode(mgo.Monotonic, true)
		if dbname != "" {
			db = session.DB(dbname)
		}
		ac.Callback(nil, nil)
	}
}

func Close(ac async.AsyncCallback) {
	opQueue <- func() {
		if !checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		session.Close()
		session = nil
		db = nil
		ac.Callback(nil, nil)
	}
}

func UseDB(dbname string, ac async.AsyncCallback) {
	opQueue <- func() {
		if !checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		db = session.DB(dbname)
		ac.Callback(nil, nil)
	}
}

func SetMode(consistency mgo.Mode, ac async.AsyncCallback) {
	opQueue <- func() {
		if !checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		session.SetMode(consistency, true)
		ac.Callback(nil, nil)
	}
}

func FindId(collectionName string, id interface{}, setupQuery func(query *mgo.Query), ac async.AsyncCallback) {
	opQueue <- func() {
		if !checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}
		q := db.C(collectionName).FindId(id)
		if setupQuery != nil {
			setupQuery(q)
		}
		var res bson.M
		err := q.One(&res)
		ac.Callback(res, err)
	}
}

func FindOne(collectionName string, query bson.M, setupQuery func(query *mgo.Query), ac async.AsyncCallback) {
	opQueue <- func() {
		if !checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		q := db.C(collectionName).Find(query)
		if setupQuery != nil {
			setupQuery(q)
		}
		var res bson.M
		err := q.One(&res)
		ac.Callback(res, err)
	}
}

func FindAll(collectionName string, query bson.M, setupQuery func(query *mgo.Query), ac async.AsyncCallback) {
	opQueue <- func() {
		if !checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}
		q := db.C(collectionName).Find(query)
		if setupQuery != nil {
			setupQuery(q)
		}
		var res []bson.M
		err := q.All(&res)
		ac.Callback(res, err)
	}
}

func Count(collectionName string, query bson.M, setupQuery func(query *mgo.Query), ac async.AsyncCallback) {
	opQueue <- func() {
		if !checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}
		q := db.C(collectionName).Find(query)
		if setupQuery != nil {
			setupQuery(q)
		}
		n, err := q.Count()
		ac.Callback(n, err)
	}
}

func Insert(collectionName string, doc bson.M, ac async.AsyncCallback) {
	opQueue <- func() {
		if !checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}
		err := db.C(collectionName).Insert(doc)
		ac.Callback(nil, err)
	}
}

func InsertMany(collectionName string, docs []bson.M, ac async.AsyncCallback) {
	opQueue <- func() {
		if !checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}
		insertDocs := make([]interface{}, len(docs))
		for i := 0; i < len(docs); i++ {
			insertDocs[i] = docs[i]
		}
		err := db.C(collectionName).Insert(insertDocs...)
		ac.Callback(nil, err)
	}
}

func UpdateId(collectionName string, id interface{}, update bson.M, ac async.AsyncCallback) {
	opQueue <- func() {
		if !checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		err := db.C(collectionName).UpdateId(id, update)
		ac.Callback(nil, err)
	}
}

func Update(collectionName string, query bson.M, update bson.M, ac async.AsyncCallback) {
	opQueue <- func() {
		if !checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		err := db.C(collectionName).Update(query, update)
		ac.Callback(nil, err)
	}
}

func UpdateAll(collectionName string, query bson.M, update bson.M, ac async.AsyncCallback) {
	opQueue <- func() {
		if !checkSessionValid() {
			ac.Callback(0, errNoSession)
			return
		}

		var updated int
		info, err := db.C(collectionName).UpdateAll(query, update)
		if info != nil {
			updated = info.Updated
		}
		ac.Callback(updated, err)
	}
}

func UpsertId(collectionName string, id interface{}, update bson.M, ac async.AsyncCallback) {
	opQueue <- func() {
		if !checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		var upsertId interface{}
		info, err := db.C(collectionName).UpsertId(id, update)
		if info != nil {
			upsertId = info.UpsertedId
		}
		ac.Callback(upsertId, err)
	}
}

func Upsert(collectionName string, query bson.M, update bson.M, ac async.AsyncCallback) {
	opQueue <- func() {
		if !checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		var upsertId interface{}
		info, err := db.C(collectionName).Upsert(query, update)
		if info != nil {
			upsertId = info.UpsertedId
		}
		ac.Callback(upsertId, err)
	}
}

func RemoveId(collectionName string, id interface{}, ac async.AsyncCallback) {
	opQueue <- func() {
		if !checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		err := db.C(collectionName).RemoveId(id)
		ac.Callback(nil, err)
	}
}

func Remove(collectionName string, query bson.M, ac async.AsyncCallback) {
	opQueue <- func() {
		if !checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		err := db.C(collectionName).Remove(query)
		ac.Callback(nil, err)
	}
}

func RemoveAll(collectionName string, query bson.M, ac async.AsyncCallback) {
	opQueue <- func() {
		if !checkSessionValid() {
			ac.Callback(0, errNoSession)
			return
		}

		var n int
		info, err := db.C(collectionName).RemoveAll(query)
		if info != nil {
			n = info.Removed
		}
		ac.Callback(n, err)
	}
}

func EnsureIndex(collectionName string, index mgo.Index, ac async.AsyncCallback) {
	opQueue <- func() {
		if !checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		err := db.C(collectionName).EnsureIndex(index)
		ac.Callback(nil, err)
	}
}

func EnsureIndexKey(collectionName string, keys []string, ac async.AsyncCallback) {
	opQueue <- func() {
		if !checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		err := db.C(collectionName).EnsureIndexKey(keys...)
		ac.Callback(nil, err)
	}
}

func DropIndex(collectionName string, keys []string, ac async.AsyncCallback) {
	opQueue <- func() {
		if !checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		err := db.C(collectionName).DropIndex(keys...)
		ac.Callback(nil, err)
	}
}

//func DropIndexName(collectionName string, indexName string, ac async.AsyncCallback) {
//	opQueue <- func() {
//		if !checkSessionValid() {
//			ac.Callback(nil, errNoSession)
//			return
//		}
//
//		err := db.C(collectionName).DropIndexName(indexName)
//		ac.Callback(nil, err)
//	}
//}

func DropCollection(collectionName string, ac async.AsyncCallback) {
	opQueue <- func() {
		if !checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		err := db.C(collectionName).DropCollection()
		ac.Callback(nil, err)
	}
}

func DropDatabase(ac async.AsyncCallback) {
	opQueue <- func() {
		if !checkSessionValid() {
			ac.Callback(nil, errNoSession)
			return
		}

		err := db.DropDatabase()
		ac.Callback(nil, err)
	}
}
