package main

import (
	"fmt"
	"github.com/xiaonanln/pktconn"

	"net"

	"time"

	"os"

	"math/rand"

	"container/heap"

	"github.com/xiaonanln/goworld/engine/binutil"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/gwutils"
	"github.com/xiaonanln/goworld/engine/netutil"
	"github.com/xiaonanln/goworld/engine/post"
	"github.com/xiaonanln/goworld/engine/proto"
)

type entityDispatchInfo struct {
	gameid             uint16
	blockUntilTime     time.Time
	pendingPacketQueue []*netutil.Packet
}

func (edi *entityDispatchInfo) blockRPC(d time.Duration) {
	t := time.Now().Add(d)
	if edi.blockUntilTime.Before(t) {
		edi.blockUntilTime = t
	}
}

func (edi *entityDispatchInfo) dispatchPacket(pkt *netutil.Packet) {
	if edi.blockUntilTime.IsZero() {
		// most common case. handle it quickly
		dispatcherService.dispatchPacketToGame(edi.gameid, pkt)
	}

	// blockUntilTime is set, need to check if block should be released
	now := time.Now()
	if now.Before(edi.blockUntilTime) {
		// keep blocking, just put the call to wait
		if len(edi.pendingPacketQueue) < consts.ENTITY_PENDING_PACKET_QUEUE_MAX_LEN {
			pkt.Retain()
			edi.pendingPacketQueue = append(edi.pendingPacketQueue, pkt)
		} else {
			gwlog.Errorf("%s.dispatchPacket: packet queue too long, packet dropped", edi)
		}
	} else {
		// time to unblock
		edi.unblock()
	}
}

func (info *entityDispatchInfo) unblock() {
	if !info.blockUntilTime.IsZero() { // entity is loading, it's done now
		//gwlog.Infof("entity is loaded now, clear loadTime")
		info.blockUntilTime = time.Time{}

		targetGame := info.gameid
		// send the cached calls to target game
		var pendingPackets []*netutil.Packet
		pendingPackets, info.pendingPacketQueue = info.pendingPacketQueue, nil
		for _, pkt := range pendingPackets {
			dispatcherService.dispatchPacketToGame(targetGame, pkt)
			pkt.Release()
		}
	}
}

type gameDispatchInfo struct {
	gameid             uint16
	clientProxy        *dispatcherClientProxy
	isBlocked          bool
	blockUntilTime     time.Time // game can be blocked
	pendingPacketQueue []*netutil.Packet
	isBanBootEntity    bool
	lbcheapentry       *lbcheapentry
}

func (gdi *gameDispatchInfo) setClientProxy(clientProxy *dispatcherClientProxy) {
	gdi.clientProxy = clientProxy
	if gdi.clientProxy != nil && !gdi.isBlocked {
		gdi.sendPendingPackets()
	}
	return
}

func (gdi *gameDispatchInfo) block(duration time.Duration) {
	gdi.isBlocked = true
	gdi.blockUntilTime = time.Now().Add(duration)
}

func (gdi *gameDispatchInfo) checkBlocked() bool {
	if gdi.isBlocked {
		if time.Now().After(gdi.blockUntilTime) {
			gdi.isBlocked = false
			return true
		}
	}
	return false
}

func (gdi *gameDispatchInfo) isConnected() bool {
	return gdi.clientProxy != nil
}

func (gdi *gameDispatchInfo) dispatchPacket(pkt *netutil.Packet) {
	if gdi.checkBlocked() && gdi.clientProxy == nil {
		// blocked from true -> false, and game is already disconnected before
		// in this case, the game should be cleaned up
		dispatcherService.handleGameDown(gdi)
	}

	if !gdi.isBlocked && gdi.clientProxy != nil {
		gdi.clientProxy.SendPacket(pkt)
	} else {
		if len(gdi.pendingPacketQueue) < consts.GAME_PENDING_PACKET_QUEUE_MAX_LEN {
			gdi.pendingPacketQueue = append(gdi.pendingPacketQueue, pkt)
			pkt.Retain()

			if len(gdi.pendingPacketQueue)%1 == 0 {
				gwlog.Warnf("game %d pending packet count = %d, blocked = %v, clientProxy = %s", gdi.gameid, len(gdi.pendingPacketQueue), gdi.isBlocked, gdi.clientProxy)
			}
		} else {
			gwlog.Errorf("packet to game %d is dropped", gdi.gameid)
		}
	}
}

