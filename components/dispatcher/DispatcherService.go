package main

import (
	"fmt"

	"net"

	"sync"

	"time"

	"sync/atomic"

	"github.com/xiaonanln/goSyncQueue"
	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/netutil"
	"github.com/xiaonanln/goworld/proto"
)

type callQueueItem struct {
	packet *netutil.Packet
}

type EntityDispatchInfo struct {
	sync.RWMutex

	gameid             uint16
	loadTime           time.Time
	migrateTime        time.Time
	pendingPacketQueue sync_queue.SyncQueue
}

func newEntityDispatchInfo() *EntityDispatchInfo {
	return &EntityDispatchInfo{
		pendingPacketQueue: sync_queue.NewSyncQueue(),
	}
}

func (info *EntityDispatchInfo) startMigrate() {
	info.migrateTime = time.Now()
}

func (info *EntityDispatchInfo) startLoad() {
	info.loadTime = time.Now()
}

func (info *EntityDispatchInfo) isBlockingRPC() bool {
	if info.migrateTime.IsZero() && info.loadTime.IsZero() {
		// most common case
		return false
	}

	now := time.Now()
	return now.Before(info.migrateTime.Add(consts.DISPATCHER_MIGRATE_TIMEOUT)) || now.Before(info.loadTime.Add(consts.DISPATCHER_LOAD_TIMEOUT))
}

type DispatcherService struct {
	config            *config.DispatcherConfig
	gameClients       []*DispatcherClientProxy
	gateClients       []*DispatcherClientProxy
	chooseClientIndex int64

	entityDispatchInfosLock sync.RWMutex
	entityDispatchInfos     map[common.EntityID]*EntityDispatchInfo

	servicesLock       sync.Mutex
	registeredServices map[string]entity.EntityIDSet

	clientsLock          sync.RWMutex
	targetServerOfClient map[common.ClientID]uint16
}

func newDispatcherService() *DispatcherService {
	cfg := config.Get()
	gameCount := len(cfg.Games)
	gateCount := len(cfg.Gates)
	return &DispatcherService{
		config:            &cfg.Dispatcher,
		gameClients:       make([]*DispatcherClientProxy, gameCount),
		gateClients:       make([]*DispatcherClientProxy, gateCount),
		chooseClientIndex: 0,

		entityDispatchInfos:  map[common.EntityID]*EntityDispatchInfo{},
		registeredServices:   map[string]entity.EntityIDSet{},
		targetServerOfClient: map[common.ClientID]uint16{},
	}
}

func (service *DispatcherService) getEntityDispatcherInfoForRead(entityID common.EntityID) (info *EntityDispatchInfo) {
	service.entityDispatchInfosLock.RLock()
	info = service.entityDispatchInfos[entityID] // can be nil
	if info != nil {
		info.RLock()
	}
	service.entityDispatchInfosLock.RUnlock()
	return
}

func (service *DispatcherService) getEntityDispatcherInfoForWrite(entityID common.EntityID) (info *EntityDispatchInfo) {
	service.entityDispatchInfosLock.RLock()
	info = service.entityDispatchInfos[entityID] // can be nil
	if info != nil {
		info.Lock()
	}
	service.entityDispatchInfosLock.RUnlock()
	return
}

func (service *DispatcherService) newEntityDispatcherInfo(entityID common.EntityID) (info *EntityDispatchInfo) {
	info = newEntityDispatchInfo()
	service.entityDispatchInfosLock.Lock()
	service.entityDispatchInfos[entityID] = info
	service.entityDispatchInfosLock.Unlock()
	return
}

func (service *DispatcherService) delEntityDispatchInfo(entityID common.EntityID) {
	service.entityDispatchInfosLock.Lock()
	delete(service.entityDispatchInfos, entityID)
	service.entityDispatchInfosLock.Unlock()
}

//func (service *DispatcherService) setEntityDispatcherInfo(entityID common.EntityID) (info *EntityDispatchInfo) {
//	service.entityDispatchInfosLock.RUnlock()
//	info = service.entityDispatchInfos[entityID]
//	service.entityDispatchInfosLock.RUnlock()
//
//	if info == nil {
//		service.entityDispatchInfosLock.Lock()
//		info = service.entityDispatchInfos[entityID] // need to re-retrive info after write-lock
//		if info == nil {
//			info = &EntityDispatchInfo{
//				pendingPacketQueue: sync_queue.NewSyncQueue(),
//			}
//			service.entityDispatchInfos[entityID] = info
//		}
//		service.entityDispatchInfosLock.Unlock()
//	}
//	return
//}

