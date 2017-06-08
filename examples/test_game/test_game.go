package main

import (
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/components/game"
)

func init() {

}

type gameDelegate struct {
	game.GameDelegate
}

func main() {
	goworld.SetSpaceDelegate(&SpaceDelegate{})
	goworld.RegisterEntity("OnlineService", &OnlineService{})
	goworld.RegisterEntity("Monster", &Monster{})
	goworld.RegisterEntity("Avatar", &Avatar{})

	goworld.Run(&gameDelegate{})
}

func (game gameDelegate) OnReady() {
	game.GameDelegate.OnReady()
	goworld.CreateEntity("OnlineService")
	goworld.CreateSpace()
}