func (gdi *gameDispatchInfo) unblock() {
	if !gdi.blockUntilTime.IsZero() {
		gdi.blockUntilTime = time.Time{}

		if gdi.clientProxy != nil {
			gdi.sendPendingPackets()
		}
	}
}

func (gdi *gameDispatchInfo) sendPendingPackets() {
	// send the cached calls to target game
	var pendingPackets []*netutil.Packet
	pendingPackets, gdi.pendingPacketQueue = gdi.pendingPacketQueue, nil
	for _, pkt := range pendingPackets {
		gdi.clientProxy.SendPacket(pkt)
		pkt.Release()
	}
}

func (gdi *gameDispatchInfo) clearPendingPackets() {
	var pendingPackets []*netutil.Packet
	pendingPackets, gdi.pendingPacketQueue = gdi.pendingPacketQueue, nil
	for _, pkt := range pendingPackets {
		pkt.Release()
	}
}

// DispatcherService implements the dispatcher service
type DispatcherService struct {
	dispid                uint16
	config                *config.DispatcherConfig
	games                 map[uint16]*gameDispatchInfo
	bootGames             []uint16
	gates                 map[uint16]*dispatcherClientProxy
	messageQueue          chan *pktconn.Packet
	entityDispatchInfos   map[common.EntityID]*entityDispatchInfo
	kvregRegisterMap      map[string]string
	entitySyncInfosToGame map[uint16]*netutil.Packet // cache entity sync infos to gates
	ticker                <-chan time.Time
	lbcheap               lbcheap // heap for game load balancing
	chooseGameIdx         int     // choose game in a round robin way
	isDeploymentReady     bool    // whether or not the deployment is ready
}

func newDispatcherService(dispid uint16) *DispatcherService {
	cfg := config.GetDispatcher(dispid)
	ds := &DispatcherService{
		dispid:                dispid,
		config:                cfg,
		messageQueue:          make(chan *pktconn.Packet, consts.DISPATCHER_SERVICE_PACKET_QUEUE_SIZE),
		games:                 map[uint16]*gameDispatchInfo{},
		gates:                 map[uint16]*dispatcherClientProxy{},
		entityDispatchInfos:   map[common.EntityID]*entityDispatchInfo{},
		kvregRegisterMap:      map[string]string{},
		entitySyncInfosToGame: map[uint16]*netutil.Packet{},
		ticker:                time.Tick(consts.DISPATCHER_SERVICE_TICK_INTERVAL),
		lbcheap:               nil,
		isDeploymentReady:     false,
	}

	ds.recalcBootGames()

	return ds
}

