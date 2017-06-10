package main

import (
	"net"

	"fmt"

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
		if msgtype == proto.MT_CALL_ENTITY_METHOD {
			eid := pkt.ReadEntityID()
			method := pkt.ReadVarStr()
			dcp.owner.HandleCallEntityMethod(dcp, pkt, eid, method)
		} else if msgtype == proto.MT_NOTIFY_CREATE_ENTITY {
			eid := pkt.ReadEntityID()
			dcp.owner.HandleNotifyCreateEntity(dcp, pkt, eid)
		} else if msgtype == proto.MT_CREATE_ENTITY_ANYWHERE {
			typeName := pkt.ReadVarStr()
			dcp.owner.HandleCreateEntityAnywhere(dcp, pkt, typeName)
		} else if msgtype == proto.MT_DECLARE_SERVICE {
			eid := pkt.ReadEntityID()
			dcp.owner.HandleDeclareService(dcp, pkt, eid)
		} else if msgtype == proto.MT_SET_GAME_ID {
			gameid := int(pkt.ReadUint16())
			dcp.gameid = gameid
			dcp.owner.HandleSetGameID(dcp, pkt, gameid)
		} else {
			gwlog.TraceError("unknown msgtype %d from %s", msgtype, dcp)
		}
	}
}

func (dcp *DispatcherClientProxy) String() string {
	return fmt.Sprintf("DispatcherClientProxy<%d|%s>", dcp.gameid, dcp.RemoteAddr())
}
