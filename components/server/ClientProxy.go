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

	for {
		var msgtype proto.MsgType_t
		pkt, err := cp.Recv(&msgtype)
		if err != nil {
			panic(err)
		}

		entityID := pkt.ReadEntityID()
		method := pkt.ReadVarStr()
		var args []interface{}
		pkt.ReadMessage(&args)
		gwlog.Info("Recv %s %s %v", entityID, method, args)
	}
}