func (service *DispatcherService) messageLoop() {
	for {
		select {
		case _pkt := <-service.messageQueue:
			pkt := (*netutil.Packet)(_pkt)
			dcp := pkt.Src.Tag.(*dispatcherClientProxy)

			msgtype := proto.MsgType(pkt.ReadUint16())

			if msgtype >= proto.MT_REDIRECT_TO_GATEPROXY_MSG_TYPE_START && msgtype <= proto.MT_REDIRECT_TO_GATEPROXY_MSG_TYPE_STOP {
				service.handleDoSomethingOnSpecifiedClient(dcp, pkt)
			} else {
				switch msgtype {
				case proto.MT_SYNC_POSITION_YAW_FROM_CLIENT:
					service.handleSyncPositionYawFromClient(dcp, pkt)
				case proto.MT_SYNC_POSITION_YAW_ON_CLIENTS:
					service.handleSyncPositionYawOnClients(dcp, pkt)
				case proto.MT_CALL_ENTITY_METHOD:
					service.handleCallEntityMethod(dcp, pkt)
				case proto.MT_CALL_ENTITY_METHOD_FROM_CLIENT:
					service.handleCallEntityMethodFromClient(dcp, pkt)
				case proto.MT_QUERY_SPACE_GAMEID_FOR_MIGRATE:
					service.handleQuerySpaceGameIDForMigrate(dcp, pkt)
				case proto.MT_MIGRATE_REQUEST:
					service.handleMigrateRequest(dcp, pkt)
				case proto.MT_REAL_MIGRATE:
					service.handleRealMigrate(dcp, pkt)
				case proto.MT_CALL_FILTERED_CLIENTS:
					service.handleCallFilteredClientProxies(dcp, pkt)
				case proto.MT_NOTIFY_CLIENT_CONNECTED:
					service.handleNotifyClientConnected(dcp, pkt)
				case proto.MT_NOTIFY_CLIENT_DISCONNECTED:
					service.handleNotifyClientDisconnected(dcp, pkt)
				case proto.MT_LOAD_ENTITY_SOMEWHERE:
					service.handleLoadEntitySomewhere(dcp, pkt)
				case proto.MT_NOTIFY_CREATE_ENTITY:
					eid := pkt.ReadEntityID()
					service.handleNotifyCreateEntity(dcp, pkt, eid)
				case proto.MT_NOTIFY_DESTROY_ENTITY:
					eid := pkt.ReadEntityID()
					service.handleNotifyDestroyEntity(dcp, pkt, eid)
				case proto.MT_CREATE_ENTITY_SOMEWHERE:
					service.handleCreateEntitySomewhere(dcp, pkt)
				case proto.MT_GAME_LBC_INFO:
					service.handleGameLBCInfo(dcp, pkt)
				case proto.MT_CALL_NIL_SPACES:
					service.handleCallNilSpaces(dcp, pkt)
				case proto.MT_CANCEL_MIGRATE:
					service.handleCancelMigrate(dcp, pkt)
				case proto.MT_KVREG_REGISTER:
					service.handleKvregRegister(dcp, pkt)
				case proto.MT_SET_GAME_ID:
					// this is a game server
					service.handleSetGameID(dcp, pkt)
				case proto.MT_SET_GATE_ID:
					// this is a gate
					service.handleSetGateID(dcp, pkt)
				case proto.MT_START_FREEZE_GAME:
					// freeze the game
					service.handleStartFreezeGame(dcp, pkt)
				default:
					gwlog.TraceError("unknown msgtype %d from %s", msgtype, dcp)
				}
			}

			pkt.Release()
			break
		case <-service.ticker:
			post.Tick()
			service.sendEntitySyncInfosToGames()
			break
		}
	}
}

func (service *DispatcherService) terminate() {
	gwlog.Infof("Dispatcher terminated gracefully.")
	os.Exit(0)
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
	return fmt.Sprintf("DispatcherService<%d>", dispid)
}

