package main

import (
	"net"

	"fmt"

	"os"

	"time"

	"github.com/xiaonanln/goworld/engine/consts"
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
	//if consts.DISPATCHER_CLIENT_PROXY_BUFFERED_DELAY > 0 {
	//	conn = netutil.NewBufferedConnection(conn, consts.DISPATCHER_CLIENT_PROXY_BUFFERED_DELAY)
	//}
	gwc := proto.NewGoWorldConnection(netutil.NewBufferedReadConnection(conn), false)

	dcp := &dispatcherClientProxy{
		GoWorldConnection: gwc,
		owner:             owner,
	}
	return dcp
}

func (dcp *dispatcherClientProxy) startAutoFlush() {
	go func() {
		gwc := dcp.GoWorldConnection
		defer gwlog.Debugf("%s: auto flush routine quited", gwc)
		for !gwc.IsClosed() {
			time.Sleep(time.Millisecond * 10)
			dcp.beforeFlush()
			err := gwc.Flush("dispatcherClientProxy")
			if err != nil {
				break
			}
		}
	}()
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
			if netutil.IsTemporaryNetError(err) {
				continue
			}

			gwlog.Panic(err)
		}

		if consts.DEBUG_PACKETS {
			gwlog.Debugf("%s.RecvPacket: msgtype=%v, payload=%v", dcp, msgtype, pkt.Payload())
		}
		if msgtype == proto.MT_SYNC_POSITION_YAW_FROM_CLIENT {
			dcp.owner.handleSyncPositionYawFromClient(dcp, pkt)
		} else if msgtype == proto.MT_SYNC_POSITION_YAW_ON_CLIENTS {
			dcp.owner.handleSyncPositionYawOnClients(dcp, pkt)
		} else if msgtype == proto.MT_CALL_ENTITY_METHOD {
			dcp.owner.handleCallEntityMethod(dcp, pkt)
		} else if msgtype >= proto.MT_REDIRECT_TO_GATEPROXY_MSG_TYPE_START && msgtype <= proto.MT_REDIRECT_TO_GATEPROXY_MSG_TYPE_STOP {
			dcp.owner.handleDoSomethingOnSpecifiedClient(dcp, pkt)
		} else if msgtype == proto.MT_CALL_ENTITY_METHOD_FROM_CLIENT {
			dcp.owner.handleCallEntityMethodFromClient(dcp, pkt)
		} else if msgtype == proto.MT_MIGRATE_REQUEST {
			dcp.owner.handleMigrateRequest(dcp, pkt)
		} else if msgtype == proto.MT_REAL_MIGRATE {
			dcp.owner.handleRealMigrate(dcp, pkt)
		} else if msgtype == proto.MT_CALL_FILTERED_CLIENTS {
			dcp.owner.handleCallFilteredClientProxies(dcp, pkt)
		} else if msgtype == proto.MT_NOTIFY_CLIENT_CONNECTED {
			dcp.owner.handleNotifyClientConnected(dcp, pkt)
		} else if msgtype == proto.MT_NOTIFY_CLIENT_DISCONNECTED {
			dcp.owner.handleNotifyClientDisconnected(dcp, pkt)
		} else if msgtype == proto.MT_LOAD_ENTITY_ANYWHERE {
			dcp.owner.handleLoadEntityAnywhere(dcp, pkt)
		} else if msgtype == proto.MT_NOTIFY_CREATE_ENTITY {
			eid := pkt.ReadEntityID()
			dcp.owner.handleNotifyCreateEntity(dcp, pkt, eid)
		} else if msgtype == proto.MT_NOTIFY_DESTROY_ENTITY {
			eid := pkt.ReadEntityID()
			dcp.owner.handleNotifyDestroyEntity(dcp, pkt, eid)
		} else if msgtype == proto.MT_CREATE_ENTITY_ANYWHERE {
			dcp.owner.handleCreateEntityAnywhere(dcp, pkt)
		} else if msgtype == proto.MT_DECLARE_SERVICE {
			dcp.owner.handleDeclareService(dcp, pkt)
		} else if msgtype == proto.MT_SET_GAME_ID {
			// this is a game server
			gameid := pkt.ReadUint16()
			isReconnect := pkt.ReadBool()
			isRestore := pkt.ReadBool()
			if gameid <= 0 {
				gwlog.Panicf("invalid gameid: %d", gameid)
			}
			if dcp.gameid > 0 || dcp.gateid > 0 {
				gwlog.Panicf("already set gameid=%d, gateid=%d", dcp.gameid, dcp.gateid)
			}
			dcp.gameid = gameid
			dcp.startAutoFlush()
			dcp.owner.handleSetGameID(dcp, pkt, gameid, isReconnect, isRestore)
		} else if msgtype == proto.MT_SET_GATE_ID {
			// this is a gate
			gateid := pkt.ReadUint16()
			if gateid <= 0 {
				gwlog.Panicf("invalid gateid: %d", gateid)
			}
			if dcp.gameid > 0 || dcp.gateid > 0 {
				gwlog.Panicf("already set gameid=%d, gateid=%d", dcp.gameid, dcp.gateid)
			}
			dcp.gateid = gateid
			dcp.startAutoFlush()
			dcp.owner.handleSetGateID(dcp, pkt, gateid)
		} else if msgtype == proto.MT_START_FREEZE_GAME {
			// freeze the game
			dcp.owner.handleStartFreezeGame(dcp, pkt)
		} else {
			gwlog.TraceError("unknown msgtype %d from %s", msgtype, dcp)
			if consts.DEBUG_MODE {
				os.Exit(2)
			}
		}

		pkt.Release()
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
