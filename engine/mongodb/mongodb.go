package mongodb

import (
	"github.com/xiaonanln/goworld/engine/async"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	session *mgo.Session
	db      *mgo.Database
)

func Dial(url string, dbname string, async AsyncRequest) {
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

func UseDB(dbname string, async AsyncRequest) {
	db = session.DB(dbname)
	async.Done()
}

func SetMode(consistency mgo.Mode, async AsyncRequest) {
	session.SetMode(consistency, true)
	async.Done()
}

func Insert(collectionName string, docs ...interface{}) {
	err := db.C(collectionName).Insert(docs...)
	if err != nil {
		async.Error(err)
	} else {
		async.Done()
	}
}

func FindOne(collectionName string, query interface{}, async AsyncRequest) {
	q := db.C(collectionName).Find(query)

	var res bson.M
	err := q.One(&res)
	if err != nil {
		async.Error(err)
		return
	}
	async.Done(res)
}

func UpdateId(collectionName string, id interface{}, async AsyncRequest) {
	q := db.C(collectionName).FindId(id)
	var res bson.M
	err := q.One(&res)
	if err != nil {
		async.Error(err)
		return
	}

	async.Done(res)
}
