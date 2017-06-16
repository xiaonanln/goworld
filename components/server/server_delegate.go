package server

import "github.com/xiaonanln/goworld/gwlog"

type IServerDelegate interface {
	OnServerReady()
}

type ServerDelegate struct {
}

func (gd *ServerDelegate) OnServerReady() {
	gwlog.Info("server %d is ready.", serverid)
}
