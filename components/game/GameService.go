package game

import (
	"fmt"
	"os"

	"time"

	"github.com/xiaonanln/go-xnsyncutil/xnsyncutil"
	timer "github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/netutil"
	"github.com/xiaonanln/goworld/post"
	"github.com/xiaonanln/goworld/proto"
)

type packetQueueItem struct { // packet queue from dispatcher client
	msgtype proto.MsgType_t
	packet  *netutil.Packet
}

type GameService struct {
	config       *config.GameConfig
	id           uint16
	gameDelegate IServerDelegate
	//registeredServices map[string]entity.EntityIDSet

	packetQueue           chan packetQueueItem
	isAllServersConnected bool
	terminating           xnsyncutil.AtomicBool
	terminated            *xnsyncutil.OneTimeCond
}

func newGameService(gameid uint16, delegate IServerDelegate) *GameService {
	return &GameService{
		id:           gameid,
		gameDelegate: delegate,
		//registeredServices: map[string]entity.EntityIDSet{},
		packetQueue: make(chan packetQueueItem, consts.GAME_SERVICE_PACKET_QUEUE_SIZE),
		terminated:  xnsyncutil.NewOneTimeCond(),
	}
}

func (gs *GameService) run() {
	netutil.ServeForever(gs.serveRoutine)
}

func (gs *GameService) serveRoutine() {
	cfg := config.GetServer(gameid)
	gs.config = cfg
	gwlog.Info("Read game %d config: \n%s\n", gameid, config.DumpPretty(cfg))

	ticker := time.Tick(consts.SERVER_TICK_INTERVAL)
	// here begins the main loop of Server
	for {
		select {
		case item := <-gs.packetQueue:
			msgtype, pkt := item.msgtype, item.packet
			if msgtype == proto.MT_CALL_ENTITY_METHOD_FROM_CLIENT {
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
				sid := pkt.ReadUint16()
				gs.HandleNotifyClientConnected(clientid, sid)
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
			} else if msgtype == proto.MT_NOTIFY_ALL_SERVERS_CONNECTED {
				gs.HandleNotifyAllServersConnected()
			} else {
				gwlog.TraceError("unknown msgtype: %v", msgtype)
				if consts.DEBUG_MODE {
					os.Exit(2)
				}
			}

			pkt.Release()
		case <-ticker:
			if gs.terminating.Load() {
				// game is terminating, run the terminating process
				gs.doTerminate()
			}

			timer.Tick()
		}

		// after handling packets or firing timers, check the posted functions
		post.Tick()
	}
}

func (gs *GameService) doTerminate() {
	// destroy all entities
	entity.OnServerTerminating()
	gwlog.Info("All entities saved & destroyed, game service terminated.")
	gs.terminated.Signal() // signal terminated condition
	for {                  // enter the endless loop, not serving anything anymore
		time.Sleep(time.Second)
	}
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

func (gs *GameService) HandleNotifyAllServersConnected() {
	// all games are connected
	gs.gameDelegate.OnServerReady()
}

func (gs *GameService) HandleCallEntityMethod(entityID common.EntityID, method string, args [][]byte, clientid common.ClientID) {
	if consts.DEBUG_PACKETS {
		gwlog.Debug("%s.HandleCallEntityMethod: %s.%s(%v)", gs, entityID, method, args)
	}
	entity.OnCall(entityID, method, args, clientid)
}

func (gs *GameService) HandleNotifyClientConnected(clientid common.ClientID, sid uint16) {
	client := entity.MakeGameClient(clientid, sid)
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
	_ = pkt.ReadUint16() // targetServer is not userful

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
	if consts.DEBUG_PACKETS {
		gwlog.Debug("%s.HandleRealMigrate: entity %s migrating to space %s, typeName=%s, migrateData=%v, client=%s@%d", gs, eid, spaceID, typeName, migrateData, clientid, clientsrv)
	}

	entity.OnRealMigrate(eid, spaceID, x, y, z, typeName, migrateData, clientid, clientsrv)
}

func (gs *GameService) terminate() {
	gs.terminating.Store(true)
}
