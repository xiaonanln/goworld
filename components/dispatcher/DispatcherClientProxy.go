package main

import (
	"net"

	"fmt"

	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/netutil"
	"github.com/xiaonanln/goworld/proto"
)

type DispatcherClientProxy struct {
	proto.GoWorldConnection
	owner    *DispatcherService
	serverid uint16
}

func newDispatcherClientProxy(owner *DispatcherService, conn net.Conn) *DispatcherClientProxy {
	return &DispatcherClientProxy{
		GoWorldConnection: proto.NewGoWorldConnection(conn, false), // Using send queue slows down dispatcher performance
		owner:             owner,
	}
}

func (dcp *DispatcherClientProxy) serve() {
	// Serve the dispatcher client from server / gate
	defer func() {
		dcp.Close()
		dcp.owner.HandleDispatcherClientDisconnect(dcp)
		err := recover()
		if err != nil && !netutil.IsConnectionClosed(err) {
			gwlog.TraceError("Client %s paniced with error: %v", dcp, err)
		}
	}()

	gwlog.Info("New dispatcher client: %s", dcp)
	for {
		var msgtype proto.MsgType_t
		pkt, err := dcp.Recv(&msgtype)
		if err != nil {
			gwlog.Panic(err)
		}

		if consts.DEBUG_PACKETS {
			gwlog.Debug("%s.RecvPacket: msgtype=%v, payload=%v", dcp, msgtype, pkt.Payload())
		}

		if msgtype == proto.MT_CALL_ENTITY_METHOD {
			dcp.owner.HandleCallEntityMethod(dcp, pkt)
		} else if msgtype == proto.MT_CREATE_ENTITY_ON_CLIENT {
			dcp.owner.HandleCreateEntityOnClient(dcp, pkt)
		} else if msgtype == proto.MT_DESTROY_ENTITY_ON_CLIENT {
			dcp.owner.HandleDestroyEntityOnClient(dcp, pkt)
		} else if msgtype == proto.MT_NOTIFY_ATTR_CHANGE_ON_CLIENT {
			dcp.owner.HandleNotifyAttrChangeOnClient(dcp, pkt)
		} else if msgtype == proto.MT_NOTIFY_ATTR_DEL_ON_CLIENT {
			dcp.owner.HandleNotifyAttrDelOnClient(dcp, pkt)
		} else if msgtype == proto.MT_CALL_ENTITY_METHOD_FROM_CLIENT {
			dcp.owner.HandleCallEntityMethodFromClient(dcp, pkt)
		} else if msgtype == proto.MT_MIGRATE_REQUEST {
			dcp.owner.HandleMigrateRequest(dcp, pkt)
		} else if msgtype == proto.MT_REAL_MIGRATE {
			dcp.owner.HandleRealMigrate(dcp, pkt)
		} else if msgtype == proto.MT_NOTIFY_CLIENT_CONNECTED {
			dcp.owner.HandleNotifyClientConnected(dcp, pkt)
		} else if msgtype == proto.MT_NOTIFY_CLIENT_DISCONNECTED {
			dcp.owner.HandleNotifyClientDisconnected(dcp, pkt)
		} else if msgtype == proto.MT_LOAD_ENTITY_ANYWHERE {
			dcp.owner.HandleLoadEntityAnywhere(dcp, pkt)
		} else if msgtype == proto.MT_NOTIFY_CREATE_ENTITY {
			eid := pkt.ReadEntityID()
			dcp.owner.HandleNotifyCreateEntity(dcp, pkt, eid)
		} else if msgtype == proto.MT_NOTIFY_DESTROY_ENTITY {
			eid := pkt.ReadEntityID()
			dcp.owner.HandleNotifyDestroyEntity(dcp, pkt, eid)
		} else if msgtype == proto.MT_CREATE_ENTITY_ANYWHERE {
			dcp.owner.HandleCreateEntityAnywhere(dcp, pkt)
		} else if msgtype == proto.MT_DECLARE_SERVICE {
			dcp.owner.HandleDeclareService(dcp, pkt)
		} else if msgtype == proto.MT_SET_SERVER_ID {
			serverid := pkt.ReadUint16()
			isReconnect := pkt.ReadBool()
			dcp.serverid = serverid
			dcp.owner.HandleSetServerID(dcp, pkt, serverid, isReconnect)
		} else {
			gwlog.TraceError("unknown msgtype %d from %s", msgtype, dcp)
		}
	}
}

func (dcp *DispatcherClientProxy) String() string {
	return fmt.Sprintf("DispatcherClientProxy<%d|%s>", dcp.serverid, dcp.RemoteAddr())
}