func (service *DispatcherService) setEntityDispatcherInfoForWrite(entityID common.EntityID) (info *EntityDispatchInfo) {
	service.entityDispatchInfosLock.RLock()
	info = service.entityDispatchInfos[entityID]

	if info != nil {
		info.Lock()
	}

	service.entityDispatchInfosLock.RUnlock()

	if info == nil {
		service.entityDispatchInfosLock.Lock()
		info = service.entityDispatchInfos[entityID] // need to re-retrive info after write-lock
		if info == nil {
			info = &EntityDispatchInfo{
				pendingPacketQueue: sync_queue.NewSyncQueue(),
			}
			service.entityDispatchInfos[entityID] = info
		}

		info.Lock()

		service.entityDispatchInfosLock.Unlock()
	}

	return
}

func (service *DispatcherService) String() string {
	return "DispatcherService"
}

func (service *DispatcherService) run() {
	host := fmt.Sprintf("%s:%d", service.config.Ip, service.config.Port)
	netutil.ServeTCPForever(host, service)
}

func (service *DispatcherService) ServeTCPConnection(conn net.Conn) {
	tcpConn := conn.(*net.TCPConn)
	tcpConn.SetReadBuffer(consts.DISPATCHER_CLIENT_PROXY_READ_BUFFER_SIZE)
	tcpConn.SetWriteBuffer(consts.DISPATCHER_CLIENT_PROXY_WRITE_BUFFER_SIZE)

	client := newDispatcherClientProxy(service, conn)
	client.serve()
}

func (service *DispatcherService) HandleSetServerID(dcp *DispatcherClientProxy, pkt *netutil.Packet, gameid uint16, isReconnect bool) {
	if consts.DEBUG_PACKETS {
		gwlog.Debug("%s.HandleSetServerID: dcp=%s, gameid=%d, isReconnect=%v", service, dcp, gameid, isReconnect)
	}
	if gameid <= 0 {
		gwlog.Panicf("invalid gameid: %d", gameid)
	}

	olddcp := service.gameClients[gameid-1] // should be nil, unless reconnect
	service.gameClients[gameid-1] = dcp
	// notify all games that all games connected to dispatcher now!
	if service.isAllGameClientsConnected() {
		pkt.ClearPayload() // reuse this packet
		pkt.AppendUint16(proto.MT_NOTIFY_ALL_SERVERS_CONNECTED)
		if olddcp == nil {
			// for the first time that all games connected to dispatcher, notify all games
			gwlog.Info("All games(%d) are connected", len(service.gameClients))
			service.broadcastToGameClients(pkt)
		} else { // dispatcher reconnected, only notify this game
			dcp.SendPacket(pkt)
		}
	}

	if olddcp != nil && !isReconnect {
		// game was connected, but a new instance is replaced, so we need to wipe the entities on that game
		service.cleanupEntitiesOfServer(gameid)
	}

	return
}

func (service *DispatcherService) HandleSetGateID(dcp *DispatcherClientProxy, pkt *netutil.Packet, gateid uint16) {
	service.gateClients[gateid-1] = dcp
}

func (service *DispatcherService) isAllGameClientsConnected() bool {
	for _, client := range service.gameClients {
		if client == nil {
			return false
		}
	}
	return true
}

func (service *DispatcherService) dispatcherClientOfGame(gameid uint16) *DispatcherClientProxy {
	return service.gameClients[gameid-1]
}

func (service *DispatcherService) dispatcherClientOfGate(gateid uint16) *DispatcherClientProxy {
	return service.gateClients[gateid-1]
}

// Choose a dispatcher client for sending Anywhere packets
func (service *DispatcherService) chooseGameDispatcherClient() *DispatcherClientProxy {
	index := atomic.LoadInt64(&service.chooseClientIndex)
	client := service.gameClients[index]
	atomic.StoreInt64(&service.chooseClientIndex, int64((int(index)+1)%len(service.gameClients)))
	return client
}

