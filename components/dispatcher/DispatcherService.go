package main

import (
	"fmt"

	"net"

	"time"

	"os"

	"github.com/xiaonanln/go-xnsyncutil/xnsyncutil"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/gwutils"
	"github.com/xiaonanln/goworld/engine/netutil"
	"github.com/xiaonanln/goworld/engine/proto"
)

type callQueueItem struct {
	packet *netutil.Packet
}

type entityDispatchInfo struct {
	gameid             uint16
	blockUntilTime     time.Time
	pendingPacketQueue []*netutil.Packet
}

func newEntityDispatchInfo() *entityDispatchInfo {
	return &entityDispatchInfo{
		pendingPacketQueue: nil,
	}
}

func (info *entityDispatchInfo) blockRPC(d time.Duration) {
	t := time.Now().Add(d)
	if info.blockUntilTime.Before(t) {
		info.blockUntilTime = t
	}
}

func (info *entityDispatchInfo) isBlockingRPC() bool {
	if info.blockUntilTime.IsZero() {
		// most common case
		return false
	}

	now := time.Now()
	return now.Before(info.blockUntilTime)
}

type dispatcherMessage struct {
	dcp *dispatcherClientProxy
	proto.Message
}

type dispatcherCommandType int

const (
	dispatcherCommandExit dispatcherCommandType = iota
)

// DispatcherService implements the dispatcher service
type DispatcherService struct {
	dispid            uint16
	config            *config.DispatcherConfig
	gameClients       []*dispatcherClientProxy
	gateClients       []*dispatcherClientProxy
	messageQueue      chan dispatcherMessage
	commandQueue      chan dispatcherCommandType
	chooseClientIndex xnsyncutil.AtomicInt

	entityDispatchInfos map[common.EntityID]*entityDispatchInfo

	registeredServices map[string]entity.EntityIDSet

	targetGameOfClient map[common.ClientID]uint16

	entitySyncInfosToGame [][]byte // cache entity sync infos to gates
}

func newDispatcherService(dispid uint16) *DispatcherService {
	cfg := config.GetDispatcher(dispid)
	gameCount := len(config.GetGameIDs())
	gateCount := len(config.GetGateIDs())
	ds := &DispatcherService{
		dispid:            dispid,
		config:            cfg,
		messageQueue:      make(chan dispatcherMessage, consts.DISPATCHER_SERVICE_PACKET_QUEUE_SIZE),
		commandQueue:      make(chan dispatcherCommandType),
		gameClients:       make([]*dispatcherClientProxy, gameCount),
		gateClients:       make([]*dispatcherClientProxy, gateCount),
		chooseClientIndex: 0,

		entityDispatchInfos: map[common.EntityID]*entityDispatchInfo{},
		registeredServices:  map[string]entity.EntityIDSet{},
		targetGameOfClient:  map[common.ClientID]uint16{},

		entitySyncInfosToGame: make([][]byte, gameCount),
	}

	return ds
}

