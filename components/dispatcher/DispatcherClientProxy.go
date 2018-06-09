package main

import (
	"net"

	"fmt"

	"github.com/xiaonanln/goworld/engine/gwioutil"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/netutil"
	"github.com/xiaonanln/goworld/engine/post"
	"github.com/xiaonanln/goworld/engine/proto"
)

type dispatcherClientProxy struct {
	*proto.GoWorldConnection
	owner  *DispatcherService
	gameid uint16
	gateid uint16
}

func newDispatcherClientProxy(owner *DispatcherService, _conn net.Conn) *dispatcherClientProxy {
	conn := netutil.NetConnection{_conn}
	gwc := proto.NewGoWorldConnection(netutil.NewBufferedConnection(conn), false, "")

	dcp := &dispatcherClientProxy{
		GoWorldConnection: gwc,
		owner:             owner,
	}
	return dcp
}

func (dcp *dispatcherClientProxy) serve() {
	// Serve the dispatcher client from server / gate
	defer func() {
		dcp.Close()
		post.Post(func() {
			dcp.owner.handleDispatcherClientDisconnect(dcp)
		})
		err := recover()
		if err != nil && !netutil.IsConnectionError(err) {
			gwlog.TraceError("Client %s paniced with error: %v", dcp, err)
		}
	}()

	gwlog.Infof("New dispatcher client: %s", dcp)
	for {
		var msgtype proto.MsgType
		pkt, err := dcp.Recv(&msgtype)

		if err != nil {
			if gwioutil.IsTimeoutError(err) {
				continue
			} else if netutil.IsConnectionError(err) {
				break
			}

			gwlog.Panic(err)
		}

		//
		//if consts.DEBUG_PACKETS {
		//	gwlog.Debugf("%s.RecvPacket: msgtype=%v, payload=%v", dcp, msgtype, pkt.Payload())
		//}

		// pass the packet to the dispatcher service
		dcp.owner.messageQueue <- dispatcherMessage{dcp, proto.Message{msgtype, pkt}}
	}
}

func (dcp *dispatcherClientProxy) String() string {
	if dcp.gameid > 0 {
		return fmt.Sprintf("dispatcherClientProxy<game%d|%s>", dcp.gameid, dcp.RemoteAddr())
	} else if dcp.gateid > 0 {
		return fmt.Sprintf("dispatcherClientProxy<gate%d|%s>", dcp.gateid, dcp.RemoteAddr())
	} else {
		return fmt.Sprintf("dispatcherClientProxy<%s>", dcp.RemoteAddr())
	}
}
