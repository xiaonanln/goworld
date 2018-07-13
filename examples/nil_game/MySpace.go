package main

import (
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

// MySpace is the custom space type
type MySpace struct {
	goworld.Space // Space type should always inherit from entity.Space
}

// OnGameReady is called when the game server is ready
func (space *MySpace) OnGameReady() {
	gwlog.Infof("Game %d Is Ready", goworld.GetGameID())
}
