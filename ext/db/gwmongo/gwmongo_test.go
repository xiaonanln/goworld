package gwmongo

import (
	"testing"

	"time"

	"sync"

	"github.com/xiaonanln/goworld/engine/post"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var wait sync.WaitGroup
var mongodb *DB

func TestDial(t *testing.T) {
	wait.Add(1)
	Dial("mongodb://localhost:27017/", "goworld", func(res interface{}, err error) {
		checkRequest(t, err, res)
		mongodb = res.(*DB)
		wait.Done()
	})
	wait.Wait()
}

func TestClose(t *testing.T) {
	wait.Add(1)
	mongodb.Close(func(res interface{}, err error) {
		checkRequest(t, err, res)
		wait.Done()
	})
	wait.Wait()
	wait.Add(1)
	Dial("mongodb://localhost:27017/", "goworld", func(res interface{}, err error) {
		checkRequest(t, err, res)
		mongodb = res.(*DB)
		wait.Done()
	})
	wait.Wait()
}

func TestSetMode(t *testing.T) {
	wait.Add(1)
	mongodb.SetMode(mgo.SecondaryPreferred, func(res interface{}, err error) {
		checkRequest(t, err, res)
	})
	mongodb.SetMode(mgo.Monotonic, func(res interface{}, err error) {
		checkRequest(t, err, res)
		wait.Done()
	})
	wait.Wait()
}

func TestUseDB(t *testing.T) {
	wait.Add(1)
	mongodb.UseDB("abc", func(res interface{}, err error) {
		checkRequest(t, err, res)
		wait.Done()
	})
	wait.Wait()

	wait.Add(1)
	mongodb.UseDB("goworld", func(res interface{}, err error) {
		checkRequest(t, err, res)
		wait.Done()
	})
	wait.Wait()
}

func TestInsert(t *testing.T) {
	wait.Add(1)
	mongodb.Insert("mongodb_test", bson.M{"a": 1, "b": 2}, func(res interface{}, err error) {
		wait.Done()
	})
	wait.Wait()
}

func TestInsertMany(t *testing.T) {
	wait.Add(1)
	mongodb.InsertMany("mongodb_test", []bson.M{
		{"c": 1, "d": 1},
		{"c": 2, "d": 2},
		{"c": 3, "d": 3},
	}, func(res interface{}, err error) {
		wait.Done()
	})
	wait.Wait()
}

func TestCount(t *testing.T) {
	wait.Add(1)
	mongodb.Count("mongodb_test", bson.M{"c": 1}, nil, func(res interface{}, err error) {
		checkRequest(t, err, res)
		count := res.(int)
		t.Logf("Count returns %d", count)
		wait.Done()
	})
	wait.Wait()
}

func TestFindOne(t *testing.T) {
	wait.Add(1)
	mongodb.FindOne("mongodb_test", bson.M{"c": 2}, func(query *mgo.Query) {
		query.Limit(2)
		query.Sort("d", "a", "b")
		query.Select(bson.M{"_id": 0})
	}, func(res interface{}, err error) {
		checkRequest(t, err, res)
		t.Logf("FindOne: %v", res.(bson.M))
		wait.Done()
	})

	wait.Wait()
}

func TestFindAll(t *testing.T) {
	wait.Add(1)
	mongodb.FindAll("mongodb_test", bson.M{"c": 2}, func(query *mgo.Query) {
		query.Sort("d", "a", "b")
		query.Select(bson.M{"_id": 0})
	}, func(res interface{}, err error) {
		checkRequest(t, err, res)
		t.Logf("FindAll: %v", res.([]bson.M))
		wait.Done()
	})

	wait.Wait()
}

func TestFindId(t *testing.T) {
	id := bson.NewObjectId()
	mongodb.Insert("mongodb_test", bson.M{"_id": id, "TestFindId": 1}, nil)

	wait.Add(1)
	mongodb.FindId("mongodb_test", id, nil, func(res interface{}, err error) {
		checkRequest(t, err, res)
		t.Logf("FindId: %v", res.(bson.M))
		wait.Done()
	})
	wait.Wait()
}

func TestUpdateId(t *testing.T) {
	id := bson.NewObjectId()
	mongodb.Insert("mongodb_test", bson.M{"_id": id, "TestUpdateId": 1}, nil)

	wait.Add(1)
	mongodb.UpdateId("mongodb_test", id, bson.M{"$set": bson.M{"TestUpdateId": 2}}, func(res interface{}, err error) {
		checkRequest(t, err, res)
		wait.Done()
	})
	wait.Wait()

	wait.Add(1)
	mongodb.UpdateId("mongodb_test", bson.NewObjectId(), bson.M{"$set": bson.M{"TestUpdateId": 2}}, func(res interface{}, err error) {
		if err == nil {
			t.Errorf("should returns error")
		}
		checkRequest(t, err, res)
		wait.Done()
	})
	wait.Wait()
}

func TestUpdate(t *testing.T) {
	wait.Add(1)
	mongodb.Update("mongodb_test", bson.M{"TestFindId": 1}, bson.M{"$set": bson.M{"TestUpdate": 1}}, func(res interface{}, err error) {
		checkRequest(t, err, res)
		wait.Done()
	})
	wait.Wait()

	wait.Add(1)
	mongodb.FindOne("mongodb_test", bson.M{"TestUpdate": 1}, nil, func(res interface{}, err error) {
		checkRequest(t, err, res)
		t.Logf("Update: %v", res.(bson.M))
		wait.Done()
	})
	wait.Wait()
}

func TestUpdateAll(t *testing.T) {
	wait.Add(1)
	mongodb.UpdateAll("mongodb_test", bson.M{"c": 2}, bson.M{"$set": bson.M{"c": "3"}}, func(res interface{}, err error) {
		checkRequest(t, err, res)
		t.Logf("UpdateAll: %v", res.(int))
		wait.Done()
	})
	wait.Wait()
}

func TestUpsertId(t *testing.T) {
	wait.Add(1)
	id := bson.NewObjectId()
	mongodb.UpsertId("mongodb_test", id, bson.M{"TestUpsertId": 1}, func(res interface{}, err error) {
		checkRequest(t, err, res)
	})

	mongodb.UpsertId("mongodb_test", id, bson.M{"$set": bson.M{"TestUpdateId": 2}}, func(res interface{}, err error) {
		checkRequest(t, err, res)
		wait.Done()
	})
	wait.Wait()
}

func TestUpsert(t *testing.T) {
	wait.Add(1)

	mongodb.Upsert("mongodb_test", bson.M{"TestUpsert": 1}, bson.M{"TestUpsert": 1}, func(res interface{}, err error) {
		checkRequest(t, err, res)
	})

	mongodb.Upsert("mongodb_test", bson.M{"TestUpsert": 1}, bson.M{"TestUpsert": 1}, func(res interface{}, err error) {
		checkRequest(t, err, res)
		wait.Done()
	})

	wait.Wait()
}

func TestEnsureIndex(t *testing.T) {
	wait.Add(1)
	mongodb.EnsureIndex("mongodb_test", mgo.Index{
		Key: []string{"a"},
	}, func(res interface{}, err error) {
		checkRequest(t, err, res)
		wait.Done()
	})
	wait.Wait()
}

func TestEnsureIndexKey(t *testing.T) {
	wait.Add(1)
	mongodb.EnsureIndexKey("mongodb_test", []string{"a", "b", "c"}, func(res interface{}, err error) {
		checkRequest(t, err, res)
		wait.Done()
	})
	wait.Wait()
}

func TestDropIndex(t *testing.T) {
	wait.Add(1)
	mongodb.DropIndex("mongodb_test", []string{"a"}, func(res interface{}, err error) {
		checkRequest(t, err, res)
		wait.Done()
	})
	wait.Wait()
}

func TestRemoveId(t *testing.T) {
	var id interface{}
	wait.Add(1)

	mongodb.Upsert("mongodb_test", bson.M{"TestRemoveId": 1}, bson.M{"TestRemoveId": 1}, func(res interface{}, err error) {
		checkRequest(t, err, res)
		id = res
		mongodb.RemoveId("mongodb_test", id, func(res interface{}, err error) {
			checkRequest(t, err, res)
		})
		mongodb.RemoveId("mongodb_test", id, func(res interface{}, err error) {
			checkRequest(t, err, res)
			if err != mgo.ErrNotFound {
				t.Errorf("error should be not found")
			}
			wait.Done()
		})
	})
	wait.Wait()
}

func TestRemove(t *testing.T) {
	wait.Add(1)

	mongodb.Upsert("mongodb_test", bson.M{"TestRemove": 1}, bson.M{"TestRemove": 1}, func(res interface{}, err error) {
		checkRequest(t, err, res)
		mongodb.Remove("mongodb_test", bson.M{"TestRemove": 1}, func(res interface{}, err error) {
			checkRequest(t, err, res)
		})
		mongodb.Remove("mongodb_test", bson.M{"TestRemove": 1}, func(res interface{}, err error) {
			checkRequest(t, err, res)
			if err != mgo.ErrNotFound {
				t.Errorf("error should be not found")
			}
			wait.Done()
		})
	})
	wait.Wait()
}

func TestRemoveAll(t *testing.T) {
	wait.Add(1)

	mongodb.Insert("mongodb_test", bson.M{"TestRemove": 1}, nil)
	mongodb.Insert("mongodb_test", bson.M{"TestRemove": 1}, nil)
	mongodb.Insert("mongodb_test", bson.M{"TestRemove": 1}, func(res interface{}, err error) {
		checkRequest(t, err, res)
		mongodb.RemoveAll("mongodb_test", bson.M{"TestRemove": 1}, func(res interface{}, err error) {
			checkRequest(t, err, res)
			if res.(int) != 3 {
				t.Errorf("should remove 3 docs")
			}
		})
		mongodb.RemoveAll("mongodb_test", bson.M{"TestRemove": 1}, func(res interface{}, err error) {
			checkRequest(t, err, res)
			if res.(int) != 0 {
				t.Errorf("should remove 0 docs")
			}
			wait.Done()
		})
	})
	wait.Wait()
}

func TestDropCollection(t *testing.T) {
	wait.Add(1)
	mongodb.DropCollection("mongodb_test", func(res interface{}, err error) {
		checkRequest(t, err, res)
		wait.Done()
	})
	wait.Wait()
}

func checkRequest(t *testing.T, err error, res interface{}) {
	if err != nil {
		t.Logf("error = %v, res = %v", err, res)
	} else {
		t.Logf("success, res = %v", res)
	}
}

func init() {
	go func() {
		for {
			post.Tick()
			time.Sleep(time.Millisecond * 10)
		}
	}()
}
