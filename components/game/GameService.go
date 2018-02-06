package game

import (
	"fmt"
	"os"

	"time"

	"io/ioutil"

	"github.com/xiaonanln/go-xnsyncutil/xnsyncutil"
	"github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld/engine/async"
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
)

const (
	rsNotRunning = iota
	rsRunning
	rsTerminating
	rsTerminated
	rsFreezing
	rsFreezed
)

type _GameService struct {
	config       *config.GameConfig
	id           uint16
	gameDelegate IGameDelegate
	//registeredServices map[string]entity.EntityIDSet

	packetQueue                    chan proto.Message
	isAllGamesConnected            bool
	runState                       xnsyncutil.AtomicInt
	nextCollectEntitySyncInfosTime time.Time
	dispatcherStartFreezeAcks      []bool
	positionSyncInterval           time.Duration
	//collectEntitySyncInfosRequest chan struct{}
	//collectEntitySycnInfosReply   chan interface{}
}

func newGameService(gameid uint16, delegate IGameDelegate) *_GameService {
	return &_GameService{
		id:           gameid,
		gameDelegate: delegate,
		//registeredServices: map[string]entity.EntityIDSet{},
		packetQueue: make(chan proto.Message, consts.GAME_SERVICE_PACKET_QUEUE_SIZE),
		//terminated:         xnsyncutil.NewOneTimeCond(),
		//dumpNotify:         xnsyncutil.NewOneTimeCond(),
		//dumpFinishedNotify: xnsyncutil.NewOneTimeCond(),
		//collectEntitySyncInfosRequest: make(chan struct{}),
		//collectEntitySycnInfosReply:   make(chan interface{}),
	}
}

func (gs *_GameService) run(restore bool) {
	gs.runState.Store(rsRunning)

	if !restore {
		entity.CreateSpaceLocally(0) // create to be the nil space
	} else {
		// restoring from freezed states
		err := gs.doRestore()
		if err != nil {
			gwlog.Fatalf("Restore from freezed states failed: %+v", err)
		}
	}

	fmt.Fprintf(gwlog.GetOutput(), "%s\n", consts.GAME_STARTED_TAG)
	gwutils.RepeatUntilPanicless(gs.serveRoutine)
}

func (gs *_GameService) serveRoutine() {
	cfg := config.GetGame(gameid)
	gs.config = cfg
	gs.positionSyncInterval = time.Millisecond * time.Duration(cfg.PositionSyncIntervalMS)
	if gs.positionSyncInterval < consts.GAME_SERVICE_TICK_INTERVAL {
		gwlog.Warnf("%s: entity position sync interval is too small: %s, so reset to %s", gs, gs.positionSyncInterval, consts.GAME_SERVICE_TICK_INTERVAL)
		gs.positionSyncInterval = consts.GAME_SERVICE_TICK_INTERVAL
	}

	gwlog.Infof("Read game %d config: \n%s\n", gameid, config.DumpPretty(cfg))

	ticker := time.Tick(consts.GAME_SERVICE_TICK_INTERVAL)
	// here begins the main loop of Game
	for {
		isTick := false
		select {
		case item := <-gs.packetQueue:
			msgtype, pkt := item.MsgType, item.Packet
			if msgtype == proto.MT_SYNC_POSITION_YAW_FROM_CLIENT {
				gs.HandleSyncPositionYawFromClient(pkt)
			} else if msgtype == proto.MT_CALL_ENTITY_METHOD_FROM_CLIENT {
				eid := pkt.ReadEntityID()
				method := pkt.ReadVarStr()
				args := pkt.ReadArgs()
				clientid := pkt.ReadClientID()
				gs.HandleCallEntityMethod(eid, method, args, clientid)
			} else if msgtype == proto.MT_CALL_ENTITY_METHOD {
				eid := pkt.ReadEntityID()
				method := pkt.ReadVarStr()
				args := pkt.ReadArgs()
				gs.HandleCallEntityMethod(eid, method, args, "")
			} else if msgtype == proto.MT_MIGRATE_REQUEST { // migrate request sent to dispatcher is sent back
				gs.HandleMigrateRequestAck(pkt)
			} else if msgtype == proto.MT_REAL_MIGRATE {
				gs.HandleRealMigrate(pkt)
			} else if msgtype == proto.MT_NOTIFY_CLIENT_CONNECTED {
				clientid := pkt.ReadClientID()
				gid := pkt.ReadUint16()
				gs.HandleNotifyClientConnected(clientid, gid)
			} else if msgtype == proto.MT_NOTIFY_CLIENT_DISCONNECTED {
				clientid := pkt.ReadClientID()
				gs.HandleNotifyClientDisconnected(clientid)
			} else if msgtype == proto.MT_LOAD_ENTITY_ANYWHERE {
				eid := pkt.ReadEntityID()
				typeName := pkt.ReadVarStr()
				gs.HandleLoadEntityAnywhere(typeName, eid)
			} else if msgtype == proto.MT_CREATE_ENTITY_ANYWHERE {
				entityid := pkt.ReadEntityID()
				typeName := pkt.ReadVarStr()
				var data map[string]interface{}
				pkt.ReadData(&data)
				gs.HandleCreateEntityAnywhere(entityid, typeName, data)
			} else if msgtype == proto.MT_DECLARE_SERVICE {
				eid := pkt.ReadEntityID()
				serviceName := pkt.ReadVarStr()
				gs.HandleDeclareService(eid, serviceName)
			} else if msgtype == proto.MT_UNDECLARE_SERVICE {
				eid := pkt.ReadEntityID()
				serviceName := pkt.ReadVarStr()
				gs.HandleUndeclareService(eid, serviceName)
			} else if msgtype == proto.MT_NOTIFY_ALL_GAMES_CONNECTED {
				gs.HandleNotifyAllGamesConnected()
			} else if msgtype == proto.MT_NOTIFY_GATE_DISCONNECTED {
				gateid := pkt.ReadUint16()
				gs.HandleGateDisconnected(gateid)
			} else if msgtype == proto.MT_START_FREEZE_GAME_ACK {
				dispid := pkt.ReadUint16()
				gs.HandleStartFreezeGameAck(dispid)
			} else {
				gwlog.TraceError("unknown msgtype: %v", msgtype)
				if consts.DEBUG_MODE {
					os.Exit(2)
				}
			}

			pkt.Release()
		case <-ticker:
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
			if !gs.nextCollectEntitySyncInfosTime.Before(now) {
				gs.nextCollectEntitySyncInfosTime = now.Add(gs.positionSyncInterval)
				entity.CollectEntitySyncInfos()
			}
		}
	}
}

