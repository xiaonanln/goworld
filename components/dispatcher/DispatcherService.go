package main

import (
	"fmt"

	"net"

	"sync"

	"time"

	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/netutil"
	"github.com/xiaonanln/goworld/proto"
)

type EntityDispatchInfo struct {
	serverid    uint16
	migrateTime int64
}

type DispatcherService struct {
	sync.RWMutex

	config            *config.DispatcherConfig
	clients           []*DispatcherClientProxy
	chooseClientIndex int

	entityDispatchInfo   map[common.EntityID]*EntityDispatchInfo
	registeredServices   map[string]entity.EntityIDSet
	targetServerOfClient map[common.ClientID]uint16
}

func newDispatcherService() *DispatcherService {
	cfg := config.Get()
	serverCount := len(cfg.Servers)
	return &DispatcherService{
		config:            &cfg.Dispatcher,
		clients:           make([]*DispatcherClientProxy, serverCount),
		chooseClientIndex: 0,

		entityDispatchInfo:   map[common.EntityID]*EntityDispatchInfo{},
		registeredServices:   map[string]entity.EntityIDSet{},
		targetServerOfClient: map[common.ClientID]uint16{},
	}
}

func (service *DispatcherService) String() string {
	return fmt.Sprintf("DispatcherService<C%d|E%d>", len(service.clients), len(service.entityDispatchInfo))
}

func (service *DispatcherService) run() {
	host := fmt.Sprintf("%s:%d", service.config.Ip, service.config.Port)
	netutil.ServeTCPForever(host, service)
}

func (service *DispatcherService) ServeTCPConnection(conn net.Conn) {
	client := newDispatcherClientProxy(service, conn)
	client.serve()
}

func (service *DispatcherService) HandleSetServerID(dcp *DispatcherClientProxy, pkt *netutil.Packet, serverid uint16, isReconnect bool) {
	if consts.DEBUG_PACKETS {
		gwlog.Debug("%s.HandleSetServerID: dcp=%s, serverid=%d, isReconnect=%v", service, dcp, serverid, isReconnect)
	}
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
		} else { // dispatcher reconnected, only notify this server
			dcp.SendPacket(pkt)
		}
	}
	pkt.Release()

	if olddcp != nil && !isReconnect {
		// server was connected, but a new instance is replaced, so we need to wipe the entities on that server
		service.cleanupEntitiesOfServer(serverid)
	}

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

func (service *DispatcherService) HandleDispatcherClientDisconnect(dcp *DispatcherClientProxy) {
	// nothing to do when client disconnected
}

// Entity is create on the target server
func (service *DispatcherService) HandleNotifyCreateEntity(dcp *DispatcherClientProxy, pkt *netutil.Packet, entityID common.EntityID) {
	if consts.DEBUG_PACKETS {
		gwlog.Debug("%s.HandleNotifyCreateEntity: dcp=%s, entityID=%s", service, dcp, entityID)
	}
	service.Lock()
	service.entityLocs[entityID] = dcp.serverid
	service.Unlock()
	pkt.Release()
}

