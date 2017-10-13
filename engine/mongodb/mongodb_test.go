package mongodb

import (
	"testing"

	"time"

	"sync"

	"github.com/xiaonanln/goworld/engine/async"
	"github.com/xiaonanln/goworld/engine/post"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var wait sync.WaitGroup

func TestDial(t *testing.T) {
	wait.Add(1)
	Dial("mongodb://localhost:27017/", "goworld", async.NewAsyncRequest(func(err error, res ...interface{}) {
		checkRequest(t, err, res)
		wait.Done()
	}))
	wait.Wait()
}

func TestClose(t *testing.T) {
	wait.Add(1)
	Close(async.NewAsyncRequest(func(err error, res ...interface{}) {
		checkRequest(t, err, res)
		wait.Done()
	}))
	wait.Wait()
	wait.Add(1)
	Dial("mongodb://localhost:27017/", "goworld", async.NewAsyncRequest(func(err error, res ...interface{}) {
		checkRequest(t, err, res)
		wait.Done()
	}))
	wait.Wait()
}

func TestSetMode(t *testing.T) {
	wait.Add(1)
	SetMode(mgo.SecondaryPreferred, async.NewAsyncRequest(func(err error, res ...interface{}) {
		checkRequest(t, err, res)
		wait.Done()
	}))
	wait.Wait()
	wait.Add(1)
	SetMode(mgo.Monotonic, async.NewAsyncRequest(func(err error, res ...interface{}) {
		checkRequest(t, err, res)
		wait.Done()
	}))
	wait.Wait()
}

func TestUseDB(t *testing.T) {
	wait.Add(1)
	UseDB("abc", async.NewAsyncRequest(func(err error, res ...interface{}) {
		checkRequest(t, err, res)
		wait.Done()
	}))
	wait.Wait()

	wait.Add(1)
	UseDB("goworld", async.NewAsyncRequest(func(err error, res ...interface{}) {
		checkRequest(t, err, res)
		wait.Done()
	}))
	wait.Wait()
}

func TestInsert(t *testing.T) {
	wait.Add(1)
	Insert("mongodb_test", bson.M{"a": 1, "b": 2}, async.NewAsyncRequest(func(err error, res ...interface{}) {
		wait.Done()
	}))
	wait.Wait()
}

func TestInsertMany(t *testing.T) {
	wait.Add(1)
	InsertMany("mongodb_test", []bson.M{
		{"c": 1, "d": 1},
		{"c": 2, "d": 2},
		{"c": 3, "d": 3},
	}, async.NewAsyncRequest(func(err error, res ...interface{}) {
		wait.Done()
	}))
	wait.Wait()
}

func TestCount(t *testing.T) {
	wait.Add(1)
	Count("mongodb_test", bson.M{"c": 1}, nil, async.NewAsyncRequest(func(err error, res ...interface{}) {
		checkRequest(t, err, res)
		count := res[0].(int)
		t.Logf("Count returns %d", count)
		wait.Done()
	}))
	wait.Wait()
}

func TestFindOne(t *testing.T) {
	wait.Add(1)
	FindOne("mongodb_test", bson.M{"c": 2}, func(query *mgo.Query) {
		query.Limit(2)
		query.Sort("d", "a", "b")
		query.Select(bson.M{"_id": 0})
	}, async.NewAsyncRequest(func(err error, res ...interface{}) {
		checkRequest(t, err, res)
		t.Logf("FindOne: %v", res[0].(bson.M))
		wait.Done()
	}))

	wait.Wait()
}

func TestFindAll(t *testing.T) {
	wait.Add(1)
	FindAll("mongodb_test", bson.M{"c": 2}, func(query *mgo.Query) {
		query.Sort("d", "a", "b")
		query.Select(bson.M{"_id": 0})
	}, async.NewAsyncRequest(func(err error, res ...interface{}) {
		checkRequest(t, err, res)
		t.Logf("FindAll: %v", res[0].([]bson.M))
		wait.Done()
	}))

	wait.Wait()
}

func TestFindId(t *testing.T) {
	id := bson.NewObjectId()
	Insert("mongodb_test", bson.M{"_id": id, "TestFindId": 1}, nil)

	wait.Add(1)
	FindId("mongodb_test", id, nil, async.NewAsyncRequest(func(err error, res ...interface{}) {
		checkRequest(t, err, res)
		t.Logf("FindId: %v", res[0].(bson.M))
		wait.Done()
	}))
	wait.Wait()
}

func TestUpdate(t *testing.T) {
	wait.Add(1)
	Update("mongodb_test", bson.M{"TestFindId": 1}, bson.M{"$set": bson.M{"TestUpdate": 1}}, async.NewAsyncRequest(func(err error, res ...interface{}) {
		checkRequest(t, err, res)
		wait.Done()
	}))
	wait.Wait()

	wait.Add(1)
	FindOne("mongodb_test", bson.M{"TestUpdate": 1}, nil, async.NewAsyncRequest(func(err error, res ...interface{}) {
		checkRequest(t, err, res)
		t.Logf("Update: %v", res[0].(bson.M))
		wait.Done()
	}))
	wait.Wait()
}

func TestUpdateAll(t *testing.T) {
	wait.Add(1)
	UpdateAll("mongodb_test", bson.M{"c": 2}, bson.M{"$set": bson.M{"c": "3"}}, async.NewAsyncRequest(func(err error, res ...interface{}) {
		checkRequest(t, err, res)
		t.Logf("UpdateAll: %v", res[0].(int))
		wait.Done()
	}))
	wait.Wait()
}

func TestDropCollection(t *testing.T) {
	wait.Add(1)
	DropCollection("mongodb_test", async.NewAsyncRequest(func(err error, res ...interface{}) {
		checkRequest(t, err, res)
		wait.Done()
	}))
	wait.Wait()
}

//func TestDropDatabase(t *testing.T) {
//	wait.Add(1)
//	DropDatabase(async.NewAsyncRequest(func(err error, res ...interface{}) {
//		checkRequest(t, err, res)
//		wait.Done()
//	}))
//	wait.Wait()
//}

func checkRequest(t *testing.T, err error, res []interface{}) {
	if err != nil {
		t.Errorf("error = %v, res = %v", err, res)
	} else {
		t.Logf("success, res = %v", res)
	}
}

func init() {
	go func() {
		for {
			post.Tick()
			time.Sleep(time.Millisecond * 100)
		}
	}()
}