func (gs *_GameService) waitPostsComplete() {
	gwlog.Infof("waiting for posts to complete ...")
	post.Tick() // just tick is Ok, tick will consume all posts
}

func (gs *_GameService) doTerminate() {
	// wait for all posts to complete
	gs.waitPostsComplete()
	// wait for all async to clear
	for async.WaitClear() { // wait for all async to stop
		gs.waitPostsComplete()
	}

	// destroy all entities
	entity.OnGameTerminating()
	gwlog.Infof("All entities saved & destroyed, game service terminated.")
	gs.runState.Store(rsTerminated)

	for {
		time.Sleep(time.Second)
	}
}

var freezePacker = netutil.MessagePackMsgPacker{}

func (gs *_GameService) doFreeze() {
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

func (gs *_GameService) doRestore() error {
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

func (gs *_GameService) String() string {
	return fmt.Sprintf("_GameService<%d>", gs.id)
}

func (gs *_GameService) HandleCreateEntityAnywhere(entityid common.EntityID, typeName string, data map[string]interface{}) {
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleCreateEntityAnywhere: %s, typeName=%s, data=%v", gs, entityid, typeName, data)
	}
	entity.OnCreateEntityAnywhere(entityid, typeName, data)
}

func (gs *_GameService) HandleLoadEntityAnywhere(typeName string, entityID common.EntityID) {
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleLoadEntityAnywhere: typeName=%s, entityID=%s", gs, typeName, entityID)
	}
	entity.LoadEntityLocally(typeName, entityID)
}

func (gs *_GameService) HandleDeclareService(entityID common.EntityID, serviceName string) {
	// tell the entity that it is registered successfully
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleDeclareService: %s declares %s", gs, entityID, serviceName)
	}
	entity.OnDeclareService(serviceName, entityID)
}

func (gs *_GameService) HandleUndeclareService(entityID common.EntityID, serviceName string) {
	// tell the entity that it is registered successfully
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.HandleUndeclareService: %s undeclares %s", gs, entityID, serviceName)
	}
	entity.OnUndeclareService(serviceName, entityID)
}

func (gs *_GameService) HandleNotifyAllGamesConnected() {
	// all games are connected
	gwlog.Infof("All games connected.")
	gs.gameDelegate.OnGameReady()
}

func (gs *_GameService) HandleGateDisconnected(gateid uint16) {
	entity.OnGateDisconnected(gateid)
}

func (gs *_GameService) HandleStartFreezeGameAck(dispid uint16) {
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

func (gs *_GameService) HandleSyncPositionYawFromClient(pkt *netutil.Packet) {
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

func (gs *_GameService) HandleCallEntityMethod(entityID common.EntityID, method string, args [][]byte, clientid common.ClientID) {
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleCallEntityMethod: %s.%s(%v)", gs, entityID, method, args)
	}
	entity.OnCall(entityID, method, args, clientid)
}

func (gs *_GameService) HandleNotifyClientConnected(clientid common.ClientID, gateid uint16) {
	client := entity.MakeGameClient(clientid, gateid)
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleNotifyClientConnected: %s", gs, client)
	}

	// create a boot entity for the new client and set the client as the OWN CLIENT of the entity
	entity.CreateEntityLocally(gs.config.BootEntity, nil, client)
}

func (gs *_GameService) HandleNotifyClientDisconnected(clientid common.ClientID) {
	if consts.DEBUG_CLIENTS {
		gwlog.Debugf("%s.handleNotifyClientDisconnected: %s", gs, clientid)
	}
	// find the owner of the client, and notify lose client
	entity.OnClientDisconnected(clientid)
}

func (gs *_GameService) HandleMigrateRequestAck(pkt *netutil.Packet) {
	eid := pkt.ReadEntityID()
	spaceid := pkt.ReadEntityID()
	spaceLoc := pkt.ReadUint16()

	if consts.DEBUG_PACKETS {
		gwlog.Debugf("Entity %s is migrating to space %s at game %d", eid, spaceid, spaceLoc)
	}

	entity.OnMigrateRequestAck(eid, spaceid, spaceLoc)
}

func (gs *_GameService) HandleRealMigrate(pkt *netutil.Packet) {
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

func (gs *_GameService) terminate() {
	gs.runState.Store(rsTerminating)
}

func (gs *_GameService) startFreeze() {
	dispatcherNum := len(config.GetDispatcherIDs())
	gs.dispatcherStartFreezeAcks = make([]bool, dispatcherNum)
	dispatchercluster.SendStartFreezeGame(gameid)
}
