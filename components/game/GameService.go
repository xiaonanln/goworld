package game

import (
	"fmt"

	"time"

	"io/ioutil"

	"github.com/xiaonanln/go-xnsyncutil/xnsyncutil"
	"github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld/engine/async"
	"github.com/xiaonanln/goworld/engine/binutil"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/dispatchercluster"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/gwutils"
	"github.com/xiaonanln/goworld/engine/kvdb"
	"github.com/xiaonanln/goworld/engine/netutil"
	"github.com/xiaonanln/goworld/engine/post"
	"github.com/xiaonanln/goworld/engine/proto"
	"github.com/xiaonanln/goworld/engine/srvdis"
)

const (
	rsNotRunning = iota
	rsRunning
	rsTerminating
	rsTerminated
	rsFreezing
	rsFreezed
)

type GameService struct {
	config *config.GameConfig
	id     uint16
	//registeredServices map[string]common.EntityIDSet

	packetQueue                    chan proto.Message
	runState                       xnsyncutil.AtomicInt
	nextCollectEntitySyncInfosTime time.Time
	dispatcherStartFreezeAcks      []bool
	positionSyncInterval           time.Duration
	ticker                         <-chan time.Time
	isGameConnected                []bool
	//collectEntitySyncInfosRequest chan struct{}
	//collectEntitySycnInfosReply   chan interface{}
}

func newGameService(gameid uint16) *GameService {
	//cfg := config.GetGame(gameid)
	totalGameNum := config.GetGamesNum()
	return &GameService{
		id: gameid,
		//registeredServices: map[string]common.EntityIDSet{},
		packetQueue:     make(chan proto.Message, consts.GAME_SERVICE_PACKET_QUEUE_SIZE),
		ticker:          time.Tick(consts.GAME_SERVICE_TICK_INTERVAL),
		isGameConnected: make([]bool, totalGameNum),
		//terminated:         xnsyncutil.NewOneTimeCond(),
		//dumpNotify:         xnsyncutil.NewOneTimeCond(),
		//dumpFinishedNotify: xnsyncutil.NewOneTimeCond(),
		//collectEntitySyncInfosRequest: make(chan struct{}),
		//collectEntitySycnInfosReply:   make(chan interface{}),
	}
}

func (gs *GameService) run(restore bool) {
	gs.runState.Store(rsRunning)

	if !restore {
		entity.CreateNilSpace(gameid) // create the nil space
	} else {
		// restoring from freezed states
		err := gs.doRestore()
		if err != nil {
			gwlog.Fatalf("Restore from freezed states failed: %+v", err)
		}
	}

	binutil.PrintSupervisorTag(consts.GAME_STARTED_TAG)
	gwutils.RepeatUntilPanicless(gs.serveRoutine)
}

