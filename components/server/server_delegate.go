package server

import "github.com/xiaonanln/goworld/gwlog"

type IServerDelegate interface {
	OnReady()
}

type ServerDelegate struct {
}

func (gd *ServerDelegate) OnReady() {
	gwlog.Info("server %d is ready.", serverid)
}
