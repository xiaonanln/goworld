package game

import (
	"fmt"
	"github.com/xiaonanln/pktconn"

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
	"github.com/xiaonanln/goworld/engine/gwvar"
	"github.com/xiaonanln/goworld/engine/kvdb"
	"github.com/xiaonanln/goworld/engine/kvreg"
	"github.com/xiaonanln/goworld/engine/netutil"
	"github.com/xiaonanln/goworld/engine/post"
	"github.com/xiaonanln/goworld/engine/proto"
	"github.com/xiaonanln/goworld/engine/service"
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

	packetQueue                    chan *pktconn.Packet
	runState                       xnsyncutil.AtomicInt
	nextCollectEntitySyncInfosTime time.Time
	dispatcherStartFreezeAcks      []bool
	positionSyncInterval           time.Duration
	ticker                         <-chan time.Time
	onlineGames                    common.Uint16Set
	isDeploymentReady              bool
}

func newGameService(gameid uint16) *GameService {
	//cfg := config.GetGame(gameid)
	return &GameService{
		id: gameid,
		//registeredServices: map[string]common.EntityIDSet{},
		packetQueue: make(chan *pktconn.Packet, consts.GAME_SERVICE_PACKET_QUEUE_SIZE),
		ticker:      time.Tick(consts.GAME_SERVICE_TICK_INTERVAL),
		onlineGames: common.Uint16Set{},
		//terminated:         xnsyncutil.NewOneTimeCond(),
		//dumpNotify:         xnsyncutil.NewOneTimeCond(),
		//dumpFinishedNotify: xnsyncutil.NewOneTimeCond(),
		//collectEntitySyncInfosRequest: make(chan struct{}),
		//collectEntitySycnInfosReply:   make(chan interface{}),
	}
}

func (gs *GameService) run() {
	gs.runState.Store(rsRunning)
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
		case _pkt := <-gs.packetQueue:
			pkt := (*netutil.Packet)(_pkt)
			msgtype := proto.MsgType(pkt.ReadUint16())

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
				eid := pkt.ReadEntityID()
				gid := pkt.ReadUint16()
				gs.HandleNotifyClientConnected(clientid, eid, gid)
			case proto.MT_NOTIFY_CLIENT_DISCONNECTED:
				eid := pkt.ReadEntityID()
				clientid := pkt.ReadClientID()
				gs.HandleNotifyClientDisconnected(eid, clientid)
			case proto.MT_LOAD_ENTITY_SOMEWHERE:
				_ = pkt.ReadUint16()
				eid := pkt.ReadEntityID()
				typeName := pkt.ReadVarStr()
				gs.HandleLoadEntitySomewhere(typeName, eid)
			case proto.MT_CREATE_ENTITY_SOMEWHERE:
				_ = pkt.ReadUint16() // gameid
				entityid := pkt.ReadEntityID()
				typeName := pkt.ReadVarStr()
				var data map[string]interface{}
				pkt.ReadData(&data)
				gs.HandleCreateEntitySomewhere(entityid, typeName, data)
			case proto.MT_CALL_NIL_SPACES:
				_ = pkt.ReadUint16() // ignore except gameid
				method := pkt.ReadVarStr()
				args := pkt.ReadArgs()
				gs.HandleCallNilSpaces(method, args)
			case proto.MT_KVREG_REGISTER:
				gs.HandleKvregRegister(pkt)
			case proto.MT_NOTIFY_GATE_DISCONNECTED:
				gateid := pkt.ReadUint16()
				gs.HandleGateDisconnected(gateid)
			case proto.MT_START_FREEZE_GAME_ACK:
				dispid := pkt.ReadUint16()
				gs.HandleStartFreezeGameAck(dispid)
			case proto.MT_NOTIFY_GAME_CONNECTED:
				gs.handleNotifyGameConnected(pkt)
			case proto.MT_NOTIFY_GAME_DISCONNECTED:
				gs.handleNotifyGameDisconnected(pkt)
			case proto.MT_NOTIFY_DEPLOYMENT_READY:
				gs.handleNotifyDeploymentReady(pkt)
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

func (gs *GameService) String() string {
	return fmt.Sprintf("GameService<%d>", gs.id)
}

func (gs *GameService) HandleCreateEntitySomewhere(entityid common.EntityID, typeName string, data map[string]interface{}) {
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleCreateEntityAnywhere: %s, typeName=%s, data=%v", gs, entityid, typeName, data)
	}
	entity.OnCreateEntitySomewhere(entityid, typeName, data)
}

func (gs *GameService) HandleLoadEntitySomewhere(typeName string, entityID common.EntityID) {
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleLoadEntityAnywhere: typeName=%s, entityID=%s", gs, typeName, entityID)
	}
	entity.OnLoadEntitySomewhere(typeName, entityID)
}

