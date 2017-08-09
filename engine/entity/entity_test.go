package entity

import (
	"testing"

	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

type TestEntity struct {
	Entity
}

//func init() {
//	gwlog.Panicf("should not goes here")
//}

func TestRegisterEntity(t *testing.T) {
	RegisterEntity("TestEntity", &TestEntity{}, false, false)
}

func TestGenEntityID(t *testing.T) {
	eid := common.GenEntityID()
	gwlog.Infof("TestGenEntityID: %s", eid)
}

func TestEntityManager(t *testing.T) {

}
