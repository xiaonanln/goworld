package server

import (
	"fmt"

	"net"

	"sync"

	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/components/dispatcher/dispatcher_client"
	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/netutil"
	"github.com/xiaonanln/goworld/proto"
)

type GateService struct {
	listenAddr        string
	clientProxies     map[common.ClientID]*ClientProxy
	clientProxiesLock sync.RWMutex
	//packetQueue chan packetQueueItem
}

func newGateService() *GateService {
	return &GateService{
		//packetQueue: make(chan packetQueueItem, consts.DISPATCHER_CLIENT_PACKET_QUEUE_SIZE),
		clientProxies: map[common.ClientID]*ClientProxy{},
	}
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
	gs.clientProxiesLock.Lock()
	gs.clientProxies[cp.clientid] = cp
	gs.clientProxiesLock.Unlock()
	dispatcher_client.GetDispatcherClientForSend().SendNotifyClientConnected(cp.clientid)
	if consts.DEBUG_CLIENTS {
		gwlog.Debug("%s.ServeTCPConnection: client %s connected", gs, cp)
	}
	cp.serve()
}

func (gs *GateService) onClientProxyClose(cp *ClientProxy) {
	gs.clientProxiesLock.Lock()
	delete(gs.clientProxies, cp.clientid)
	gs.clientProxiesLock.Unlock()

	dispatcher_client.GetDispatcherClientForSend().SendNotifyClientDisconnected(cp.clientid)
	if consts.DEBUG_CLIENTS {
		gwlog.Debug("%s.onClientProxyClose: client %s disconnected", gs, cp)
	}
}

func (gs *GateService) HandleDispatcherClientPacket(msgtype proto.MsgType_t, packet *netutil.Packet) {
	if consts.DEBUG_PACKETS {
		gwlog.Debug("%s.HandleDispatcherClientPacket: msgtype=%v, packet=%v", gs, msgtype, packet.Payload())
	}
	_ = packet.ReadUint16() // sid
	clientid := packet.ReadClientID()
	gs.clientProxiesLock.RLock()
	clientproxy := gs.clientProxies[clientid]
	gs.clientProxiesLock.RUnlock()
	if clientproxy != nil {
		clientproxy.SendPacketRelease(packet)
	}

	//typeName := packet.ReadVarStr()
	//entityid := packet.ReadEntityID()

}