func (service *DispatcherService) run() {
	binutil.PrintSupervisorTag(consts.DISPATCHER_STARTED_TAG)
	go gwutils.RepeatUntilPanicless(service.messageLoop)
	netutil.ServeTCPForever(service.config.ListenAddr, service)
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
	isBanBootEntity := pkt.ReadBool()
	numEntities := pkt.ReadUint32() // number of entities on the game
	//for i := uint32(0); i < numEntities; i++ {
	//	eid := pkt.ReadEntityID()
	//}

	gwlog.Infof("%s: connection %s set gameid=%d, isReconnect=%v, isRestore=%v, isBanBootEntity=%v, numEntities=%d", service, dcp, gameid, isReconnect, isRestore, isBanBootEntity, numEntities)

	if gameid <= 0 {
		gwlog.Panicf("invalid gameid: %d", gameid)
	}
	if dcp.gameid > 0 || dcp.gateid > 0 {
		gwlog.Panicf("already set gameid=%d, gateid=%d", dcp.gameid, dcp.gateid)
	}
	dcp.gameid = gameid

	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleSetGameID: dcp=%s, gameid=%d, isReconnect=%v", service, dcp, gameid, isReconnect)
	}

	gdi := service.games[gameid]
	if gdi == nil {
		// new game connected, create dispatch info for the game
		lbcheapentry := &lbcheapentry{gameid, len(service.lbcheap), 0, 0}
		gdi = &gameDispatchInfo{gameid: gameid, isBanBootEntity: isBanBootEntity, lbcheapentry: lbcheapentry}
		service.games[gameid] = gdi
		heap.Push(&service.lbcheap, lbcheapentry)
		service.lbcheap.validateHeapIndexes()

		if !isBanBootEntity {
			service.bootGames = append(service.bootGames, gameid)
		}
	} else if gdi.clientProxy != nil {
		gdi.clientProxy.Close()
		service.handleGameDisconnected(gdi.clientProxy)
	}

	oldIsBanBootEntity := gdi.isBanBootEntity
	gdi.isBanBootEntity = isBanBootEntity
	gdi.setClientProxy(dcp) // should be nil, unless reconnect
	gdi.unblock()           // unlock game dispatch info if new game is connected
	if oldIsBanBootEntity != isBanBootEntity {
		service.recalcBootGames() // recalc if necessary
	}

	// restore all entities for the game from the packet
	var rejectEntities []common.EntityID
	for i := uint32(0); i < numEntities; i++ {
		eid := pkt.ReadEntityID()
		edi := service.setEntityDispatcherInfoForWrite(eid)
		if edi.gameid == gameid {
			// the current game for the entity is not changed
			edi.unblock()
		} else if edi.gameid == 0 {
			// the entity has no game yet, set to this game
			edi.gameid = gameid
			edi.unblock()
		} else {
			// the entity is on other game ... need to tell the game to destroy his version of entity
			rejectEntities = append(rejectEntities, eid)
		}
	}

	gwlog.Infof("%s: %s set gameid = %d, numEntities = %d, rejectEntites = %d, services = %v", service, dcp, gameid, numEntities, len(rejectEntities), service.kvregRegisterMap)
	// reuse the packet to send SET_GAMEID_ACK with all connected gameids
	connectedGameIDs := service.getConnectedGameIDs()

	dcp.SendSetGameIDAck(service.dispid, service.isDeploymentReady, connectedGameIDs, rejectEntities, service.kvregRegisterMap)
	service.sendNotifyGameConnected(gameid)
	service.checkDeploymentReady()
	return
}

func (service *DispatcherService) getConnectedGameIDs() (gameids []uint16) {
	for _, gdi := range service.games {
		if gdi.clientProxy != nil {
			gameids = append(gameids, gdi.gameid)
		}
	}
	return
}

//
//func (service *DispatcherService) sendSetGameIDAck(pkt *netutil.Packet) {
//}

func (service *DispatcherService) sendNotifyGameConnected(gameid uint16) {
	pkt := proto.MakeNotifyGameConnectedPacket(gameid)
	service.broadcastToGamesExcept(pkt, gameid)
	pkt.Release()
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
	gwlog.Infof("Gate %d is connected: %s", gateid, dcp)

	olddcp := service.gates[gateid]
	if olddcp != nil {
		gwlog.Warnf("Gate %d connection %s is replaced by new connection %s", gateid, olddcp, dcp)
		olddcp.Close()
		service.handleGateDisconnected(olddcp)
	}

	service.gates[gateid] = dcp
	service.checkDeploymentReady()
}

func (service *DispatcherService) checkDeploymentReady() {
	if service.isDeploymentReady {
		// if deployment was ever ready, it is already ready forever
		return
	}

	deployCfg := config.GetDeployment()
	numGates := len(service.gates)
	if numGates < deployCfg.DesiredGates {
		gwlog.Infof("%s check deployment ready: %d/%d gates", service, numGates, deployCfg.DesiredGates)
		return
	}
	numGames := 0
	for _, gdi := range service.games {
		if gdi.isBlocked || gdi.clientProxy != nil {
			numGames += 1
		}
	}
	gwlog.Infof("%s check deployment ready: %d/%d games %d/%d gates", service, numGames, deployCfg.DesiredGames, numGates, deployCfg.DesiredGates)
	if numGames < deployCfg.DesiredGames {
		// games not ready
		return
	}

	// now deployment is ready
	service.isDeploymentReady = true
	// now the deployment is ready for only once
	// broadcast deployment ready to all games
	pkt := proto.MakeNotifyDeploymentReadyPacket()
	service.broadcastToGamesRelease(pkt)
}