func (service *DispatcherService) messageLoop() {
	for {
		select {
		case msg := <-service.messageQueue:
			dcp := msg.dcp
			msgtype := msg.MsgType
			pkt := msg.Packet

			if msgtype == proto.MT_SYNC_POSITION_YAW_FROM_CLIENT {
				service.handleSyncPositionYawFromClient(dcp, pkt)
			} else if msgtype == proto.MT_SYNC_POSITION_YAW_ON_CLIENTS {
				service.handleSyncPositionYawOnClients(dcp, pkt)
			} else if msgtype == proto.MT_CALL_ENTITY_METHOD {
				service.handleCallEntityMethod(dcp, pkt)
			} else if msgtype >= proto.MT_REDIRECT_TO_GATEPROXY_MSG_TYPE_START && msgtype <= proto.MT_REDIRECT_TO_GATEPROXY_MSG_TYPE_STOP {
				service.handleDoSomethingOnSpecifiedClient(dcp, pkt)
			} else if msgtype == proto.MT_CALL_ENTITY_METHOD_FROM_CLIENT {
				service.handleCallEntityMethodFromClient(dcp, pkt)
			} else if msgtype == proto.MT_MIGRATE_REQUEST {
				service.handleMigrateRequest(dcp, pkt)
			} else if msgtype == proto.MT_REAL_MIGRATE {
				service.handleRealMigrate(dcp, pkt)
			} else if msgtype == proto.MT_CALL_FILTERED_CLIENTS {
				service.handleCallFilteredClientProxies(dcp, pkt)
			} else if msgtype == proto.MT_NOTIFY_CLIENT_CONNECTED {
				service.handleNotifyClientConnected(dcp, pkt)
			} else if msgtype == proto.MT_NOTIFY_CLIENT_DISCONNECTED {
				service.handleNotifyClientDisconnected(dcp, pkt)
			} else if msgtype == proto.MT_LOAD_ENTITY_ANYWHERE {
				service.handleLoadEntityAnywhere(dcp, pkt)
			} else if msgtype == proto.MT_NOTIFY_CREATE_ENTITY {
				eid := pkt.ReadEntityID()
				service.handleNotifyCreateEntity(dcp, pkt, eid)
			} else if msgtype == proto.MT_NOTIFY_DESTROY_ENTITY {
				eid := pkt.ReadEntityID()
				service.handleNotifyDestroyEntity(dcp, pkt, eid)
			} else if msgtype == proto.MT_CREATE_ENTITY_ANYWHERE {
				service.handleCreateEntityAnywhere(dcp, pkt)
			} else if msgtype == proto.MT_DECLARE_SERVICE {
				service.handleDeclareService(dcp, pkt)
			} else if msgtype == proto.MT_SET_GAME_ID {
				// this is a game server
				service.handleSetGameID(dcp, pkt)
			} else if msgtype == proto.MT_SET_GATE_ID {
				// this is a gate
				service.handleSetGateID(dcp, pkt)
			} else if msgtype == proto.MT_START_FREEZE_GAME {
				// freeze the game
				service.handleStartFreezeGame(dcp, pkt)
			} else {
				gwlog.TraceError("unknown msgtype %d from %s", msgtype, dcp)
				if consts.DEBUG_MODE {
					os.Exit(2)
				}
			}

			pkt.Release()
			break
		case cmd := <-service.commandQueue:
			service.handleCommand(cmd)
			break
		}
	}
}

func (service *DispatcherService) handleCommand(cmd dispatcherCommandType) {
	if cmd == dispatcherCommandExit {
		//gamesNum := dispatcherService.connectedGameClientsNum()
		//if gamesNum > 0 {
		//	gwlog.Warnf("%d games is connected to dispatcher, can not quit!", gamesNum)
		//	return
		//}

		gwlog.Infof("Dispatcher terminated gracefully.")
		os.Exit(0)
	}
}

func (service *DispatcherService) getEntityDispatcherInfoForRead(entityID common.EntityID) (info *entityDispatchInfo) {
	info = service.entityDispatchInfos[entityID] // can be nil
	return
}

func (service *DispatcherService) getEntityDispatcherInfoForWrite(entityID common.EntityID) (info *entityDispatchInfo) {
	info = service.entityDispatchInfos[entityID] // can be nil
	return
}

func (service *DispatcherService) newEntityDispatcherInfo(entityID common.EntityID) (info *entityDispatchInfo) {
	info = newEntityDispatchInfo()
	service.entityDispatchInfos[entityID] = info
	return
}

func (service *DispatcherService) delEntityDispatchInfo(entityID common.EntityID) {
	delete(service.entityDispatchInfos, entityID)
}

func (service *DispatcherService) setEntityDispatcherInfoForWrite(entityID common.EntityID) (info *entityDispatchInfo) {
	info = service.entityDispatchInfos[entityID]

	if info == nil {
		info = &entityDispatchInfo{}
		service.entityDispatchInfos[entityID] = info
	}

	return
}

func (service *DispatcherService) String() string {
	return "DispatcherService"
}

func (service *DispatcherService) run() {
	host := fmt.Sprintf("%s:%d", service.config.BindIp, service.config.BindPort)
	fmt.Fprintf(gwlog.GetOutput(), "%s\n", consts.DISPATCHER_STARTED_TAG)
	go gwutils.RepeatUntilPanicless(service.messageLoop)
	netutil.ServeTCPForever(host, service)
}

// ServeTCPConnection handles dispatcher client connections to dispatcher
func (service *DispatcherService) ServeTCPConnection(conn net.Conn) {
	tcpConn := conn.(*net.TCPConn)
	tcpConn.SetReadBuffer(consts.DISPATCHER_CLIENT_PROXY_READ_BUFFER_SIZE)
	tcpConn.SetWriteBuffer(consts.DISPATCHER_CLIENT_PROXY_WRITE_BUFFER_SIZE)

	client := newDispatcherClientProxy(service, conn)
	client.serve()
}

