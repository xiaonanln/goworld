package main

import (
	"net"

	"fmt"

	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/gwioutil"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/netutil"
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
		dcp.owner.handleDispatcherClientDisconnect(dcp)
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

		// pass the packet to the dispatcher service
		if consts.DEBUG_PACKETS {
			gwlog.Debugf("%s.RecvPacket: msgtype=%v, payload=%v", dcp, msgtype, pkt.Payload())
		}

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

func (dcp *dispatcherClientProxy) beforeFlush() {
	// Collect all entity sync infos to this game before flush
	if dcp.gameid > 0 {
		entitySyncInfos := dcp.owner.popEntitySyncInfosToGame(dcp.gameid)
		if len(entitySyncInfos) > 0 {
			// send the entity sync infos to this game
			packet := netutil.NewPacket()
			packet.AppendUint16(proto.MT_SYNC_POSITION_YAW_FROM_CLIENT)
			packet.AppendBytes(entitySyncInfos)
			dcp.SendPacket(packet)
			packet.Release()
		}
	}
}
