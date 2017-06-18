package main

import (
	"fmt"

	"net"

	"sync"

	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/netutil"
	"github.com/xiaonanln/goworld/proto"
)

type DispatcherService struct {
	sync.RWMutex

	config            *config.DispatcherConfig
	clients           []*DispatcherClientProxy
	chooseClientIndex int

	entityLocs           map[common.EntityID]uint16
	registeredServices   map[string]common.EntityID
	targetServerOfClient map[common.ClientID]uint16
}

func newDispatcherService() *DispatcherService {
	cfg := config.Get()
	serverCount := len(cfg.Servers)
	return &DispatcherService{
		config:            &cfg.Dispatcher,
		clients:           make([]*DispatcherClientProxy, serverCount),
		chooseClientIndex: 0,

		entityLocs:           map[common.EntityID]uint16{},
		registeredServices:   map[string]common.EntityID{},
		targetServerOfClient: map[common.ClientID]uint16{},
	}
}

func (service *DispatcherService) String() string {
	return fmt.Sprintf("DispatcherService<C%d|E%d>", len(service.clients), len(service.entityLocs))
}

func (service *DispatcherService) run() {
	host := fmt.Sprintf("%s:%d", service.config.Ip, service.config.Port)
	netutil.ServeTCPForever(host, service)
}

func (service *DispatcherService) ServeTCPConnection(conn net.Conn) {
	client := newDispatcherClientProxy(service, conn)
	client.serve()
}

func (service *DispatcherService) HandleSetServerID(dcp *DispatcherClientProxy, pkt *netutil.Packet, serverid uint16) {
	gwlog.Debug("%s.HandleSetServerID: dcp=%s, serverid=%d", service, dcp, serverid)
	if serverid <= 0 {
		gwlog.Panicf("invalid serverid: %d", serverid)
	}

	olddcp := service.clients[serverid-1] // should be nil, unless reconnect
	service.clients[serverid-1] = dcp
	// notify all servers that all servers connected to dispatcher now!
	if service.isAllClientsConnected() {
		pkt.ClearPayload() // reuse this packet
		pkt.AppendUint16(proto.MT_NOTIFY_ALL_SERVERS_CONNECTED)
		if olddcp == nil {
			// for the first time that all servers connected to dispatcher, notify all servers
			gwlog.Info("All servers(%d) are connected", len(service.clients))
			service.broadcastToDispatcherClients(pkt)
		} else {
			dcp.SendPacket(pkt)
		}
	}
	pkt.Release()
	return
}

func (service *DispatcherService) isAllClientsConnected() bool {
	for _, client := range service.clients {
		if client == nil {
			return false
		}
	}
	return true
}

func (service *DispatcherService) dispatcherClientOfServer(serverid uint16) *DispatcherClientProxy {
	return service.clients[serverid-1]
}

// Choose a dispatcher client for sending Anywhere packets
func (service *DispatcherService) chooseDispatcherClient() *DispatcherClientProxy {
	client := service.clients[service.chooseClientIndex]
	service.chooseClientIndex = (service.chooseClientIndex + 1) % len(service.clients)
	return client
	//startIndex := service.chooseClientIndex
	//clients := service.clients
	//clientsNum := len(clients)
	//if clients[startIndex] != nil { // most of time
	//	service.chooseClientIndex = (service.chooseClientIndex + 1) % clientsNum
	//	return clients[startIndex]
	//} else {
	//	index := (startIndex + 1) % clientsNum
	//	for index != startIndex {
	//		if clients[index] != nil {
	//			service.chooseClientIndex = (index + 1) % clientsNum
	//			return clients[index]
	//		}
	//		index = (index + 1) % clientsNum
	//	}
	//
	//	return nil // non-nil client not found, should not happen
	//}
}

//
//func (service *DispatcherService) HandleDispatcherClientDisconnect(dcp *DispatcherClientProxy) {
//	gwlog.Panic(service, dcp)
//	service.Lock()
//	sid := dcp.serverid
//	if service.clients[sid] != dcp {
//		// should never happen
//		service.Unlock()
//		return
//	}
//	service.clients[sid] = nil
//	remove := entity.EntityIDSet{}
//	for eid, loc := range service.entityLocs {
//		if loc == sid {
//			remove.Add(eid)
//		}
//	}
//
//	for eid := range remove {
//		service.entityLocs[eid]
//	}
//}

// Entity is create on the target server
func (service *DispatcherService) HandleNotifyCreateEntity(dcp *DispatcherClientProxy, pkt *netutil.Packet, entityID common.EntityID) {
	gwlog.Debug("%s.HandleNotifyCreateEntity: dcp=%s, entityID=%s", service, dcp, entityID)
	service.Lock()
	service.entityLocs[entityID] = dcp.serverid
	service.Unlock()
	pkt.Release()
}