func (service *DispatcherService) handleSetGameID(dcp *dispatcherClientProxy, pkt *netutil.Packet) {

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
	dcp.startAutoFlush() // TODO: why start autoflush after gameid is set ?

	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleSetGameID: dcp=%s, gameid=%d, isReconnect=%v", service, dcp, gameid, isReconnect)
	}

	olddcp := service.gameClients[gameid-1] // should be nil, unless reconnect
	service.gameClients[gameid-1] = dcp

	if !isRestore {
		// notify all games that all games connected to dispatcher now!
		if service.isAllGameClientsConnected() {
			pkt.ClearPayload() // reuse this packet
			pkt.AppendUint16(proto.MT_NOTIFY_ALL_GAMES_CONNECTED)
			if olddcp == nil {
				// for the first time that all games connected to dispatcher, notify all games
				gwlog.Infof("All games(%d) are connected", len(service.gameClients))
				service.broadcastToGameClients(pkt)
			} else { // dispatcher reconnected, only notify this game
				dcp.SendPacket(pkt)
			}
		}
	}

	if olddcp != nil && !isReconnect && !isRestore {
		// game was connected, but a new instance is replaced, so we need to wipe the entities on that game
		service.cleanupEntitiesOfGame(gameid)
	}

	if isRestore {
		gwlog.Debugf("Game %d restored: %s", dcp.gameid, dcp)
	}

	return
}

func (service *DispatcherService) handleSetGateID(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	gateid := pkt.ReadUint16()
	if gateid <= 0 {
		gwlog.Panicf("invalid gateid: %d", gateid)
	}
	if dcp.gameid > 0 || dcp.gateid > 0 {
		gwlog.Panicf("already set gameid=%d, gateid=%d", dcp.gameid, dcp.gateid)
	}
	dcp.gateid = gateid
	dcp.startAutoFlush()

	service.gateClients[gateid-1] = dcp
}

func (service *DispatcherService) handleStartFreezeGame(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	// freeze the game, which block all entities of that game
	gwlog.Infof("Handling start freeze game ...")
	gameid := dcp.gameid

	for _, info := range service.entityDispatchInfos {
		if info.gameid == gameid {
			info.blockRPC(consts.DISPATCHER_FREEZE_GAME_TIMEOUT)
		}
	}

	// tell the game to start real freeze, using the packet
	pkt.ClearPayload()
	pkt.AppendUint16(proto.MT_START_FREEZE_GAME_ACK)
	pkt.AppendUint16(dispid)
	dcp.SendPacket(pkt)
}

func (service *DispatcherService) isAllGameClientsConnected() bool {
	for _, client := range service.gameClients {
		if client == nil {
			return false
		}
	}
	return true
}

func (service *DispatcherService) connectedGameClientsNum() int {
	num := 0
	for _, client := range service.gameClients {
		if client != nil {
			num += 1
		}
	}
	return num
}

func (service *DispatcherService) dispatcherClientOfGame(gameid uint16) *dispatcherClientProxy {
	return service.gameClients[gameid-1]
}

func (service *DispatcherService) dispatcherClientOfGate(gateid uint16) *dispatcherClientProxy {
	return service.gateClients[gateid-1]
}

// Choose a dispatcher client for sending Anywhere packets
func (service *DispatcherService) chooseGameDispatcherClient() *dispatcherClientProxy {
	index := service.chooseClientIndex.Load()
	client := service.gameClients[index]
	service.chooseClientIndex.Store((index + 1) % len(service.gameClients))
	return client
}

func (service *DispatcherService) handleDispatcherClientDisconnect(dcp *dispatcherClientProxy) {
	// nothing to do when client disconnected
	defer func() {
		if err := recover(); err != nil {
			gwlog.Infof("handleDispatcherClientDisconnect paniced: %v", err)
		}
	}()
	gwlog.Warnf("%s disconnected", dcp)
	if dcp.gateid > 0 {
		// gate disconnected, notify all clients disconnected
		service.handleGateDown(dcp.gateid)
	}
}

