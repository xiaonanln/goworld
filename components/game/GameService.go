package game

import (
	"fmt"
	"os"

	"time"

	"io/ioutil"

	"github.com/xiaonanln/go-xnsyncutil/xnsyncutil"
	timer "github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/components/dispatcher/dispatcher_client"
	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/kvdb"
	"github.com/xiaonanln/goworld/netutil"
	"github.com/xiaonanln/goworld/post"
	"github.com/xiaonanln/goworld/proto"
)

const (
	rsNotRunning = iota
	rsRunning
	rsTerminating
	rsTerminated
	rsFreezing
	rsFreezed
)

type packetQueueItem struct { // packet queue from dispatcher client
	msgtype proto.MsgType_t
	packet  *netutil.Packet
}

type GameService struct {
	config       *config.GameConfig
	id           uint16
	gameDelegate IGameDelegate
	//registeredServices map[string]entity.EntityIDSet

	packetQueue         chan packetQueueItem
	isAllGamesConnected bool
	runState            xnsyncutil.AtomicInt
}

func newGameService(gameid uint16, delegate IGameDelegate) *GameService {
	return &GameService{
		id:           gameid,
		gameDelegate: delegate,
		//registeredServices: map[string]entity.EntityIDSet{},
		packetQueue: make(chan packetQueueItem, consts.GAME_SERVICE_PACKET_QUEUE_SIZE),
		//terminated:         xnsyncutil.NewOneTimeCond(),
		//dumpNotify:         xnsyncutil.NewOneTimeCond(),
		//dumpFinishedNotify: xnsyncutil.NewOneTimeCond(),
	}
}

func (gs *GameService) run(restore bool) {
	gs.runState.Store(rsRunning)

	if !restore {
		entity.CreateSpaceLocally(0) // create to be the nil space
	} else {
		// restoring from freezed states
		err := gs.doRestore()
		if err != nil {
			gwlog.Fatal("Restore from freezed states failed: %+v", err)
		}
	}

	netutil.ServeForever(gs.serveRoutine)
}

func (gs *GameService) serveRoutine() {
	cfg := config.GetGame(gameid)
	gs.config = cfg
	gwlog.Info("Read game %d config: \n%s\n", gameid, config.DumpPretty(cfg))

	ticker := time.Tick(consts.GAME_SERVICE_TICK_INTERVAL)
	// here begins the main loop of Game
	for {
		select {
		case item := <-gs.packetQueue:
			msgtype, pkt := item.msgtype, item.packet
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
				typeName := pkt.ReadVarStr()
				var data map[string]interface{}
				pkt.ReadData(&data)
				gs.HandleCreateEntityAnywhere(typeName, data)
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
				gs.HandleStartFreezeGameAck()
			} else {
				gwlog.TraceError("unknown msgtype: %v", msgtype)
				if consts.DEBUG_MODE {
					os.Exit(2)
				}
			}

			pkt.Release()
		case <-ticker:
			runState := gs.runState.Load()
			if runState == rsTerminating {
				// game is terminating, run the terminating process
				gs.doTerminate()
			} else if runState == rsFreezing {
				//game is freezing, run freeze process
				gs.doFreeze()
			}

			timer.Tick()
		}

		// after handling packets or firing timers, check the posted functions
		post.Tick()
	}
}

func (gs *GameService) waitPostsComplete() {
	post.Tick() // just tick is Ok, tick will consume all posts
}

func (gs *GameService) doTerminate() {
	// wait for all posts to complete
	gs.waitPostsComplete()

	// destroy all entities
	entity.OnGameTerminating()
	gwlog.Info("All entities saved & destroyed, game service terminated.")
	gs.runState.Store(rsTerminated)

	for {
		time.Sleep(time.Second)
	}
}

var freezePacker = netutil.JSONMsgPacker{}

func (gs *GameService) doFreeze() {
	// wait for all posts to complete

	kvdb.Close()
	kvdb.WaitTerminated()
	gs.waitPostsComplete()

	// save all entities
	entity.SaveAllEntities()
	// destroy all entities
	freeze := func() error {
		freezeEntity, err := entity.Freeze(gameid)
		if err != nil {
			return err
		}
		freezeData, err := freezePacker.PackMsg(freezeEntity, nil)
		if err != nil {
			return err
		}
		freezeFilename := freezeFilename(gameid)
		err = ioutil.WriteFile(freezeFilename, freezeData, 0644)
		if err != nil {
			return err
		}

		return nil
	}

	err := freeze()
	if err != nil {
		gwlog.Error("Game freeze failed: %s", err)
		kvdb.Initialize() // restore kvdb module
		gs.runState.Store(rsRunning)
		return
	}

	gs.runState.Store(rsFreezed)
	gwlog.Info("All entities saved & freezed, game service terminated.")
	for {
		time.Sleep(time.Second)
	}
}

func freezeFilename(gameid uint16) string {
	return fmt.Sprintf("game%d_freezed.dat", gameid)
}

func (gs *GameService) doRestore() error {
	freezeFilename := freezeFilename(gameid)
	data, err := ioutil.ReadFile(freezeFilename)
	if err != nil {
		return err
	}

	var freezeEntity entity.FreezeData
	freezePacker.UnpackMsg(data, &freezeEntity)

	return entity.RestoreFreezedEntities(&freezeEntity)
}

func (gs *GameService) String() string {
	return fmt.Sprintf("GameService<%d>", gs.id)
}

func (gs *GameService) HandleCreateEntityAnywhere(typeName string, data map[string]interface{}) {
	if consts.DEBUG_PACKETS {
		gwlog.Debug("%s.HandleCreateEntityAnywhere: typeName=%s, data=%v", gs, typeName, data)
	}
	entity.CreateEntityLocally(typeName, data, nil)
}

