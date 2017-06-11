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

	entityLocs         map[common.EntityID]int
	registeredServices map[string]common.EntityID
}

func newDispatcherService(cfg *config.DispatcherConfig) *DispatcherService {
	return &DispatcherService{
		config:             cfg,
		clients:            []*DispatcherClientProxy{},
		chooseClientIndex:  0,
		entityLocs:         map[common.EntityID]int{},
		registeredServices: map[string]common.EntityID{},
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

func (service *DispatcherService) HandleSetGameID(dcp *DispatcherClientProxy, pkt *netutil.Packet, gameid int) {
	gwlog.Debug("%s.HandleSetGameID: dcp=%s, gameid=%d", service, dcp, gameid)
	if gameid <= 0 {
		gwlog.Panicf("invalid gameid: %d", gameid)
	}

	for gameid > len(service.clients) {
		service.clients = append(service.clients, nil)
	}
	service.clients[gameid-1] = dcp
	pkt.Release()
	return
}

func (service *DispatcherService) dispatcherClientOfGame(gameid int) *DispatcherClientProxy {
	return service.clients[gameid-1]
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

// Entity is create on the target game
func (service *DispatcherService) HandleNotifyCreateEntity(dcp *DispatcherClientProxy, pkt *netutil.Packet, entityID common.EntityID) {
	gwlog.Debug("%s.HandleNotifyCreateEntity: dcp=%s, entityID=%s", service, dcp, entityID)
	service.entityLocs[entityID] = dcp.gameid
	pkt.Release()
}

func (service *DispatcherService) HandleLoadEntityAnywhere(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	//typeName := pkt.ReadVarStr()
	//eid := pkt.ReadEntityID()
	gwlog.Debug("%s.HandleLoadEntityAnywhere: dcp=%s, pkt=%v", service, dcp, pkt.Payload())
	service.chooseDispatcherClient().SendPacketRelease(pkt)
}

func (service *DispatcherService) HandleCreateEntityAnywhere(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	gwlog.Debug("%s.HandleCreateEntityAnywhere: dcp=%s, pkt=%s", service, dcp, pkt.Payload())
	service.chooseDispatcherClient().SendPacketRelease(pkt)
}

func (service *DispatcherService) HandleDeclareService(dcp *DispatcherClientProxy, pkt *netutil.Packet, entityID common.EntityID) {
	gwlog.Debug("%s.HandleDeclareService: dcp=%s, entityID=%s", service, dcp, entityID)
	service.entityLocs[entityID] = dcp.gameid
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

	gameid := service.entityLocs[entityID]
	if gameid == 0 {
		// game not found
		gwlog.Warn("Entity %s not found when calling method %s", entityID, method)
		return
	}

	service.dispatcherClientOfGame(gameid).SendPacketRelease(pkt)
}

func (service *DispatcherService) broadcastToDispatcherClients(pkt *netutil.Packet) {
	for _, dcp := range service.clients {
		if dcp != nil {
			dcp.SendPacket(pkt)
		}
	}
}
