package game

import "github.com/xiaonanln/goworld/gwlog"

type IServerDelegate interface {
	OnServerReady()
}

type GameDelegate struct {
}

func (gd *GameDelegate) OnServerReady() {
	gwlog.Info("server %d is ready.", serverid)
}