func (service *DispatcherService) HandleDispatcherClientDisconnect(dcp *DispatcherClientProxy) {
	// nothing to do when client disconnected
}

// Entity is create on the target game
func (service *DispatcherService) HandleNotifyCreateEntity(dcp *DispatcherClientProxy, pkt *netutil.Packet, entityID common.EntityID) {
	if consts.DEBUG_PACKETS {
		gwlog.Debug("%s.HandleNotifyCreateEntity: dcp=%s, entityID=%s", service, dcp, entityID)
	}
	entityDispatchInfo := service.setEntityDispatcherInfoForWrite(entityID)
	defer entityDispatchInfo.Unlock()

	entityDispatchInfo.gameid = dcp.gameid
	if !entityDispatchInfo.loadTime.IsZero() { // entity is loading, it's done now
		//gwlog.Info("entity is loaded now, clear loadTime")
		entityDispatchInfo.loadTime = time.Time{}
		service.sendPendingPackets(entityDispatchInfo)
	}
}

func (service *DispatcherService) HandleNotifyDestroyEntity(dcp *DispatcherClientProxy, pkt *netutil.Packet, entityID common.EntityID) {
	if consts.DEBUG_PACKETS {
		gwlog.Debug("%s.HandleNotifyDestroyEntity: dcp=%s, entityID=%s", service, dcp, entityID)
	}
	service.delEntityDispatchInfo(entityID)
}

func (service *DispatcherService) HandleNotifyClientConnected(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	clientid := pkt.ReadClientID()
	targetServer := service.chooseGameDispatcherClient()

	service.clientsLock.Lock()
	service.targetServerOfClient[clientid] = targetServer.gameid // owner is not determined yet, set to "" as placeholder
	service.clientsLock.Unlock()

	if consts.DEBUG_CLIENTS {
		gwlog.Debug("Target game of client %s is SET to %v on connected", clientid, targetServer.gameid)
	}

	pkt.AppendUint16(dcp.gateid)
	targetServer.SendPacket(pkt)
}

func (service *DispatcherService) HandleNotifyClientDisconnected(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	clientid := pkt.ReadClientID() // client disconnected

	service.clientsLock.Lock()
	targetSid := service.targetServerOfClient[clientid]
	delete(service.targetServerOfClient, clientid)
	service.clientsLock.Unlock()

	if consts.DEBUG_CLIENTS {
		gwlog.Debug("Target game of client %s is %v, disconnecting ...", clientid, targetSid)
	}

	if targetSid != 0 { // if found the owner, tell it
		service.dispatcherClientOfGame(targetSid).SendPacket(pkt) // tell the game that the client is down
	}
}

func (service *DispatcherService) HandleLoadEntityAnywhere(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	//typeName := pkt.ReadVarStr()
	//eid := pkt.ReadEntityID()
	if consts.DEBUG_PACKETS {
		gwlog.Debug("%s.HandleLoadEntityAnywhere: dcp=%s, pkt=%v", service, dcp, pkt.Payload())
	}
	eid := pkt.ReadEntityID() // field 1

	entityDispatchInfo := service.setEntityDispatcherInfoForWrite(eid)
	defer entityDispatchInfo.Unlock()

	if entityDispatchInfo.gameid == 0 { // entity not loaded, try load now
		dcp := service.chooseGameDispatcherClient()
		entityDispatchInfo.gameid = dcp.gameid
		entityDispatchInfo.startLoad()
		dcp.SendPacket(pkt)
	} else {
		// entity already loaded
	}
}

func (service *DispatcherService) HandleCreateEntityAnywhere(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	if consts.DEBUG_PACKETS {
		gwlog.Debug("%s.HandleCreateEntityAnywhere: dcp=%s, pkt=%s", service, dcp, pkt.Payload())
	}
	service.chooseGameDispatcherClient().SendPacket(pkt)
}

