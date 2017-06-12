package server

import (
	"fmt"

	"net"

	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/vacuum/netutil"
)

type GateService struct {
	listenAddr string
}

func newGateService() *GateService {
	return &GateService{}
}

func (gs *GateService) run() {
	cfg := config.GetServer(serverid)
	gs.listenAddr = fmt.Sprintf("%s:%d", cfg.Ip, cfg.Port)
	netutil.ServeTCPForever(gs.listenAddr, gs)
}

func (gs *GateService) String() string {
	return fmt.Sprintf("GateService<%s>", gs.listenAddr)
}

func (gs *GateService) ServeTCPConnection(conn net.Conn) {
	cp := newClientProxy(conn)
	if consts.DEBUG_CLIENTS {
		gwlog.Debug("%s.ServeTCPConnection: new client %s", gs, cp)
	}
	cp.serve()
}
