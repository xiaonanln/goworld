package main

import (
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/components/game"
	"github.com/xiaonanln/goworld/entity"
)

type TestEntity struct {
	entity.Entity
}

func init() {

}

type gameDelegate struct {
	game.GameDelegate
}

func main() {
	goworld.SetSpaceDelegate(&SpaceDelegate{})
	goworld.RegisterEntity("TestEntity", &TestEntity{})
	goworld.Run(&gameDelegate{})
}

func (game gameDelegate) OnReady() {
	game.GameDelegate.OnReady()
	// create the space
	goworld.CreateSpace()
	//eid1 := goworld.createEntity("TestEntity")
	//eid2 := goworld.createEntity("TestEntity")
}

type SpaceDelegate struct {
	entity.DefaultSpaceDelegate // override from default space delegate
}

func (delegate *SpaceDelegate) OnSpaceCreated(space *entity.Space) {
	delegate.DefaultSpaceDelegate.OnSpaceCreated(space)

	space.CreateEntity("TestEntity")
}