func (service *DispatcherService) HandleDeclareService(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	entityID := pkt.ReadEntityID()
	serviceName := pkt.ReadVarStr()
	if consts.DEBUG_PACKETS {
		gwlog.Debug("%s.HandleDeclareService: dcp=%s, entityID=%s, serviceName=%s", service, dcp, entityID, serviceName)
	}

	entityDispatchInfo := service.setEntityDispatcherInfoForWrite(entityID)
	entityDispatchInfo.gameid = dcp.gameid
	entityDispatchInfo.Unlock()

	service.servicesLock.Lock()
	if _, ok := service.registeredServices[serviceName]; !ok {
		service.registeredServices[serviceName] = entity.EntityIDSet{}
	}

	service.registeredServices[serviceName].Add(entityID)
	service.broadcastToGameClients(pkt)
	service.servicesLock.Unlock()
}

func (service *DispatcherService) handleServiceDown(serviceName string, eid common.EntityID) {
	pkt := netutil.NewPacket()
	pkt.AppendUint16(proto.MT_UNDECLARE_SERVICE)
	pkt.AppendEntityID(eid)
	pkt.AppendVarStr(serviceName)

	service.broadcastToGameClients(pkt)
	pkt.Release()
}

func (service *DispatcherService) HandleCallEntityMethod(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	entityID := pkt.ReadEntityID()

	if consts.DEBUG_PACKETS {
		gwlog.Debug("%s.HandleCallEntityMethod: dcp=%s, entityID=%s", service, dcp, entityID)
	}

	entityDispatchInfo := service.getEntityDispatcherInfoForRead(entityID)
	if entityDispatchInfo == nil {
		// entity not exists ?
		gwlog.Error("%s.HandleCallEntityMethod: entity %s not found", service, entityID)
		return
	}

	defer entityDispatchInfo.RUnlock()

	if !entityDispatchInfo.isBlockingRPC() {
		service.dispatcherClientOfGame(entityDispatchInfo.gameid).SendPacket(pkt)
	} else {
		// if migrating, just put the call to wait
		if entityDispatchInfo.pendingPacketQueue.Len() < consts.ENTITY_PENDING_PACKET_QUEUE_MAX_LEN {
			pkt.AddRefCount(1)
			entityDispatchInfo.pendingPacketQueue.Push(callQueueItem{
				packet: pkt,
			})
		} else {
			gwlog.Error("%s.HandleCallEntityMethod %s: packet queue too long, packet dropped", service, entityID)
		}
	}
}

func (service *DispatcherService) HandleCallEntityMethodFromClient(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	entityID := pkt.ReadEntityID()

	if consts.DEBUG_PACKETS {
		gwlog.Debug("%s.HandleCallEntityMethodFromClient: entityID=%s, payload=%v", service, entityID, pkt.Payload())
	}

	entityDispatchInfo := service.getEntityDispatcherInfoForRead(entityID)
	if entityDispatchInfo == nil {
		gwlog.Error("%s.HandleCallEntityMethodFromClient: entity %s is not found: %v", service, entityID, service.entityDispatchInfos)
		return
	}

	defer entityDispatchInfo.RUnlock()

	if !entityDispatchInfo.isBlockingRPC() {
		service.dispatcherClientOfGame(entityDispatchInfo.gameid).SendPacket(pkt)
	} else {
		// if migrating, just put the call to wait
		if entityDispatchInfo.pendingPacketQueue.Len() < consts.ENTITY_PENDING_PACKET_QUEUE_MAX_LEN {
			pkt.AddRefCount(1)
			entityDispatchInfo.pendingPacketQueue.Push(callQueueItem{
				packet: pkt,
			})
		} else {
			gwlog.Error("%s.HandleCallEntityMethodFromClient %s: packet queue too long, packet dropped", service, entityID)
		}
	}

}

func (service DispatcherService) HandleDoSomethingOnSpecifiedClient(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	gid := pkt.ReadUint16()
	service.dispatcherClientOfGate(gid).SendPacket(pkt)
}

func (service *DispatcherService) HandleCallFilteredClientProxies(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	service.broadcastToGateClients(pkt)
}

