package entity

import (
	"testing"

	"github.com/xiaonanln/goworld/gwlog"
)

type TestEntity struct {
	Entity
}

//func init() {
//	gwlog.Panicf("should not goes here")
//}

func TestRegisterEntity(t *testing.T) {
	RegisterEntity("TestEntity", &TestEntity{})
}

func TestGenEntityID(t *testing.T) {
	eid := GenEntityID()
	gwlog.Info("TestGenEntityID: %s", eid)
}

func TestEntityManager(t *testing.T) {

}
