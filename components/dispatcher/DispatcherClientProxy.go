package main

import (
	"net"

	"fmt"

	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/netutil"
	"github.com/xiaonanln/goworld/proto"
)

type DispatcherClientProxy struct {
	proto.GoWorldConnection
	owner  *DispatcherService
	gameid int
}

func newDispatcherClientProxy(owner *DispatcherService, conn net.Conn) *DispatcherClientProxy {
	return &DispatcherClientProxy{
		GoWorldConnection: proto.NewGoWorldConnection(conn),
		owner:             owner,
	}
}

func (dcp *DispatcherClientProxy) serve() {
	// Serve the dispatcher client from game / gate
	defer func() {
		dcp.Close()

		err := recover()
		if err != nil && !netutil.IsConnectionClosed(err) {
			gwlog.Error("Client %s paniced with error: %v", dcp, err)
		}
	}()

	gwlog.Info("New dispatcher client: %s", dcp)
	for {
		var msgtype proto.MsgType_t
		pkt, err := dcp.Recv(&msgtype)
		if err != nil {
			gwlog.Panic(err)
		}

		gwlog.Info("%s.RecvPacket: msgtype=%v, payload=%v", dcp, msgtype, pkt.Payload())
		if msgtype == proto.MT_SET_GAME_ID {
			gameid := int(pkt.ReadUint16())
			dcp.gameid = gameid
			dcp.owner.HandleSetGameID(dcp, gameid)
		} else if msgtype == proto.MT_NOTIFY_CREATE_ENTITY {
			eid := entity.EntityID(pkt.ReadBytes(entity.ENTITYID_LENGTH))
			dcp.owner.HandleNotifyCreateEntity(dcp, eid)
		}
	}
}

func (dcp *DispatcherClientProxy) String() string {
	return fmt.Sprintf("DispatcherClientProxy<%d|%s>", dcp.gameid, dcp.RemoteAddr())
}