func (service *DispatcherService) handleStartFreezeGame(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	// freeze the game, which block all entities of that game
	gwlog.Infof("Handling start freeze game ...")
	gameid := dcp.gameid
	gdi := service.games[gameid]
	if gdi == nil {
		gwlog.Panicf("%s handleStartFreezeGame: game%d not found", service, gameid)
		return
	}

	gdi.block(consts.DISPATCHER_FREEZE_GAME_TIMEOUT)

	// tell the game to start real freeze, re-using the packet
	pkt.ClearPayload()
	pkt.AppendUint16(proto.MT_START_FREEZE_GAME_ACK)
	pkt.AppendUint16(dispid)
	dcp.SendPacket(pkt)
}

func (service *DispatcherService) isAllGameClientsConnected() bool {
	for _, gdi := range service.games {
		if !gdi.isConnected() {
			return false
		}
	}
	return true
}

func (service *DispatcherService) connectedGameClientsNum() int {
	num := 0
	for _, gdi := range service.games {
		if gdi.isConnected() {
			num += 1
		}
	}
	return num
}

func (service *DispatcherService) dispatchPacketToGame(gameid uint16, pkt *netutil.Packet) {
	gdi := service.games[gameid]
	if gdi != nil {
		gdi.dispatchPacket(pkt)
	} else {
		gwlog.Errorf("%s: dispatchPacketToGame: game%d is not found", service, gameid)
	}
}

func (service *DispatcherService) dispatcherClientOfGate(gateid uint16) *dispatcherClientProxy {
	return service.gates[gateid]
}

// Choose a dispatcher client for sending Anywhere packets
func (service *DispatcherService) chooseGame() *gameDispatchInfo {
	if len(service.lbcheap) == 0 {
		return nil
	}

	top := service.lbcheap[0]
	gwlog.Infof("%s: choose game by lbc: gameid=%d", service, top.gameid)
	gdi := service.games[top.gameid]

	// after game is chosen, udpate CPU percent by a bit
	service.lbcheap.chosen(0)
	service.lbcheap.validateHeapIndexes()
	return gdi
}

// Choose a dispatcher client for sending Anywhere packets
func (service *DispatcherService) chooseGameForBootEntity() *gameDispatchInfo {

	if len(service.bootGames) > 0 {
		gameid := service.bootGames[service.chooseGameIdx%len(service.bootGames)]
		service.chooseGameIdx += 1
		return service.games[gameid]
	} else {
		gwlog.Errorf("%s chooseGameForBootEntity: no game", service)
		return nil
	}
}

func (service *DispatcherService) handleDispatcherClientDisconnect(dcp *dispatcherClientProxy) {
	// nothing to do when client disconnected
	defer func() {
		if err := recover(); err != nil {
			gwlog.Errorf("handleDispatcherClientDisconnect paniced: %v", err)
		}
	}()
	gwlog.Warnf("%s disconnected", dcp)
	if dcp.gateid > 0 {
		// gate disconnected, notify all clients disconnected
		service.handleGateDisconnected(dcp)
	} else if dcp.gameid > 0 {
		service.handleGameDisconnected(dcp)
	}
}

func (service *DispatcherService) handleGateDisconnected(dcp *dispatcherClientProxy) {
	gateid := dcp.gateid
	gwlog.Warnf("Gate %d connection %s is down!", gateid, dcp)

	curdcp := service.gates[gateid]
	if curdcp != dcp {
		gwlog.Errorf("Gate %d connection %s is down, but the current connection is %s", gateid, dcp, curdcp)
		return
	}

	// should always goes here
	delete(service.gates, gateid)
	// notify all games of gate down
	pkt := netutil.NewPacket()
	pkt.AppendUint16(proto.MT_NOTIFY_GATE_DISCONNECTED)
	pkt.AppendUint16(gateid)
	service.broadcastToGamesRelease(pkt)
}

func (service *DispatcherService) handleGameDisconnected(dcp *dispatcherClientProxy) {
	gameid := dcp.gameid
	gwlog.Errorf("%s: game %d is down: %s", service, gameid, dcp)
	gdi := service.games[gameid]
	if gdi == nil {
		gwlog.Errorf("%s handleGameDisconnected: game%d is not found", service, gameid)
		return
	}

	if dcp != gdi.clientProxy {
		// this connection is not the current connection
		gwlog.Errorf("%s: game%d connection %s is disconnected, but the current connection is %s", service, gameid, dcp, gdi.clientProxy)
		return
	}

	gdi.clientProxy = nil // connection down, set clientProxy = nil
	if !gdi.isBlocked {
		// game is down, we need to clear all
		service.handleGameDown(gdi)
	} else {
		// game is freezed, wait for restore, setup a timer to cleanup later if restore is not success
	}

}