func (service *DispatcherService) handleGateDown(gateid uint16) {
	pkt := netutil.NewPacket()
	pkt.AppendUint16(proto.MT_NOTIFY_GATE_DISCONNECTED)
	pkt.AppendUint16(gateid)
	service.broadcastToGameClients(pkt)
	pkt.Release()
}

// Entity is create on the target game
func (service *DispatcherService) handleNotifyCreateEntity(dcp *dispatcherClientProxy, pkt *netutil.Packet, entityID common.EntityID) {
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleNotifyCreateEntity: dcp=%s, entityID=%s", service, dcp, entityID)
	}
	entityDispatchInfo := service.setEntityDispatcherInfoForWrite(entityID)

	entityDispatchInfo.gameid = dcp.gameid

	if !entityDispatchInfo.blockUntilTime.IsZero() { // entity is loading, it's done now
		//gwlog.Infof("entity is loaded now, clear loadTime")
		entityDispatchInfo.blockUntilTime = time.Time{}
		service.sendPendingPackets(entityDispatchInfo)
	}
}

func (service *DispatcherService) handleNotifyDestroyEntity(dcp *dispatcherClientProxy, pkt *netutil.Packet, entityID common.EntityID) {
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleNotifyDestroyEntity: dcp=%s, entityID=%s", service, dcp, entityID)
	}
	service.delEntityDispatchInfo(entityID)
}

func (service *DispatcherService) handleNotifyClientConnected(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	clientid := pkt.ReadClientID()
	targetGame := service.chooseGameDispatcherClient()

	service.targetGameOfClient[clientid] = targetGame.gameid // owner is not determined yet, set to "" as placeholder

	if consts.DEBUG_CLIENTS {
		gwlog.Debugf("Target game of client %s is SET to %v on connected", clientid, targetGame.gameid)
	}

	pkt.AppendUint16(dcp.gateid)
	targetGame.SendPacket(pkt)
}

func (service *DispatcherService) handleNotifyClientDisconnected(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	clientid := pkt.ReadClientID() // client disconnected

	targetSid := service.targetGameOfClient[clientid]
	delete(service.targetGameOfClient, clientid)

	if consts.DEBUG_CLIENTS {
		gwlog.Debugf("Target game of client %s is %v, disconnecting ...", clientid, targetSid)
	}

	if targetSid != 0 { // if found the owner, tell it
		service.dispatcherClientOfGame(targetSid).SendPacket(pkt) // tell the game that the client is down
	}
}

func (service *DispatcherService) handleLoadEntityAnywhere(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	//typeName := pkt.ReadVarStr()
	//eid := pkt.ReadEntityID()
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleLoadEntityAnywhere: dcp=%s, pkt=%v", service, dcp, pkt.Payload())
	}
	eid := pkt.ReadEntityID() // field 1

	entityDispatchInfo := service.setEntityDispatcherInfoForWrite(eid)

	if entityDispatchInfo.gameid == 0 { // entity not loaded, try load now
		dcp := service.chooseGameDispatcherClient()
		entityDispatchInfo.gameid = dcp.gameid
		entityDispatchInfo.blockRPC(consts.DISPATCHER_LOAD_TIMEOUT)
		dcp.SendPacket(pkt)
	}
}

func (service *DispatcherService) handleCreateEntityAnywhere(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleCreateEntityAnywhere: dcp=%s, pkt=%s", service, dcp, pkt.Payload())
	}
	// TODO: record entityid -> gameid ?
	service.chooseGameDispatcherClient().SendPacket(pkt)
}

func (service *DispatcherService) handleDeclareService(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	entityID := pkt.ReadEntityID()
	serviceName := pkt.ReadVarStr()
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleDeclareService: dcp=%s, entityID=%s, serviceName=%s", service, dcp, entityID, serviceName)
	}

	entityDispatchInfo := service.setEntityDispatcherInfoForWrite(entityID)
	entityDispatchInfo.gameid = dcp.gameid

	if _, ok := service.registeredServices[serviceName]; !ok {
		service.registeredServices[serviceName] = entity.EntityIDSet{}
	}

	service.registeredServices[serviceName].Add(entityID)
	service.broadcastToGameClients(pkt)
}

func (service *DispatcherService) handleServiceDown(serviceName string, eid common.EntityID) {
	pkt := netutil.NewPacket()
	pkt.AppendUint16(proto.MT_UNDECLARE_SERVICE)
	pkt.AppendEntityID(eid)
	pkt.AppendVarStr(serviceName)
	service.broadcastToGameClients(pkt)
	pkt.Release()
}

