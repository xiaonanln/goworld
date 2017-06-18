package server

import (
	"fmt"
	"os"

	"time"

	timer "github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/netutil"
	"github.com/xiaonanln/goworld/proto"
	"github.com/xiaonanln/goworld/storage"
)

type packetQueueItem struct { // packet queue from dispatcher client
	msgtype proto.MsgType_t
	pkt     *netutil.Packet
}

type GameService struct {
	config         *config.ServerConfig
	id             uint16
	serverDelegate IServerDelegate
	//registeredServices map[string]entity.EntityIDSet

	packetQueue           chan packetQueueItem
	isAllServersConnected bool
}

func newGameService(serverid uint16, delegate IServerDelegate) *GameService {
	return &GameService{
		id:             serverid,
		serverDelegate: delegate,
		//registeredServices: map[string]entity.EntityIDSet{},
		packetQueue: make(chan packetQueueItem, consts.DISPATCHER_CLIENT_PACKET_QUEUE_SIZE),
	}
}

func (gs *GameService) run() {
	cfg := config.GetServer(serverid)
	gs.config = cfg
	fmt.Fprintf(os.Stderr, "Read server %d config: \n%s\n", serverid, config.DumpPretty(cfg))

	// initializing storage
	storage.Initialize()

	ticker := time.Tick(consts.SERVER_TICK_INTERVAL)
	// here begins the main loop of Server
	tickCount := 0
	for {
		select {
		case item := <-gs.packetQueue:
			msgtype, pkt := item.msgtype, item.pkt
			if msgtype == proto.MT_CALL_ENTITY_METHOD_FROM_CLIENT {
				eid := pkt.ReadEntityID()
				method := pkt.ReadVarStr()
				var args []interface{}
				pkt.ReadData(&args)
				clientid := pkt.ReadClientID()
				gs.HandleCallEntityMethod(eid, method, args, clientid)
			} else if msgtype == proto.MT_CALL_ENTITY_METHOD {
				eid := pkt.ReadEntityID()
				method := pkt.ReadVarStr()
				var args []interface{}
				pkt.ReadData(&args)
				gs.HandleCallEntityMethod(eid, method, args, "")
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
			}

			pkt.Release()
		case <-ticker:
			timer.Tick()
			tickCount += 1
			if tickCount%100 == 0 {
				os.Stderr.Write([]byte{'|'})
			}
		}
	}
}

func (gs *GameService) String() string {
	return fmt.Sprintf("GameService<%d>", gs.id)
}

func (gs *GameService) HandleCreateEntityAnywhere(typeName string, data map[string]interface{}) {
	gwlog.Debug("%s.HandleCreateEntityAnywhere: typeName=%s, data=%v", gs, typeName, data)
	entity.CreateEntityLocally(typeName, data, nil)
}

func (gs *GameService) HandleLoadEntityAnywhere(typeName string, entityID common.EntityID) {
	gwlog.Debug("%s.HandleLoadEntityAnywhere: typeName=%s, entityID=%s", gs, typeName, entityID)
	entity.LoadEntityLocally(typeName, entityID)
}

func (gs *GameService) HandleDeclareService(entityID common.EntityID, serviceName string) {
	// tell the entity that it is registered successfully
	gwlog.Debug("%s.HandleDeclareService: %s declares %s", gs, entityID, serviceName)
	entity.OnDeclareService(serviceName, entityID)
}

func (gs *GameService) HandleUndeclareService(entityID common.EntityID, serviceName string) {
	// tell the entity that it is registered successfully
	gwlog.Debug("%s.HandleUndeclareService: %s undeclares %s", gs, entityID, serviceName)
	entity.OnUndeclareService(serviceName, entityID)
}

func (gs *GameService) HandleNotifyAllServersConnected() {
	// all servers are connected
	gs.serverDelegate.OnServerReady()
}

func (gs *GameService) HandleCallEntityMethod(entityID common.EntityID, method string, args []interface{}, clientid common.ClientID) {
	gwlog.Debug("%s.HandleCallEntityMethod: %s.%s()", gs, entityID, method)
	entity.OnCall(entityID, method, args, clientid)
}

func (gs *GameService) HandleNotifyClientConnected(clientid common.ClientID, sid uint16) {
	client := entity.MakeGameClient(clientid, sid)
	gwlog.Debug("%s.HandleNotifyClientConnected: %s", gs, client)

	// create a boot entity for the new client and set the client as the OWN CLIENT of the entity
	entity.CreateEntityLocally(gs.config.BootEntity, nil, client)
}

func (gs *GameService) HandleNotifyClientDisconnected(clientid common.ClientID) {
	gwlog.Debug("%s.HandleNotifyClientDisconnected: %s", gs, clientid)
	// find the owner of the client, and notify lose client
	entity.OnClientDisconnected(clientid)
}