func (service *DispatcherService) handleGameDown(gdi *gameDispatchInfo) {
	gameid := gdi.gameid
	gwlog.Infof("%s: game%d is down, cleaning up...", service, gameid)
	service.cleanupEntitiesOfGame(gameid)
	gdi.clearPendingPackets()

	// send gamedown packet to all games
	service.broadcastToGamesRelease(proto.MakeNotifyGameDisconnectedPacket(gameid))
}

func (service *DispatcherService) cleanupEntitiesOfGame(gameid uint16) {
	cleanEids := common.EntityIDSet{} // get all clean eids
	for eid, dispatchInfo := range service.entityDispatchInfos {
		if dispatchInfo.gameid == gameid {
			cleanEids.Add(eid)
		}
	}

	for eid := range cleanEids {
		service.cleanupEntityInfo(eid)
	}

	gwlog.Infof("%s: game%d is down, %d entities cleaned", service, gameid, len(cleanEids))
}

// Entity is create on the target game
func (service *DispatcherService) handleNotifyCreateEntity(dcp *dispatcherClientProxy, pkt *netutil.Packet, entityID common.EntityID) {
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleNotifyCreateEntity: dcp=%s, entityID=%s", service, dcp, entityID)
	}
	entityDispatchInfo := service.setEntityDispatcherInfoForWrite(entityID)
	entityDispatchInfo.gameid = dcp.gameid
	entityDispatchInfo.unblock()
}

func (service *DispatcherService) handleNotifyDestroyEntity(dcp *dispatcherClientProxy, pkt *netutil.Packet, entityID common.EntityID) {
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleNotifyDestroyEntity: dcp=%s, entityID=%s", service, dcp, entityID)
	}
	service.cleanupEntityInfo(entityID)
}

func (service *DispatcherService) cleanupEntityInfo(entityID common.EntityID) {
	service.delEntityDispatchInfo(entityID)
}

func (service *DispatcherService) handleNotifyClientConnected(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	targetGame := service.chooseGameForBootEntity()
	pkt.AppendUint16(dcp.gateid)
	targetGame.dispatchPacket(pkt)
}

func (service *DispatcherService) handleNotifyClientDisconnected(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	ownerEntityID := pkt.ReadEntityID() // owner entity's ID for the client
	edi := service.entityDispatchInfos[ownerEntityID]
	if edi != nil {
		edi.dispatchPacket(pkt)
	} else {
		gwlog.Warnf("%s: client %s is disconnected, but owner entity %s not found", service, dcp, ownerEntityID)
	}
}

func (service *DispatcherService) handleLoadEntitySomewhere(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	//typeName := pkt.ReadVarStr()
	//eid := pkt.ReadEntityID()
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleLoadEntitySomewhere: dcp=%s, pkt=%v", service, dcp, pkt.Payload())
	}
	gameid := pkt.ReadUint16() // the target game to create entity or 0 for anywhere
	eid := pkt.ReadEntityID()  // field 1

	entityDispatchInfo := service.setEntityDispatcherInfoForWrite(eid)

	if entityDispatchInfo.gameid == 0 { // entity not loaded, try load now
		var gdi *gameDispatchInfo
		if gameid == 0 {
			gdi = service.chooseGame()
		} else {
			gdi = service.games[gameid]
		}

		if gdi != nil {
			entityDispatchInfo.gameid = gdi.gameid
			entityDispatchInfo.blockRPC(consts.DISPATCHER_LOAD_TIMEOUT)
			gdi.dispatchPacket(pkt)
		} else {
			gwlog.Errorf("%s: handleLoadEntitySomewhere: no game", service)
		}
	} else if gameid != 0 && gameid != entityDispatchInfo.gameid {
		gwlog.Warnf("%s: try to load entity on game%d, but already created on game%d", service, gameid, entityDispatchInfo.gameid)
	}
}

