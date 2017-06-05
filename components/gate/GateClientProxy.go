package main

import (
	"fmt"
	"net"

	"github.com/xiaonanln/goworld/gwlog"
)

type GateClientProxy struct {
	conn net.Conn
}

func newGateClientProxy(conn net.Conn) *GateClientProxy {
	return &GateClientProxy{
		conn: conn,
	}
}

func (gcp *GateClientProxy) serve() {
	gwlog.Debug("Serving %s ...", gcp)
}
func (gcp *GateClientProxy) String() string {
	return fmt.Sprintf("GateClient<%s>", gcp.conn.RemoteAddr())
}
