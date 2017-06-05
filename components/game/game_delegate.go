package game

import "github.com/xiaonanln/goworld/gwlog"

type IGameDelegate interface {
	OnReady()
}

type GameDelegate struct {
}

func (gd *GameDelegate) OnReady() {
	gwlog.Info("game %d is ready.", gameid)
}