func (gs *GameService) HandleKvregRegister(pkt *netutil.Packet) {
	// tell the entity that it is registered successfully
	srvid := pkt.ReadVarStr()
	srvinfo := pkt.ReadVarStr()
	force := pkt.ReadBool() // force is not useful here
	gwlog.Infof("%s kvreg register: %s => %s, force %v", gs, srvid, srvinfo, force)

	kvreg.WatchKvregRegister(srvid, srvinfo)
}

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
	if gs.onlineGames.Contains(gameid) {
		// should not happen
		gwlog.Errorf("%s: handle notify game connected: game%d is connected, but it was already connected", gs, gameid)
		return
	}

	gs.onlineGames.Add(gameid)
	gwlog.Infof("%s notify game connected: %d online games currently", gs, len(gs.onlineGames))
}

func (gs *GameService) handleNotifyGameDisconnected(pkt *netutil.Packet) {
	gameid := pkt.ReadUint16()

	if !gs.onlineGames.Contains(gameid) {
		// should not happen
		gwlog.Errorf("%s: handle notify game disconnected: game%d is disconnected, but it was not connected", gs, gameid)
		return
	}

	gs.onlineGames.Remove(gameid)
	gwlog.Infof("%s notify game disconnected: %d online games left", gs, len(gs.onlineGames))
}

func (gs *GameService) handleNotifyDeploymentReady(pkt *netutil.Packet) {
	gs.onDeploymentReady()
}

func (gs *GameService) handleSetGameIDAck(pkt *netutil.Packet) {
	dispid := pkt.ReadUint16() // dispatcher  that sent the SET_GAME_ID_ACK
	isDeploymentReady := pkt.ReadBool()

	gameNum := int(pkt.ReadUint16())
	gs.onlineGames = common.Uint16Set{} // clear online games first
	for i := 0; i < gameNum; i++ {
		gameid := pkt.ReadUint16()
		gs.onlineGames.Add(gameid)
	}

	rejectEntitiesNum := pkt.ReadUint32()
	rejectEntities := make([]common.EntityID, 0, rejectEntitiesNum)
	for i := uint32(0); i < rejectEntitiesNum; i++ {
		rejectEntities = append(rejectEntities, pkt.ReadEntityID())
	}
	// remove all rejected entities
	for _, eid := range rejectEntities {
		e := entity.GetEntity(eid)
		if e != nil {
			e.Destroy()
		}
	}

	kvregMap := pkt.ReadMapStringString()
	kvreg.ClearByDispatcher(dispid)
	for srvid, srvinfo := range kvregMap {
		kvreg.WatchKvregRegister(srvid, srvinfo)
	}

	gwlog.Infof("%s: set game ID ack received, deployment ready: %v, %d online games, reject entities: %d, kvreg map: %+v",
		gs, isDeploymentReady, len(gs.onlineGames), rejectEntitiesNum, kvregMap)
	if isDeploymentReady {
		// all games are connected
		gs.onDeploymentReady()
	}
}

func (gs *GameService) onDeploymentReady() {
	if gs.isDeploymentReady {
		// should never happen, because dispatcher never send deployment ready to a game more than once
		return
	}

	gs.isDeploymentReady = true
	gwvar.IsDeploymentReady.Set(true)
	gwlog.Infof("DEPLOYMENT IS READY!")
	entity.OnGameReady()
	service.OnDeploymentReady()
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

func (gs *GameService) HandleNotifyClientConnected(clientid common.ClientID, bootEid common.EntityID, gateid uint16) {
	client := entity.MakeGameClient(clientid, gateid)
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleNotifyClientConnected: %s", gs, client)
	}

	// create a boot entity for the new client and set the client as the OWN CLIENT of the entity
	e := entity.CreateEntityLocallyWithID(gs.config.BootEntity, nil, bootEid)
	e.SetClient(client)
}

func (gs *GameService) HandleCallNilSpaces(method string, args [][]byte) {
	gwlog.Infof("%s.HandleCallNilSpaces: method=%s, argcount=%d", gs, method, len(args))
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.HandleCallNilSpaces: method=%s, argcount=%d", gs, method, len(args))
	}

	entity.OnCallNilSpaces(method, args)
}

func (gs *GameService) HandleNotifyClientDisconnected(ownerID common.EntityID, clientid common.ClientID) {
	if consts.DEBUG_CLIENTS {
		gwlog.Debugf("%s.handleNotifyClientDisconnected: %s.%s", gs, ownerID, clientid)
	}
	// find the owner of the client, and notify lose client
	entity.OnClientDisconnected(ownerID, clientid)
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
	data := pkt.ReadVarBytes()
	entity.OnRealMigrate(eid, data)
}

func (gs *GameService) terminate() {
	gs.runState.Store(rsTerminating)
}

func (gs *GameService) startFreeze() {
	dispatcherNum := len(config.GetDispatcherIDs())
	gs.dispatcherStartFreezeAcks = make([]bool, dispatcherNum)
	dispatchercluster.SendStartFreezeGame()
}

// GetOnlineGames returns all online game IDs
func GetOnlineGames() common.Uint16Set {
	return gameService.onlineGames
}