func (service *DispatcherService) HandleNotifyDestroyEntity(dcp *DispatcherClientProxy, pkt *netutil.Packet, entityID common.EntityID) {
	if consts.DEBUG_PACKETS {
		gwlog.Debug("%s.HandleNotifyDestroyEntity: dcp=%s, entityID=%s", service, dcp, entityID)
	}
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

func (service *DispatcherService) HandleNotifyClientDisconnected(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	clientid := pkt.ReadClientID() // client disconnected
	service.RLock()
	sid := service.targetServerOfClient[clientid] // target server of client
	service.RUnlock()

	service.dispatcherClientOfServer(sid).SendPacketRelease(pkt) // tell the server that the client is down
}

func (service *DispatcherService) HandleLoadEntityAnywhere(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	//typeName := pkt.ReadVarStr()
	//eid := pkt.ReadEntityID()
	if consts.DEBUG_PACKETS {
		gwlog.Debug("%s.HandleLoadEntityAnywhere: dcp=%s, pkt=%v", service, dcp, pkt.Payload())
	}
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
	if consts.DEBUG_PACKETS {
		gwlog.Debug("%s.HandleCreateEntityAnywhere: dcp=%s, pkt=%s", service, dcp, pkt.Payload())
	}
	service.chooseDispatcherClient().SendPacketRelease(pkt)
}

func (service *DispatcherService) HandleDeclareService(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	entityID := pkt.ReadEntityID()
	serviceName := pkt.ReadVarStr()
	if consts.DEBUG_PACKETS {
		gwlog.Debug("%s.HandleDeclareService: dcp=%s, entityID=%s, serviceName=%s", service, dcp, entityID, serviceName)
	}
	service.Lock()
	service.entityLocs[entityID] = dcp.serverid
	if _, ok := service.registeredServices[serviceName]; !ok {
		service.registeredServices[serviceName] = entity.EntityIDSet{}
	}
	service.registeredServices[serviceName].Add(entityID)
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

func (service *DispatcherService) handleServiceDown(serviceName string, eid common.EntityID) {
	pkt := netutil.NewPacket()
	pkt.AppendUint16(proto.MT_UNDECLARE_SERVICE)
	pkt.AppendEntityID(eid)
	pkt.AppendVarStr(serviceName)

	service.broadcastToDispatcherClients(pkt)
}

func (service *DispatcherService) HandleCallEntityMethod(dcp *DispatcherClientProxy, pkt *netutil.Packet, entityID common.EntityID, method string) {
	if consts.DEBUG_PACKETS {
		gwlog.Debug("%s.HandleCallEntityMethod: dcp=%s, entityID=%s, method=%s", service, dcp, entityID, method)
	}

	service.RLock()
	serverid := service.entityLocs[entityID] // TODO CACHE WHILE MIGRATING
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
	service.RLock()
	sid := service.targetServerOfClient[clientid]
	service.RUnlock()
	if consts.DEBUG_PACKETS {
		gwlog.Debug("%s.HandleCallEntityMethodFromClient: %s.%s %v, clientid=%s, sid=%d", service, entityid, method, args, clientid, sid)
	}
	service.dispatcherClientOfServer(sid).SendPacketRelease(pkt)
}

func (service *DispatcherService) HandleCreateEntityOnClient(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	sid := pkt.ReadUint16()
	clientid := pkt.ReadClientID()
	service.Lock()
	service.targetServerOfClient[clientid] = sid
	service.Unlock()
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

func (service *DispatcherService) HandleNotifyAttrDelOnClient(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	sid := pkt.ReadUint16()
	service.dispatcherClientOfServer(sid).SendPacketRelease(pkt)
}

func (service *DispatcherService) HandleMigrateRequest(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	entityID := pkt.ReadEntityID()
	spaceID := pkt.ReadEntityID() // TODO: no need spaceID?
	if consts.DEBUG_PACKETS {
		gwlog.Debug("Entity %s is migrating to space %s", entityID, spaceID)
	}
	// mark the entity as migrating
	service.Lock()
	spaceLoc := service.entityLocs[spaceID]
	if spaceLoc > 0 {
		service.entityMigrateTime[entityID] = time.Now().UnixNano() // TODO: handle multiple duplicate request?
	}
	service.Unlock()

	pkt.AppendUint16(spaceLoc) // append the space server location to the packet
	dcp.SendPacketRelease(pkt)
}

func (service *DispatcherService) HandleRealMigrate(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	// get spaceID and make sure it exists
	eid := pkt.ReadEntityID()
	targetServer := pkt.ReadUint16() // target server of migration
	// target space is not checked for existence, because we relay the packet anyway
	// mark the eid as migrating done
	service.Lock()
	service.entityMigrateTime[eid] = 0 // mark the entity as NOT migrating
	service.entityLocs[eid] = targetServer
	service.Unlock()

	service.dispatcherClientOfServer(targetServer).SendPacketRelease(pkt)
}

func (service *DispatcherService) broadcastToDispatcherClients(pkt *netutil.Packet) {
	for _, dcp := range service.clients {
		dcp.SendPacket(pkt)
	}
}

func (service *DispatcherService) cleanupEntitiesOfServer(targetServer uint16) {
	cleanEids := entity.EntityIDSet{} // get all clean eids
	for eid, sid := range service.entityLocs {
		if sid == targetServer {
			cleanEids.Add(eid)
		}
	}

	// for all services whose entity is cleaned, notify all servers that the service is down
	undeclaredServices := common.StringSet{}
	for serviceName, serviceEids := range service.registeredServices {
		for serviceEid := range serviceEids {
			if cleanEids.Contains(serviceEid) { // this service entity is down, tell other servers
				undeclaredServices.Add(serviceName)
				service.handleServiceDown(serviceName, serviceEid)
			}
		}
	}

	for eid := range cleanEids {
		delete(service.entityLocs, eid)
	}

	gwlog.Info("Server %d is rebooted, %d entities cleaned, undeclare services: %s", targetServer, len(cleanEids), undeclaredServices)
}
