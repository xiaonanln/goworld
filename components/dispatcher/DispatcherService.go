package main

import (
	"fmt"

	"net"

	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/netutil"
)

type DispatcherService struct {
	config            *config.DispatcherConfig
	clients           []*DispatcherClientProxy
	chooseClientIndex int

	entityLocs           map[common.EntityID]uint16
	registeredServices   map[string]common.EntityID
	targetServerOfClient map[common.ClientID]uint16
}

func newDispatcherService(cfg *config.DispatcherConfig) *DispatcherService {
	return &DispatcherService{
		config:               cfg,
		clients:              []*DispatcherClientProxy{},
		chooseClientIndex:    0,
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

	for serverid > uint16(len(service.clients)) {
		service.clients = append(service.clients, nil)
	}
	service.clients[serverid-1] = dcp
	pkt.Release()
	return
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

// Entity is create on the target server
func (service *DispatcherService) HandleNotifyCreateEntity(dcp *DispatcherClientProxy, pkt *netutil.Packet, entityID common.EntityID) {
	gwlog.Debug("%s.HandleNotifyCreateEntity: dcp=%s, entityID=%s", service, dcp, entityID)
	service.entityLocs[entityID] = dcp.serverid
	pkt.Release()
}

func (service *DispatcherService) HandleNotifyDestroyEntity(dcp *DispatcherClientProxy, pkt *netutil.Packet, entityID common.EntityID) {
	gwlog.Debug("%s.HandleNotifyDestroyEntity: dcp=%s, entityID=%s", service, dcp, entityID)
	delete(service.entityLocs, entityID)
	pkt.Release()
}

func (service *DispatcherService) HandleNotifyClientConnected(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	// Client connected at one server, create boot entity in any server, but TODO: handle reconnect
	//clientid := pkt.ReadClientID()
	pkt.AppendUint16(dcp.serverid)
	service.chooseDispatcherClient().SendPacketRelease(pkt)
}

func (service *DispatcherService) HandleLoadEntityAnywhere(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	//typeName := pkt.ReadVarStr()
	//eid := pkt.ReadEntityID()
	gwlog.Debug("%s.HandleLoadEntityAnywhere: dcp=%s, pkt=%v", service, dcp, pkt.Payload())
	// TODO: check for entity loc and make sure that it's not loaded before
	service.chooseDispatcherClient().SendPacketRelease(pkt)
}

func (service *DispatcherService) HandleCreateEntityAnywhere(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	gwlog.Debug("%s.HandleCreateEntityAnywhere: dcp=%s, pkt=%s", service, dcp, pkt.Payload())
	service.chooseDispatcherClient().SendPacketRelease(pkt)
}

func (service *DispatcherService) HandleDeclareService(dcp *DispatcherClientProxy, pkt *netutil.Packet, entityID common.EntityID) {
	gwlog.Debug("%s.HandleDeclareService: dcp=%s, entityID=%s", service, dcp, entityID)
	service.entityLocs[entityID] = dcp.serverid
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

	serverid := service.entityLocs[entityID]
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
	pkt.ReadMessage(&args)
	clientid := pkt.ReadClientID() // TODO: optimize packet structure
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

func (service *DispatcherService) broadcastToDispatcherClients(pkt *netutil.Packet) {
	for _, dcp := range service.clients {
		if dcp != nil {
			dcp.SendPacket(pkt)
		}
	}
}