func (service *DispatcherService) HandleNotifyDestroyEntity(dcp *DispatcherClientProxy, pkt *netutil.Packet, entityID common.EntityID) {
	gwlog.Debug("%s.HandleNotifyDestroyEntity: dcp=%s, entityID=%s", service, dcp, entityID)
	service.Lock()
	delete(service.entityLocs, entityID)
	service.Unlock()
	pkt.Release()
}

func (service *DispatcherService) HandleNotifyClientConnected(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	//clientid := pkt.ReadClientID()
	pkt.AppendUint16(dcp.serverid)
	service.chooseDispatcherClient().SendPacketRelease(pkt)
}

func (service *DispatcherService) HandleLoadEntityAnywhere(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	//typeName := pkt.ReadVarStr()
	//eid := pkt.ReadEntityID()
	gwlog.Debug("%s.HandleLoadEntityAnywhere: dcp=%s, pkt=%v", service, dcp, pkt.Payload())
	eid := pkt.ReadEntityID() // field 1
	service.Lock()
	sid := service.entityLocs[eid]

	if sid == 0 { // entity not loaded, try load now
		dcp := service.chooseDispatcherClient()
		service.entityLocs[eid] = dcp.serverid
		service.Unlock()
		dcp.SendPacketRelease(pkt)
	} else {
		service.Unlock()
	}
}

func (service *DispatcherService) HandleCreateEntityAnywhere(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	gwlog.Debug("%s.HandleCreateEntityAnywhere: dcp=%s, pkt=%s", service, dcp, pkt.Payload())
	service.chooseDispatcherClient().SendPacketRelease(pkt)
}

func (service *DispatcherService) HandleDeclareService(dcp *DispatcherClientProxy, pkt *netutil.Packet, entityID common.EntityID) {
	gwlog.Debug("%s.HandleDeclareService: dcp=%s, entityID=%s", service, dcp, entityID)
	service.Lock()
	service.entityLocs[entityID] = dcp.serverid
	service.Unlock()
	service.broadcastToDispatcherClients(pkt)
	pkt.Release()
	//_, ok := service.registeredServices[serviceName]
	//if ok {
	//	// already registered
	//	dcp.SendDeclareServiceReply(entityID, serviceName, false)
	//	return
	//}
	//service.registeredServices[serviceName] = entityID
	//dcp.SendDeclareServiceReply(entityID, serviceName, true)
}

func (service *DispatcherService) HandleCallEntityMethod(dcp *DispatcherClientProxy, pkt *netutil.Packet, entityID common.EntityID, method string) {
	gwlog.Debug("%s.HandleCallEntityMethod: dcp=%s, entityID=%s, method=%s", service, dcp, entityID, method)

	service.RLock()
	serverid := service.entityLocs[entityID]
	service.RUnlock()
	if serverid == 0 {
		// server not found
		gwlog.Warn("Entity %s not found when calling method %s", entityID, method)
		return
	}

	service.dispatcherClientOfServer(serverid).SendPacketRelease(pkt)
}

func (service *DispatcherService) HandleCallEntityMethodFromClient(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	entityid := pkt.ReadEntityID()
	method := pkt.ReadVarStr()
	var args []interface{}
	pkt.ReadData(&args)
	clientid := pkt.ReadClientID()
	sid := service.targetServerOfClient[clientid]
	gwlog.Info("%s.HandleCallEntityMethodFromClient: %s.%s %v, clientid=%s, sid=%d", service, entityid, method, args, clientid, sid)
	service.dispatcherClientOfServer(sid).SendPacketRelease(pkt)
}

func (service *DispatcherService) HandleCreateEntityOnClient(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	sid := pkt.ReadUint16()
	clientid := pkt.ReadClientID()
	service.targetServerOfClient[clientid] = sid
	// Server <sid> is creating entity on client <clientid>, so we can safely assumes that target entity of
	service.dispatcherClientOfServer(sid).SendPacketRelease(pkt)
}

func (service *DispatcherService) HandleDestroyEntityOnClient(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	sid := pkt.ReadUint16()
	service.dispatcherClientOfServer(sid).SendPacketRelease(pkt)
}

func (service *DispatcherService) HandleNotifyAttrChangeOnClient(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	sid := pkt.ReadUint16()
	service.dispatcherClientOfServer(sid).SendPacketRelease(pkt)
}

func (service *DispatcherService) broadcastToDispatcherClients(pkt *netutil.Packet) {
	for _, dcp := range service.clients {
		dcp.SendPacket(pkt)
	}
}