func (service *DispatcherService) HandleMigrateRequest(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	entityID := pkt.ReadEntityID()
	spaceID := pkt.ReadEntityID()
	if consts.DEBUG_PACKETS {
		gwlog.Debug("Entity %s is migrating to space %s", entityID, spaceID)
	}

	// mark the entity as migrating

	spaceDispatchInfo := service.getEntityDispatcherInfoForRead(spaceID)
	var spaceLoc uint16
	if spaceDispatchInfo != nil {
		spaceLoc = spaceDispatchInfo.gameid
	}
	spaceDispatchInfo.RUnlock()

	pkt.AppendUint16(spaceLoc) // append the space game location to the packet

	if spaceLoc > 0 { // almost true
		entityDispatchInfo := service.setEntityDispatcherInfoForWrite(entityID)
		defer entityDispatchInfo.Unlock()

		entityDispatchInfo.startMigrate()
	}

	dcp.SendPacket(pkt)
}

func (service *DispatcherService) HandleRealMigrate(dcp *DispatcherClientProxy, pkt *netutil.Packet) {
	// get spaceID and make sure it exists
	eid := pkt.ReadEntityID()
	targetServer := pkt.ReadUint16() // target game of migration
	// target space is not checked for existence, because we relay the packet anyway

	hasClient := pkt.ReadBool()
	var clientid common.ClientID
	if hasClient {
		clientid = pkt.ReadClientID()
	}

	// mark the eid as migrating done
	entityDispatchInfo := service.setEntityDispatcherInfoForWrite(eid)
	defer entityDispatchInfo.Unlock()

	entityDispatchInfo.migrateTime = time.Time{} // mark the entity as NOT migrating
	entityDispatchInfo.gameid = targetServer
	service.clientsLock.Lock()
	service.targetServerOfClient[clientid] = targetServer // migrating also change target game of client
	service.clientsLock.Unlock()

	if consts.DEBUG_CLIENTS {
		gwlog.Debug("Target game of client %s is migrated to %v along with owner %s", clientid, targetServer, eid)
	}

	service.dispatcherClientOfGame(targetServer).SendPacket(pkt)
	// send the cached calls to target game
	service.sendPendingPackets(entityDispatchInfo)
}

func (service *DispatcherService) sendPendingPackets(entityDispatchInfo *EntityDispatchInfo) {
	targetServer := entityDispatchInfo.gameid
	// send the cached calls to target game
	item, ok := entityDispatchInfo.pendingPacketQueue.TryPop()
	for ok {
		cachedPkt := item.(callQueueItem).packet
		service.dispatcherClientOfGame(targetServer).SendPacket(cachedPkt)
		cachedPkt.Release()

		item, ok = entityDispatchInfo.pendingPacketQueue.TryPop()
	}
}

func (service *DispatcherService) broadcastToGameClients(pkt *netutil.Packet) {
	for _, dcp := range service.gameClients {
		dcp.SendPacket(pkt)
	}
}

func (service *DispatcherService) broadcastToGateClients(pkt *netutil.Packet) {
	for _, dcp := range service.gateClients {
		dcp.SendPacket(pkt)
	}
}

func (service *DispatcherService) cleanupEntitiesOfServer(targetServer uint16) {
	service.entityDispatchInfosLock.Lock()
	defer service.entityDispatchInfosLock.Unlock()

	service.servicesLock.Lock()
	defer service.servicesLock.Unlock()

	cleanEids := entity.EntityIDSet{} // get all clean eids
	for eid, dispatchInfo := range service.entityDispatchInfos {
		if dispatchInfo.gameid == targetServer {
			cleanEids.Add(eid)
		}
	}

	// for all services whose entity is cleaned, notify all games that the service is down
	undeclaredServices := common.StringSet{}
	for serviceName, serviceEids := range service.registeredServices {
		var cleanEidsOfServer []common.EntityID
		for serviceEid := range serviceEids {
			if cleanEids.Contains(serviceEid) { // this service entity is down, tell other games
				undeclaredServices.Add(serviceName)
				cleanEidsOfServer = append(cleanEidsOfServer, serviceEid)
				service.handleServiceDown(serviceName, serviceEid)
			}
		}

		for _, eid := range cleanEidsOfServer {
			serviceEids.Del(eid)
		}
	}

	for eid := range cleanEids {
		delete(service.entityDispatchInfos, eid)
	}

	gwlog.Info("Server %d is rebooted, %d entities cleaned, undeclare services: %s", targetServer, len(cleanEids), undeclaredServices)
}
