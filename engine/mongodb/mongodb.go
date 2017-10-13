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

func Dial(url string, dbname string, async *async.AsyncRequest) {
	opQueue <- func() {
		if checkSessionValid() {
			async.Error(errors.Errorf("multiple dail"))
			return
		}

		var err error
		gwlog.Infof("Dailing MongoDB: %s ...", url)
		session, err = mgo.Dial(url)
		if err != nil {
			async.Error(err)
			return
		}

		session.SetMode(mgo.Monotonic, true)
		if dbname != "" {
			db = session.DB(dbname)
		}
		async.Done()
	}
}

func Close(async *async.AsyncRequest) {
	opQueue <- func() {
		if !checkSessionValid() {
			async.Error(errNoSession)
			return
		}

		session.Close()
		session = nil
		db = nil
		async.Done()
	}
}

func UseDB(dbname string, async *async.AsyncRequest) {
	opQueue <- func() {
		if !checkSessionValid() {
			async.Error(errNoSession)
			return
		}

		db = session.DB(dbname)
		async.Done()
	}
}

func SetMode(consistency mgo.Mode, async *async.AsyncRequest) {
	opQueue <- func() {
		if !checkSessionValid() {
			async.Error(errNoSession)
			return
		}

		session.SetMode(consistency, true)
		async.Done()
	}
}

func FindId(collectionName string, id interface{}, setupQuery func(query *mgo.Query), async *async.AsyncRequest) {
	opQueue <- func() {
		if !checkSessionValid() {
			async.Error(errNoSession)
			return
		}
		q := db.C(collectionName).FindId(id)
		if setupQuery != nil {
			setupQuery(q)
		}
		var res bson.M
		err := q.One(&res)
		async.Error(err, res)
	}
}

func FindOne(collectionName string, query bson.M, setupQuery func(query *mgo.Query), async *async.AsyncRequest) {
	opQueue <- func() {
		if !checkSessionValid() {
			async.Error(errNoSession)
			return
		}

		q := db.C(collectionName).Find(query)
		if setupQuery != nil {
			setupQuery(q)
		}
		var res bson.M
		err := q.One(&res)
		async.Error(err, res)
	}
}

func FindAll(collectionName string, query bson.M, setupQuery func(query *mgo.Query), async *async.AsyncRequest) {
	opQueue <- func() {
		if !checkSessionValid() {
			async.Error(errNoSession)
			return
		}
		q := db.C(collectionName).Find(query)
		if setupQuery != nil {
			setupQuery(q)
		}
		var res []bson.M
		err := q.All(&res)
		async.Error(err, res)
	}
}

func Count(collectionName string, query bson.M, setupQuery func(query *mgo.Query), async *async.AsyncRequest) {
	opQueue <- func() {
		if !checkSessionValid() {
			async.Error(errNoSession)
			return
		}
		q := db.C(collectionName).Find(query)
		if setupQuery != nil {
			setupQuery(q)
		}
		n, err := q.Count()
		async.Error(err, n)
	}
}

func Insert(collectionName string, doc bson.M, async *async.AsyncRequest) {
	opQueue <- func() {
		if !checkSessionValid() {
			async.Error(errNoSession)
			return
		}
		err := db.C(collectionName).Insert(doc)
		async.Error(err)
	}
}

func InsertMany(collectionName string, docs []bson.M, async *async.AsyncRequest) {
	opQueue <- func() {
		if !checkSessionValid() {
			async.Error(errNoSession)
			return
		}
		insertDocs := make([]interface{}, len(docs))
		for i := 0; i < len(docs); i++ {
			insertDocs[i] = docs[i]
		}
		err := db.C(collectionName).Insert(insertDocs...)
		async.Error(err)
	}
}

func UpdateId(collectionName string, id interface{}, update bson.M, async *async.AsyncRequest) {
	opQueue <- func() {
		if !checkSessionValid() {
			async.Error(errNoSession)
			return
		}

		err := db.C(collectionName).UpdateId(id, update)
		async.Error(err)
	}
}

func Update(collectionName string, query bson.M, update bson.M, async *async.AsyncRequest) {
	opQueue <- func() {
		if !checkSessionValid() {
			async.Error(errNoSession)
			return
		}

		err := db.C(collectionName).Update(query, update)
		async.Error(err)
	}
}

func UpdateAll(collectionName string, query bson.M, update bson.M, async *async.AsyncRequest) {
	opQueue <- func() {
		if !checkSessionValid() {
			async.Error(errNoSession, 0)
			return
		}

		var updated int
		info, err := db.C(collectionName).UpdateAll(query, update)
		if err == nil {
			updated = info.Updated
		}
		async.Error(err, updated)
	}
}

func DropCollection(collectionName string, async *async.AsyncRequest) {
	opQueue <- func() {
		if !checkSessionValid() {
			async.Error(errNoSession)
			return
		}

		err := db.C(collectionName).DropCollection()
		async.Error(err)
	}
}

func DropDatabase(async *async.AsyncRequest) {
	opQueue <- func() {
		if !checkSessionValid() {
			async.Error(errNoSession)
			return
		}

		err := db.DropDatabase()
		async.Error(err)
	}
}