func (gs *GameService) serveRoutine() {
	cfg := config.GetGame(gameid)
	gs.config = cfg
	gs.positionSyncInterval = time.Millisecond * time.Duration(cfg.PositionSyncIntervalMS)
	if gs.positionSyncInterval < consts.GAME_SERVICE_TICK_INTERVAL {
		gwlog.Warnf("%s: entity position sync interval is too small: %s, so reset to %s", gs, gs.positionSyncInterval, consts.GAME_SERVICE_TICK_INTERVAL)
		gs.positionSyncInterval = consts.GAME_SERVICE_TICK_INTERVAL
	}

	gwlog.Infof("Read game %d config: \n%s\n", gameid, config.DumpPretty(cfg))

	// here begins the main loop of Game
	for {
		isTick := false
		select {
		case item := <-gs.packetQueue:
			msgtype, pkt := item.MsgType, item.Packet
			switch msgtype {
			case proto.MT_SYNC_POSITION_YAW_FROM_CLIENT:
				gs.HandleSyncPositionYawFromClient(pkt)
			case proto.MT_CALL_ENTITY_METHOD_FROM_CLIENT:
				eid := pkt.ReadEntityID()
				method := pkt.ReadVarStr()
				args := pkt.ReadArgs()
				clientid := pkt.ReadClientID()
				gs.HandleCallEntityMethod(eid, method, args, clientid)
			case proto.MT_CALL_ENTITY_METHOD:
				eid := pkt.ReadEntityID()
				method := pkt.ReadVarStr()
				args := pkt.ReadArgs()
				gs.HandleCallEntityMethod(eid, method, args, "")
			case proto.MT_QUERY_SPACE_GAMEID_FOR_MIGRATE_ACK:
				gs.HandleQuerySpaceGameIDForMigrateAck(pkt)
			case proto.MT_MIGRATE_REQUEST_ACK:
				gs.HandleMigrateRequestAck(pkt)
			case proto.MT_REAL_MIGRATE:
				gs.HandleRealMigrate(pkt)
			case proto.MT_NOTIFY_CLIENT_CONNECTED:
				clientid := pkt.ReadClientID()
				gid := pkt.ReadUint16()
				gs.HandleNotifyClientConnected(clientid, gid)
			case proto.MT_NOTIFY_CLIENT_DISCONNECTED:
				clientid := pkt.ReadClientID()
				gs.HandleNotifyClientDisconnected(clientid)
			case proto.MT_LOAD_ENTITY_ANYWHERE:
				eid := pkt.ReadEntityID()
				typeName := pkt.ReadVarStr()
				gs.HandleLoadEntityAnywhere(typeName, eid)
			case proto.MT_CREATE_ENTITY_ANYWHERE:
				entityid := pkt.ReadEntityID()
				typeName := pkt.ReadVarStr()
				var data map[string]interface{}
				pkt.ReadData(&data)
				gs.HandleCreateEntityAnywhere(entityid, typeName, data)
			case proto.MT_CALL_NIL_SPACES:
				_ = pkt.ReadUint16() // ignore except gameid
				method := pkt.ReadVarStr()
				args := pkt.ReadArgs()
				gs.HandleCallNilSpaces(method, args)
			case proto.MT_SRVDIS_REGISTER:
				gs.HandleSrvdisRegister(pkt)
			case proto.MT_UNDECLARE_SERVICE:
				eid := pkt.ReadEntityID()
				serviceName := pkt.ReadVarStr()
				gs.HandleUndeclareService(eid, serviceName)
				//case proto.MT_NOTIFY_ALL_GAMES_CONNECTED:
				//	gs.handleNotifyAllGamesConnected()
			case proto.MT_NOTIFY_GATE_DISCONNECTED:
				gateid := pkt.ReadUint16()
				gs.HandleGateDisconnected(gateid)
			case proto.MT_START_FREEZE_GAME_ACK:
				dispid := pkt.ReadUint16()
				gs.HandleStartFreezeGameAck(dispid)
			case proto.MT_NOTIFY_GAME_CONNECTED:
				gs.handleNotifyGameConnected(pkt)
			case proto.MT_SET_GAME_ID_ACK:
				gs.handleSetGameIDAck(pkt)
			default:
				gwlog.TraceError("unknown msgtype: %v", msgtype)
			}

			pkt.Release()
		case <-gs.ticker:
			isTick = true
			runState := gs.runState.Load()
			if runState == rsTerminating {
				// game is terminating, run the terminating process
				gs.doTerminate()
			} else if runState == rsFreezing {
				//game is freezing, run freeze process
				gs.doFreeze()
			}

			timer.Tick()

			//case <-gs.collectEntitySyncInfosRequest: //
			//	gs.collectEntitySycnInfosReply <- 1
		}

		// after handling packets or firing timers, check the posted functions
		post.Tick()
		if isTick {
			now := time.Now()
			if !gs.nextCollectEntitySyncInfosTime.After(now) {
				gs.nextCollectEntitySyncInfosTime = now.Add(gs.positionSyncInterval)
				entity.CollectEntitySyncInfos()
			}
		}
	}
}

func (gs *GameService) waitPostsComplete() {
	gwlog.Infof("waiting for posts to complete ...")
	post.Tick() // just tick is Ok, tick will consume all posts
}