func (gs *GameService) HandleLoadEntityAnywhere(typeName string, entityID common.EntityID) {
	if consts.DEBUG_PACKETS {
		gwlog.Debug("%s.HandleLoadEntityAnywhere: typeName=%s, entityID=%s", gs, typeName, entityID)
	}
	entity.LoadEntityLocally(typeName, entityID)
}

func (gs *GameService) HandleDeclareService(entityID common.EntityID, serviceName string) {
	// tell the entity that it is registered successfully
	if consts.DEBUG_PACKETS {
		gwlog.Debug("%s.HandleDeclareService: %s declares %s", gs, entityID, serviceName)
	}
	entity.OnDeclareService(serviceName, entityID)
}

func (gs *GameService) HandleUndeclareService(entityID common.EntityID, serviceName string) {
	// tell the entity that it is registered successfully
	if consts.DEBUG_PACKETS {
		gwlog.Debug("%s.HandleUndeclareService: %s undeclares %s", gs, entityID, serviceName)
	}
	entity.OnUndeclareService(serviceName, entityID)
}

func (gs *GameService) HandleNotifyAllGamesConnected() {
	// all games are connected
	gwlog.Info("All games connected.")
	gs.gameDelegate.OnGameReady()
}

func (gs *GameService) HandleGateDisconnected(gateid uint16) {
	entity.OnGateDisconnected(gateid)
}

func (gs *GameService) HandleStartFreezeGameAck() {
	gwlog.Info("Start freeze game ACK received, start freezing ...")
	gs.runState.Store(rsFreezing)
}

func (gs *GameService) HandleSyncPositionYawFromClient(pkt *netutil.Packet) {
	//gwlog.Info("HandleSyncPositionYawFromClient: payload %d", len(pkt.UnreadPayload()))
	payload := pkt.UnreadPayload()
	payloadLen := len(payload)
	for i := 0; i < payloadLen; i += proto.SYNC_INFO_SIZE_PER_ENTITY + common.ENTITYID_LENGTH {
		eid := common.EntityID(payload[i : i+common.ENTITYID_LENGTH])
		x := netutil.PACKET_ENDIAN.Uint32(payload[i+common.ENTITYID_LENGTH : i+common.ENTITYID_LENGTH+4])
		y := netutil.PACKET_ENDIAN.Uint32(payload[i+common.ENTITYID_LENGTH+4 : i+common.ENTITYID_LENGTH+8])
		z := netutil.PACKET_ENDIAN.Uint32(payload[i+common.ENTITYID_LENGTH+8 : i+common.ENTITYID_LENGTH+12])
		yaw := netutil.PACKET_ENDIAN.Uint32(payload[i+common.ENTITYID_LENGTH+12 : i+common.ENTITYID_LENGTH+16])
		entity.OnSyncPositionYawFromClient(eid, entity.Coord(x), entity.Coord(y), entity.Coord(z), entity.Yaw(yaw))
	}
	//eid := pkt.ReadEntityID()
	//x := pkt.ReadUint32()
	//y := pkt.ReadUint32()
	//z := pkt.ReadUint32()
	//yaw := pkt.ReadUint32()
	//clientid := pkt.ReadClientID()
	//entity.OnSyncPositionYawFromClient(eid, entity.Coord(x), entity.Coord(y), entity.Coord(z), entity.Yaw(yaw), clientid)
}

func (gs *GameService) HandleCallEntityMethod(entityID common.EntityID, method string, args [][]byte, clientid common.ClientID) {
	if consts.DEBUG_PACKETS {
		gwlog.Debug("%s.HandleCallEntityMethod: %s.%s(%v)", gs, entityID, method, args)
	}
	entity.OnCall(entityID, method, args, clientid)
}

func (gs *GameService) HandleNotifyClientConnected(clientid common.ClientID, gid uint16) {
	client := entity.MakeGameClient(clientid, gid)
	if consts.DEBUG_PACKETS {
		gwlog.Debug("%s.HandleNotifyClientConnected: %s", gs, client)
	}

	// create a boot entity for the new client and set the client as the OWN CLIENT of the entity
	entity.CreateEntityLocally(gs.config.BootEntity, nil, client)
}

func (gs *GameService) HandleNotifyClientDisconnected(clientid common.ClientID) {
	if consts.DEBUG_CLIENTS {
		gwlog.Debug("%s.HandleNotifyClientDisconnected: %s", gs, clientid)
	}
	// find the owner of the client, and notify lose client
	entity.OnClientDisconnected(clientid)
}

func (gs *GameService) HandleMigrateRequestAck(pkt *netutil.Packet) {
	eid := pkt.ReadEntityID()
	spaceid := pkt.ReadEntityID()
	spaceLoc := pkt.ReadUint16()

	if consts.DEBUG_PACKETS {
		gwlog.Debug("Entity %s is migrating to space %s at game %d", eid, spaceid, spaceLoc)
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
		gwlog.Debug("%s.HandleRealMigrate: entity %s migrating to space %s, typeName=%s, migrateData=%v, timerData=%v, client=%s@%d", gs, eid, spaceID, typeName, migrateData, timerData, clientid, clientsrv)
	}

	entity.OnRealMigrate(eid, spaceID, x, y, z, typeName, migrateData, timerData, clientid, clientsrv)
}

func (gs *GameService) terminate() {
	gs.runState.Store(rsTerminating)
}

func (gs *GameService) freeze() {
	dispatcher_client.GetDispatcherClientForSend().SendStartFreezeGame(gameid)
}
