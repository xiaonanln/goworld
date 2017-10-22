package entity

import (
	"testing"

	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

type TestEntity struct {
	Entity
	TestComponent
}

type TestComponent struct {
	Component
}

//func init() {
//	gwlog.Panicf("should not goes here")
//}

func (e *TestEntity) OnInit() {
	gwlog.Infof("%s.OnInit ...", e)
}

func (e *TestEntity) OnCreated() {
	gwlog.Infof("%s.OnCreated ...", e)
}

func (e *TestEntity) OnMigrateIn() {
	gwlog.Infof("%s.OnMigrateIn ...", e)
}

func (c *TestComponent) OnInit() {
	gwlog.Infof("TestComponent.OnInit ,,,")
}

func (c *TestComponent) OnMigrateIn() {
	gwlog.Infof("TestComponent.OnMigrateIn ...")
}

func TestRegisterEntity(t *testing.T) {
	RegisterEntity("TestEntity", &TestEntity{}, false, false)
}

func TestGenEntityID(t *testing.T) {
	eid := common.GenEntityID()
	gwlog.Infof("TestGenEntityID: %s", eid)
}

func TestEntityModule(t *testing.T) {
	eid := createEntity("TestEntity", nil, Vector3{}, "", nil, nil, nil, ccMigrate)
	e := GetEntity(eid)
	te := e.I.(*TestEntity)
	t.Logf("Created entity: %s => %s", eid, te)
}

func TestTestComponent(t *testing.T) {

}