func (gs *GameService) doTerminate() {
	// wait for all posts to complete
	gwlog.Infof("Waiting for posts to complete ...")
	gs.waitPostsComplete()
	// wait for all async to clear
	gwlog.Infof("Waiting for async tasks to complete ...")
	for async.WaitClear() { // wait for all async to stop
		gs.waitPostsComplete()
	}

	// destroy all entities
	gwlog.Infof("Destroying all entities ...")
	entity.OnGameTerminating()
	gwlog.Infof("All entities saved & destroyed, game service terminated.")
	gs.runState.Store(rsTerminated)

	for {
		time.Sleep(time.Second)
	}
}

var freezePacker = netutil.MessagePackMsgPacker{}

func (gs *GameService) doFreeze() {
	// wait for all posts to complete
	st := time.Now()
	gs.waitPostsComplete()

	// wait for all async to clear
	for async.WaitClear() { // wait for all async to stop
		gs.waitPostsComplete()
	}
	gwlog.Infof("wait async & posts clear takes %s", time.Now().Sub(st))

	// destroy all entities
	freeze := func() error {
		st = time.Now()
		freezeEntity, err := entity.Freeze(gameid)
		if err != nil {
			return err
		}
		gwlog.Infof("freeze entities takes %s", time.Now().Sub(st))
		st = time.Now()
		freezeData, err := freezePacker.PackMsg(freezeEntity, nil)
		if err != nil {
			return err
		}
		gwlog.Infof("pack entities takes %s, total data size: %d", time.Now().Sub(st), len(freezeData))
		st = time.Now()
		freezeFilename := freezeFilename(gameid)
		err = ioutil.WriteFile(freezeFilename, freezeData, 0644)
		if err != nil {
			return err
		}
		gwlog.Infof("write freeze data to file takes %s", time.Now().Sub(st))

		return nil
	}

	err := freeze()
	if err != nil {
		gwlog.Errorf("Game freeze failed: %s, server has to quit", err)
		kvdb.Initialize() // restore kvdb module
		gs.runState.Store(rsRunning)
		return
	}

	gwlog.Infof("All entities saved & freezed, game service terminated.")
	gs.runState.Store(rsFreezed)
	for {
		time.Sleep(time.Second)
	}
}

func freezeFilename(gameid uint16) string {
	return fmt.Sprintf("game%d_freezed.dat", gameid)
}

func (gs *GameService) doRestore() error {
	t0 := time.Now()
	freezeFilename := freezeFilename(gameid)
	data, err := ioutil.ReadFile(freezeFilename)
	if err != nil {
		return err
	}

	t1 := time.Now()
	var freezeEntity entity.FreezeData
	freezePacker.UnpackMsg(data, &freezeEntity)
	t2 := time.Now()

	err = entity.RestoreFreezedEntities(&freezeEntity)
	t3 := time.Now()

	gwlog.Infof("Restored game service: load = %s, unpack = %s, restore = %s", t1.Sub(t0), t2.Sub(t1), t3.Sub(t2))
	return err
}

func (gs *GameService) String() string {
	return fmt.Sprintf("GameService<%d>", gs.id)
}

func (gs *GameService) HandleCreateEntityAnywhere(entityid common.EntityID, typeName string, data map[string]interface{}) {
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleCreateEntityAnywhere: %s, typeName=%s, data=%v", gs, entityid, typeName, data)
	}
	entity.OnCreateEntityAnywhere(entityid, typeName, data)
}

func (gs *GameService) HandleLoadEntityAnywhere(typeName string, entityID common.EntityID) {
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleLoadEntityAnywhere: typeName=%s, entityID=%s", gs, typeName, entityID)
	}
	entity.LoadEntityLocally(typeName, entityID)
}

func (gs *GameService) HandleSrvdisRegister(pkt *netutil.Packet) {
	// tell the entity that it is registered successfully
	srvid := pkt.ReadVarStr()
	srvinfo := pkt.ReadVarStr()
	force := pkt.ReadBool() // force is not useful here
	gwlog.Infof("%s srvdis register: %s => %s, force %v", gs, srvid, srvinfo, force)

	srvdis.WatchSrvdisRegister(srvid, srvinfo)
}

func (gs *GameService) HandleUndeclareService(entityID common.EntityID, serviceName string) {
	// tell the entity that it is registered successfully
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.HandleUndeclareService: %s undeclares %s", gs, entityID, serviceName)
	}
	entity.OnUndeclareService(serviceName, entityID)
}

