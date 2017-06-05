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
	goworld.RegisterEntity("TestEntity", &TestEntity{})
	goworld.Run(&gameDelegate{})
}

func (game gameDelegate) OnReady() {
	game.GameDelegate.OnReady()
	goworld.CreateEntity("TestEntity")
}