func (service *DispatcherService) handleCallEntityMethod(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	entityID := pkt.ReadEntityID()

	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleCallEntityMethod: dcp=%s, entityID=%s", service, dcp, entityID)
	}

	entityDispatchInfo := service.getEntityDispatcherInfoForRead(entityID)
	if entityDispatchInfo == nil {
		// entity not exists ?
		//gwlog.Errorf("%s.handleCallEntityMethod: entity %s not found", service, entityID)
		return
	}

	if !entityDispatchInfo.isBlockingRPC() {
		service.dispatcherClientOfGame(entityDispatchInfo.gameid).SendPacket(pkt)
	} else {
		// if migrating, just put the call to wait
		if len(entityDispatchInfo.pendingPacketQueue) < consts.ENTITY_PENDING_PACKET_QUEUE_MAX_LEN {
			pkt.AddRefCount(1)
			entityDispatchInfo.pendingPacketQueue = append(entityDispatchInfo.pendingPacketQueue, pkt)
		} else {
			gwlog.Errorf("%s.handleCallEntityMethod %s: packet queue too long, packet dropped", service, entityID)
		}
	}
}

func (service *DispatcherService) handleSyncPositionYawOnClients(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	gateid := pkt.ReadUint16()
	service.dispatcherClientOfGate(gateid).SendPacket(pkt)
}

func (service *DispatcherService) handleSyncPositionYawFromClient(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	// This sync packet contains position-yaw of multiple entities from a gate. Cache the packet to be send before flush?
	payload := pkt.UnreadPayload()

	for i := 0; i < len(payload); i += proto.SYNC_INFO_SIZE_PER_ENTITY + common.ENTITYID_LENGTH {
		eid := common.EntityID(payload[i : i+common.ENTITYID_LENGTH]) // the first bytes of each entry is the EntityID

		entityDispatchInfo := service.getEntityDispatcherInfoForRead(eid)
		gameid := entityDispatchInfo.gameid

		// put this sync info to the pending queue of target game
		// concat to the end of queue
		if len(service.entitySyncInfosToGame[gameid-1]) < consts.MAX_ENTITY_SYNC_INFOS_CACHE_SIZE_PER_GAME { // when game is freezed, prohibit caching too much data per game
			service.entitySyncInfosToGame[gameid-1] = append(service.entitySyncInfosToGame[gameid-1], payload[i:i+proto.SYNC_INFO_SIZE_PER_ENTITY+common.ENTITYID_LENGTH]...)
		}
	}
}

func (service *DispatcherService) popEntitySyncInfosToGame(gameid uint16) []byte {
	entitySyncInfos := service.entitySyncInfosToGame[gameid-1]
	service.entitySyncInfosToGame[gameid-1] = make([]byte, 0, len(entitySyncInfos))
	return entitySyncInfos
}

func (service *DispatcherService) handleCallEntityMethodFromClient(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	entityID := pkt.ReadEntityID()

	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleCallEntityMethodFromClient: entityID=%s, payload=%v", service, entityID, pkt.Payload())
	}

	entityDispatchInfo := service.getEntityDispatcherInfoForRead(entityID)
	if entityDispatchInfo == nil {
		gwlog.Errorf("%s.handleCallEntityMethodFromClient: entity %s is not found: %v", service, entityID, service.entityDispatchInfos)
		return
	}

	if !entityDispatchInfo.isBlockingRPC() {
		service.dispatcherClientOfGame(entityDispatchInfo.gameid).SendPacket(pkt)
	} else {
		// if migrating, just put the call to wait
		if len(entityDispatchInfo.pendingPacketQueue) < consts.ENTITY_PENDING_PACKET_QUEUE_MAX_LEN {
			pkt.AddRefCount(1)
			entityDispatchInfo.pendingPacketQueue = append(entityDispatchInfo.pendingPacketQueue, pkt)
		} else {
			gwlog.Errorf("%s.handleCallEntityMethodFromClient %s: packet queue too long, packet dropped", service, entityID)
		}
	}

}

func (service *DispatcherService) handleDoSomethingOnSpecifiedClient(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	gid := pkt.ReadUint16()
	service.dispatcherClientOfGate(gid).SendPacket(pkt)
}

func (service *DispatcherService) handleCallFilteredClientProxies(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	service.broadcastToGateClients(pkt)
}