func (service *DispatcherService) handleCreateEntitySomewhere(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleCreateEntitySomewhere: dcp=%s, pkt=%s", service, dcp, pkt.Payload())
	}
	gameid := pkt.ReadUint16()
	entityid := pkt.ReadEntityID()
	var gdi *gameDispatchInfo
	if gameid == 0 {
		// choose a random game
		gdi = service.chooseGame()
	} else {
		// choose the specified game
		gdi = service.games[gameid]
	}

	if gdi != nil {
		entityDispatchInfo := service.setEntityDispatcherInfoForWrite(entityid)
		entityDispatchInfo.gameid = gdi.gameid // setup gameid of entity
		gdi.dispatchPacket(pkt)
	} else {
		gwlog.Errorf("%s handleCreateEntitySomewhere: no game", service)
	}
}

func (service *DispatcherService) handleKvregRegister(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	srvid := pkt.ReadVarStr()
	srvinfo := pkt.ReadVarStr()
	force := pkt.ReadBool()

	curinfo := service.kvregRegisterMap[srvid]

	if force || curinfo == "" {
		service.kvregRegisterMap[srvid] = srvinfo
		service.broadcastToGames(pkt)
		gwlog.Infof("%s: kvreg register %s = %s, force %v, register ok", service, srvid, srvinfo, force)
	} else {
		gwlog.Infof("%s: kvreg register %s = %s, force %v, curinfo=%s, register failed", service, srvid, srvinfo, force, curinfo)
	}
}

func (service *DispatcherService) handleCallEntityMethod(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	entityID := pkt.ReadEntityID()

	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleCallEntityMethod: dcp=%s, entityID=%s", service, dcp, entityID)
	}

	entityDispatchInfo := service.entityDispatchInfos[entityID]
	if entityDispatchInfo != nil {
		entityDispatchInfo.dispatchPacket(pkt)
	} else {
		gwlog.Warnf("%s: entity %s is called by other entity, but dispatch info is not found", service, entityID)
	}
}

func (service *DispatcherService) handleCallNilSpaces(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	// send the packet to all games
	exceptGameID := pkt.ReadUint16()
	service.broadcastToGamesExcept(pkt, exceptGameID)
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

		entityDispatchInfo := service.entityDispatchInfos[eid]
		if entityDispatchInfo == nil {
			gwlog.Warnf("%s: entity %s is synced from client, but dispatch info is not found", service, eid)
			continue
		}

		gameid := entityDispatchInfo.gameid

		// put this sync info to the pending queue of target game
		// concat to the end of queue
		pkt := service.entitySyncInfosToGame[gameid]
		if pkt == nil {
			pkt = netutil.NewPacket()
			pkt.AppendUint16(proto.MT_SYNC_POSITION_YAW_FROM_CLIENT)
			service.entitySyncInfosToGame[gameid] = pkt
		}
		pkt.AppendBytes(payload[i : i+proto.SYNC_INFO_SIZE_PER_ENTITY+common.ENTITYID_LENGTH])
	}
}

func (service *DispatcherService) sendEntitySyncInfosToGames() {
	if len(service.entitySyncInfosToGame) == 0 {
		return
	}

	for gameid, pkt := range service.entitySyncInfosToGame {
		// send the entity sync infos to this game
		service.games[gameid].dispatchPacket(pkt)
		pkt.Release()
	}
	service.entitySyncInfosToGame = map[uint16]*netutil.Packet{}
}

func (service *DispatcherService) handleCallEntityMethodFromClient(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	entityID := pkt.ReadEntityID()

	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleCallEntityMethodFromClient: entityID=%s, payload=%v", service, entityID, pkt.Payload())
	}

	entityDispatchInfo := service.entityDispatchInfos[entityID]
	if entityDispatchInfo != nil {
		entityDispatchInfo.dispatchPacket(pkt)
	} else {
		gwlog.Warnf("%s: entity %s is called by client, but dispatch info is not found", service, entityID)
	}
}

func (service *DispatcherService) handleDoSomethingOnSpecifiedClient(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	gid := pkt.ReadUint16()
	service.dispatcherClientOfGate(gid).SendPacket(pkt)
}

