package main

import (
	"fmt"

	"net"

	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/netutil"
)

type GateServer struct {
	config *config.GateConfig
}

func newGateServer(gatecfg *config.GateConfig) *GateServer {
	return &GateServer{
		config: gatecfg,
	}
}

func (gs *GateServer) run() {
	listenAddr := fmt.Sprintf("%s:%d", gs.config.Ip, gs.config.Port)
	netutil.ServeTCPForever(listenAddr, gs)
}

func (gs *GateServer) ServeTCPConnection(conn net.Conn) {
	clientProxy := newGateClientProxy(conn)
	clientProxy.serve()
}