//func (gs *GameService) handleNotifyAllGamesConnected() {
//	// all games are connected
//	entity.OnAllGamesConnected()
//}

func (gs *GameService) HandleGateDisconnected(gateid uint16) {
	entity.OnGateDisconnected(gateid)
}

func (gs *GameService) HandleStartFreezeGameAck(dispid uint16) {
	gwlog.Infof("Start freeze game ACK of dispatcher %d is received, checking ...", dispid)
	gs.dispatcherStartFreezeAcks[dispid-1] = true
	for _, acked := range gs.dispatcherStartFreezeAcks {
		if !acked {
			return
		}
	}
	// all acks are received, enter freezing state ...
	gs.runState.Store(rsFreezing)
}

func (gs *GameService) handleNotifyGameConnected(pkt *netutil.Packet) {
	gameid := pkt.ReadUint16() // the new connected game
	if int(gameid) > config.GetGamesNum() {
		gwlog.Panicf("%s: handle notify game connected: gameid %d is out of range", gs, gameid)
		return
	}

	if gs.isGameConnected[gameid-1] {
		// should not happen
		gwlog.Errorf("%s: handle notify game connected: game%d is connected, but it was already connected", gs, gameid)
		return
	}

	gs.isGameConnected[gameid-1] = true
	if gs.isAllGamesConnected() {
		entity.OnAllGamesConnected()
	}
}

func (gs *GameService) isAllGamesConnected() bool {
	for _, connected := range gs.isGameConnected {
		if !connected {
			return false
		}
	}
	return true
}

func (gs *GameService) handleSetGameIDAck(pkt *netutil.Packet) {
	gameNum := int(pkt.ReadUint16())
	for i := range gs.isGameConnected {
		gs.isGameConnected[i] = false // set all games to be not connected
	}
	numConnectedGames := 0
	for i := 0; i < gameNum; i++ {
		gameid := pkt.ReadUint16()
		if int(gameid) > len(gs.isGameConnected) {
			gwlog.Errorf("%s: set game ID ack: gameid %s is out of range", gs, gameid)
			continue
		}
		gs.isGameConnected[gameid-1] = true
		numConnectedGames += 1
	}

	rejectEntitiesNum := pkt.ReadUint32()
	rejectEntities := make([]common.EntityID, 0, rejectEntitiesNum)
	for i := uint32(0); i < rejectEntitiesNum; i++ {
		rejectEntities = append(rejectEntities, pkt.ReadEntityID())
	}

	srvdisMap := pkt.ReadMapStringString()
	for srvid, srvinfo := range srvdisMap {
		srvdis.WatchSrvdisRegister(srvid, srvinfo)
	}

	gwlog.Infof("%s: set game ID ack received, connected games: %v, reject entities: %d, srvdis map: %+v", gs, numConnectedGames, rejectEntitiesNum, srvdisMap)
	if gs.isAllGamesConnected() {
		// all games are connected
		entity.OnAllGamesConnected()
	}
}

func (gs *GameService) HandleSyncPositionYawFromClient(pkt *netutil.Packet) {
	//gwlog.Infof("handleSyncPositionYawFromClient: payload %d", len(pkt.UnreadPayload()))
	payload := pkt.UnreadPayload()
	payloadLen := len(payload)
	for i := 0; i < payloadLen; i += proto.SYNC_INFO_SIZE_PER_ENTITY + common.ENTITYID_LENGTH {
		eid := common.EntityID(payload[i : i+common.ENTITYID_LENGTH])
		x := netutil.UnpackFloat32(netutil.NETWORK_ENDIAN, payload[i+common.ENTITYID_LENGTH:i+common.ENTITYID_LENGTH+4])
		y := netutil.UnpackFloat32(netutil.NETWORK_ENDIAN, payload[i+common.ENTITYID_LENGTH+4:i+common.ENTITYID_LENGTH+8])
		z := netutil.UnpackFloat32(netutil.NETWORK_ENDIAN, payload[i+common.ENTITYID_LENGTH+8:i+common.ENTITYID_LENGTH+12])
		yaw := netutil.UnpackFloat32(netutil.NETWORK_ENDIAN, payload[i+common.ENTITYID_LENGTH+12:i+common.ENTITYID_LENGTH+16])
		entity.OnSyncPositionYawFromClient(eid, entity.Coord(x), entity.Coord(y), entity.Coord(z), entity.Yaw(yaw))
	}
}