func (service *DispatcherService) handleMigrateRequest(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	entityID := pkt.ReadEntityID()
	spaceID := pkt.ReadEntityID()
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("Entity %s is migrating to space %s", entityID, spaceID)
	}

	// mark the entity as migrating

	spaceDispatchInfo := service.getEntityDispatcherInfoForRead(spaceID)
	var spaceLoc uint16
	if spaceDispatchInfo != nil {
		spaceLoc = spaceDispatchInfo.gameid
	}

	pkt.AppendUint16(spaceLoc) // append the space game location to the packet

	if spaceLoc > 0 { // almost true
		entityDispatchInfo := service.setEntityDispatcherInfoForWrite(entityID)

		entityDispatchInfo.blockRPC(consts.DISPATCHER_MIGRATE_TIMEOUT)
	}

	dcp.SendPacket(pkt)
}

func (service *DispatcherService) handleRealMigrate(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	// get spaceID and make sure it exists
	eid := pkt.ReadEntityID()
	targetGame := pkt.ReadUint16() // target game of migration
	// target space is not checked for existence, because we relay the packet anyway

	hasClient := pkt.ReadBool()
	var clientid common.ClientID
	if hasClient {
		clientid = pkt.ReadClientID()
	}

	// mark the eid as migrating done
	entityDispatchInfo := service.setEntityDispatcherInfoForWrite(eid)

	entityDispatchInfo.blockUntilTime = time.Time{} // mark the entity as NOT migrating
	entityDispatchInfo.gameid = targetGame
	service.targetGameOfClient[clientid] = targetGame // migrating also change target game of client

	if consts.DEBUG_CLIENTS {
		gwlog.Debugf("Target game of client %s is migrated to %v along with owner %s", clientid, targetGame, eid)
	}

	service.dispatcherClientOfGame(targetGame).SendPacket(pkt)
	// send the cached calls to target game
	service.sendPendingPackets(entityDispatchInfo)
}

func (service *DispatcherService) sendPendingPackets(entityDispatchInfo *entityDispatchInfo) {
	targetGame := entityDispatchInfo.gameid
	// send the cached calls to target game
	var pendingPackets []*netutil.Packet
	pendingPackets, entityDispatchInfo.pendingPacketQueue = entityDispatchInfo.pendingPacketQueue, nil
	for _, pkt := range pendingPackets {
		service.dispatcherClientOfGame(targetGame).SendPacket(pkt)
		pkt.Release()
	}
}

func (service *DispatcherService) broadcastToGameClients(pkt *netutil.Packet) {
	for idx, dcp := range service.gameClients {
		if dcp != nil {
			dcp.SendPacket(pkt)
		} else {
			gwlog.Errorf("Game %d is not connected to dispatcher when broadcasting", idx+1)
		}
	}
}

func (service *DispatcherService) broadcastToGateClients(pkt *netutil.Packet) {
	for idx, dcp := range service.gateClients {
		if dcp != nil {
			dcp.SendPacket(pkt)
		} else {
			gwlog.Errorf("Gate %d is not connected to dispatcher when broadcasting", idx+1)
		}
	}
}

func (service *DispatcherService) cleanupEntitiesOfGame(targetGame uint16) {
	cleanEids := entity.EntityIDSet{} // get all clean eids
	for eid, dispatchInfo := range service.entityDispatchInfos {
		if dispatchInfo.gameid == targetGame {
			cleanEids.Add(eid)
		}
	}

	// for all services whose entity is cleaned, notify all games that the service is down
	undeclaredServices := common.StringSet{}
	for serviceName, serviceEids := range service.registeredServices {
		var cleanEidsOfGame []common.EntityID
		for serviceEid := range serviceEids {
			if cleanEids.Contains(serviceEid) { // this service entity is down, tell other games
				undeclaredServices.Add(serviceName)
				cleanEidsOfGame = append(cleanEidsOfGame, serviceEid)
				service.handleServiceDown(serviceName, serviceEid)
			}
		}

		for _, eid := range cleanEidsOfGame {
			serviceEids.Del(eid)
		}
	}

	for eid := range cleanEids {
		delete(service.entityDispatchInfos, eid)
	}

	gwlog.Infof("Game %d is rebooted, %d entities cleaned, undeclare services: %s", targetGame, len(cleanEids), undeclaredServices)
}
