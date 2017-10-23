package entity

import (
	"testing"

	"reflect"

	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

type TestEntity struct {
	Entity
	TestComponent1
	//TestComponent2
}

type TestComponent1 struct {
	Component
}

type TestComponent2 struct {
	Component
}

type TestComponent3 struct {
	Component
}

type TestEntityD struct {
	TestEntity
	//TestComponent3
}

//func init() {
//	gwlog.Panicf("should not goes here")
//}

func (e *TestEntity) OnInit() {
	gwlog.Infof("TestEntity.OnInit ...")
}

func (e *TestEntity) OnCreated() {
	gwlog.Infof("%s.OnCreated ...", e)
}

func (e *TestEntity) OnMigrateIn() {
	gwlog.Infof("%s.OnMigrateIn ...", e)
}

func (c *TestComponent1) OnInit() {
	gwlog.Infof("TestComponent1.OnInit ,,,")
}

func (c *TestComponent1) OnMigrateIn() {
	gwlog.Infof("TestComponent1.OnMigrateIn ...")
}

func (c *TestComponent2) OnInit() {
	gwlog.Infof("TestComponent2.OnInit ,,,")
}

func (c *TestComponent3) OnInit() {
	gwlog.Infof("TestComponent3.OnInit ,,,")
}

func TestRegisterEntity(t *testing.T) {
	RegisterEntity("TestEntity", &TestEntity{}, false, false)
}

func TestGenEntityID(t *testing.T) {
	eid := common.GenEntityID()
	gwlog.Infof("TestGenEntityID: %s", eid)
}

func TestEntityModule(t *testing.T) {
	RegisterEntity("TestEntityD", &TestEntityD{}, false, false)
	//eid := createEntity("TestEntityD", nil, Vector3{}, "", nil, nil, nil, ccMigrate)
	//e := GetEntity(eid)
	//te := e.V.Interface().(*TestEntityD)
	//onInitFunc, ok := e.V.Type().MethodByName("OnInit")
	reflect.ValueOf(&Entity{}).MethodByName("OnInit").Call([]reflect.Value{})
	reflect.ValueOf(&TestEntity{}).MethodByName("OnInit").Call([]reflect.Value{})
	reflect.ValueOf(&TestEntityD{}).MethodByName("OnInit").Call([]reflect.Value{})

	//te.OnInit()
	//t.Logf("Created entity: %s => %s, %v, %v", eid, te, onInitFunc1, ok)
}

func TestTestComponent(t *testing.T) {

}