func (gs *GameService) HandleCallEntityMethod(entityID common.EntityID, method string, args [][]byte, clientid common.ClientID) {
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleCallEntityMethod: %s.%s(%v)", gs, entityID, method, args)
	}
	entity.OnCall(entityID, method, args, clientid)
}

func (gs *GameService) HandleNotifyClientConnected(clientid common.ClientID, gateid uint16) {
	client := entity.MakeGameClient(clientid, gateid)
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleNotifyClientConnected: %s", gs, client)
	}

	// create a boot entity for the new client and set the client as the OWN CLIENT of the entity
	entity.CreateEntityLocally(gs.config.BootEntity, nil, client)
}

func (gs *GameService) HandleCallNilSpaces(method string, args [][]byte) {
	gwlog.Infof("%s.HandleCallNilSpaces: method=%s, argcount=%d", gs, method, len(args))
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.HandleCallNilSpaces: method=%s, argcount=%d", gs, method, len(args))
	}

	entity.OnCallNilSpaces(method, args)
}

func (gs *GameService) HandleNotifyClientDisconnected(clientid common.ClientID) {
	if consts.DEBUG_CLIENTS {
		gwlog.Debugf("%s.handleNotifyClientDisconnected: %s", gs, clientid)
	}
	// find the owner of the client, and notify lose client
	entity.OnClientDisconnected(clientid)
}

func (gs *GameService) HandleQuerySpaceGameIDForMigrateAck(pkt *netutil.Packet) {
	spaceid := pkt.ReadEntityID()
	entityid := pkt.ReadEntityID()
	gameid := pkt.ReadUint16()
	entity.OnQuerySpaceGameIDForMigrateAck(entityid, spaceid, gameid)
}

func (gs *GameService) HandleMigrateRequestAck(pkt *netutil.Packet) {
	eid := pkt.ReadEntityID()
	spaceid := pkt.ReadEntityID()
	spaceLoc := pkt.ReadUint16()

	if consts.DEBUG_PACKETS {
		gwlog.Debugf("Entity %s is migrating to space %s at game %d", eid, spaceid, spaceLoc)
	}

	entity.OnMigrateRequestAck(eid, spaceid, spaceLoc)
}

func (gs *GameService) HandleRealMigrate(pkt *netutil.Packet) {
	eid := pkt.ReadEntityID()
	_ = pkt.ReadUint16() // targetGame is not userful

	hasClient := pkt.ReadBool()
	var clientid common.ClientID
	var clientsrv uint16
	if hasClient {
		clientid = pkt.ReadClientID()
		clientsrv = pkt.ReadUint16()
	}

	spaceID := pkt.ReadEntityID() // target space
	x := pkt.ReadFloat32()
	y := pkt.ReadFloat32()
	z := pkt.ReadFloat32()
	typeName := pkt.ReadVarStr()
	var migrateData map[string]interface{}
	pkt.ReadData(&migrateData)
	timerData := pkt.ReadVarBytes()
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleRealMigrate: entity %s migrating to space %s, typeName=%s, migrateData=%v, timerData=%v, client=%s@%d", gs, eid, spaceID, typeName, migrateData, timerData, clientid, clientsrv)
	}

	entity.OnRealMigrate(eid, spaceID, x, y, z, typeName, migrateData, timerData, clientid, clientsrv)
}

func (gs *GameService) terminate() {
	gs.runState.Store(rsTerminating)
}

func (gs *GameService) startFreeze() {
	dispatcherNum := len(config.GetDispatcherIDs())
	gs.dispatcherStartFreezeAcks = make([]bool, dispatcherNum)
	dispatchercluster.SendStartFreezeGame(gameid)
}

func ConnectedGamesNum() int {
	return len(gameService.isGameConnected)
}
