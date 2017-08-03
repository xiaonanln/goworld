package game

import "github.com/xiaonanln/goworld/engine/gwlog"

type IGameDelegate interface {
	OnGameReady()
}

type GameDelegate struct {
}

func (gd *GameDelegate) OnGameReady() {
	gwlog.Info("game %d is ready.", gameid)
}