func (service *DispatcherService) handleCallFilteredClientProxies(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	service.broadcastToGates(pkt)
}

func (service *DispatcherService) handleQuerySpaceGameIDForMigrate(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	spaceid := pkt.ReadEntityID()
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleQuerySpaceGameIDForMigrate: spaceid=%s", service, spaceid)
	}

	spaceDispatchInfo := service.entityDispatchInfos[spaceid]
	var gameid uint16
	if spaceDispatchInfo != nil {
		gameid = spaceDispatchInfo.gameid
	}
	pkt.AppendUint16(gameid)
	// send the packet back
	dcp.SendPacket(pkt)
}

func (service *DispatcherService) handleMigrateRequest(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	entityID := pkt.ReadEntityID()
	spaceID := pkt.ReadEntityID()
	spaceGameID := pkt.ReadUint16()

	if consts.DEBUG_PACKETS {
		gwlog.Debugf("Entity %s is migrating to space %s @ game%d", entityID, spaceID, spaceGameID)
	}

	entityDispatchInfo := service.setEntityDispatcherInfoForWrite(entityID)
	entityDispatchInfo.blockRPC(consts.DISPATCHER_MIGRATE_TIMEOUT)
	dcp.SendPacket(pkt)
}

func (service *DispatcherService) handleCancelMigrate(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	entityid := pkt.ReadEntityID()

	if consts.DEBUG_PACKETS {
		gwlog.Debugf("Entity %s cancelled migrating", entityid)
	}

	entityDispatchInfo := service.entityDispatchInfos[entityid]
	if entityDispatchInfo != nil {
		entityDispatchInfo.unblock()
	}
}

func (service *DispatcherService) handleRealMigrate(dcp *dispatcherClientProxy, pkt *netutil.Packet) {
	// get spaceID and make sure it exists
	eid := pkt.ReadEntityID()
	targetGame := pkt.ReadUint16() // target game of migration
	// target space is not checked for existence, because we relay the packet anyway

	// mark the eid as migrating done
	entityDispatchInfo := service.setEntityDispatcherInfoForWrite(eid)

	entityDispatchInfo.gameid = targetGame

	service.dispatchPacketToGame(targetGame, pkt)
	// send the cached calls to target game
	entityDispatchInfo.unblock()
}

func (service *DispatcherService) broadcastToGames(pkt *netutil.Packet) {
	for _, gdi := range service.games {
		gdi.dispatchPacket(pkt)
	}
}
func (service *DispatcherService) broadcastToGamesRelease(pkt *netutil.Packet) {
	service.broadcastToGames(pkt)
	pkt.Release()
}
func (service *DispatcherService) broadcastToGamesExcept(pkt *netutil.Packet, exceptGameID uint16) {
	for gameid, gdi := range service.games {
		if gameid == exceptGameID {
			continue
		}
		gdi.dispatchPacket(pkt)
	}
}

func (service *DispatcherService) broadcastToGates(pkt *netutil.Packet) {
	for gateid, dcp := range service.gates {
		if dcp != nil {
			dcp.SendPacket(pkt)
		} else {
			gwlog.Errorf("Gate %d is not connected to dispatcher when broadcasting", gateid)
		}
	}
}

func (service *DispatcherService) recalcBootGames() {
	var candidates []uint16
	for gameid, gdi := range service.games {
		if !gdi.isBanBootEntity {
			candidates = append(candidates, gameid)
		}
	}
	service.bootGames = candidates
}

func (service *DispatcherService) handleGameLBCInfo(dcp *dispatcherClientProxy, packet *netutil.Packet) {
	// handle game LBC info from game
	var lbcinfo proto.GameLBCInfo
	packet.ReadData(&lbcinfo)
	gwlog.Debugf("Game %d Load Balancing Info: %+v", dcp.gameid, lbcinfo)
	lbcinfo.CPUPercent *= 1 + (rand.Float64() * 0.1) // multiply CPUPercent by a random factor 1.0 ~ 1.1
	gdi := service.games[dcp.gameid]
	gdi.lbcheapentry.update(lbcinfo)
	heap.Fix(&service.lbcheap, gdi.lbcheapentry.heapidx)
	service.lbcheap.validateHeapIndexes()
}
