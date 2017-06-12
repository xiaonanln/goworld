package server

import (
	"net"

	"fmt"

	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/proto"
	"github.com/xiaonanln/vacuum/netutil"
)

type ClientProxy struct {
	proto.GoWorldConnection
}

func newClientProxy(conn net.Conn) *ClientProxy {
	return &ClientProxy{
		GoWorldConnection: proto.NewGoWorldConnection(conn),
	}
}

func (cp *ClientProxy) String() string {
	return fmt.Sprintf("ClientProxy<%s>", cp.RemoteAddr())
}

func (cp *ClientProxy) serve() {
	defer func() {
		cp.Close()
		if err := recover(); err != nil && !netutil.IsConnectionClosed(err) {
			gwlog.Error("%s error: %s", cp, err)
		} else {
			gwlog.Info("%s disconnected", cp)
		}
	}()

}
