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
	config             *config.ServerConfig
	id                 int
	serverDelegate     IServerDelegate
	registeredServices map[string]entity.EntityIDSet

	packetQueue chan packetQueueItem
}

func newGameService(serverid int, delegate IServerDelegate) *GameService {
	return &GameService{
		id:                 serverid,
		serverDelegate:     delegate,
		registeredServices: map[string]entity.EntityIDSet{},
		packetQueue:        make(chan packetQueueItem, consts.DISPATCHER_CLIENT_PACKET_QUEUE_SIZE),
	}
}

func (gs *GameService) run() {
	cfg := config.GetServer(serverid)
	gs.config = cfg
	fmt.Fprintf(os.Stderr, "Read server %d config: \n%s\n", serverid, config.DumpPretty(cfg))

	// initializing storage
	storage.Initialize()

	ticker := time.Tick(consts.SERVER_TICK_INTERVAL)
	timer.AddCallback(0, func() {
		gs.serverDelegate.OnReady()
	})

	// here begins the main loop of Server
	tickCount := 0
	for {
		select {
		case item := <-gs.packetQueue:
			msgtype, pkt := item.msgtype, item.pkt
			if msgtype == proto.MT_CALL_ENTITY_METHOD {
				eid := pkt.ReadEntityID()
				method := pkt.ReadVarStr()
				var args []interface{}
				pkt.ReadMessage(&args)
				gs.HandleCallEntityMethod(eid, method, args)
			} else if msgtype == proto.MT_NOTIFY_CLIENT_CONNECTED {
				clientid := pkt.ReadClientID()
				sid := pkt.ReadUint16()
				gs.HandleNotifyClientConnected(clientid, sid)
			} else if msgtype == proto.MT_LOAD_ENTITY_ANYWHERE {
				typeName := pkt.ReadVarStr()
				eid := pkt.ReadEntityID()
				gs.HandleLoadEntityAnywhere(typeName, eid)
			} else if msgtype == proto.MT_CREATE_ENTITY_ANYWHERE {
				typeName := pkt.ReadVarStr()
				gs.HandleCreateEntityAnywhere(typeName)
			} else if msgtype == proto.MT_DECLARE_SERVICE {
				eid := pkt.ReadEntityID()
				serviceName := pkt.ReadVarStr()
				gs.HandleDeclareService(eid, serviceName)
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

func (gs *GameService) HandleCreateEntityAnywhere(typeName string) {
	gwlog.Debug("%s.HandleCreateEntityAnywhere: typeName=%s", gs, typeName)
	entity.CreateEntityLocally(typeName, nil)
}

func (gs *GameService) HandleLoadEntityAnywhere(typeName string, entityID common.EntityID) {
	gwlog.Debug("%s.HandleLoadEntityAnywhere: typeName=%s, entityID=%s", gs, typeName, entityID)
	entity.LoadEntityLocally(typeName, entityID)
}

func (gs *GameService) HandleDeclareService(entityID common.EntityID, serviceName string) {
	// tell the entity that it is registered successfully
	gwlog.Debug("%s.HandleDeclareService: %s declares %s", gs, entityID, serviceName)
	eids, ok := gs.registeredServices[serviceName]
	if !ok {
		eids = entity.EntityIDSet{}
		gs.registeredServices[serviceName] = eids
	}
	eids.Add(entityID)
}

func (gs *GameService) HandleCallEntityMethod(entityID common.EntityID, method string, args []interface{}) {
	gwlog.Debug("%s.HandleCallEntityMethod: %s.%s()", gs, entityID, method)
	entity.OnCall(entityID, method, args)
}

func (gs *GameService) HandleNotifyClientConnected(clientid common.ClientID, sid uint16) {
	client := entity.MakeGameClient(clientid, sid)
	gwlog.Debug("%s.HandleNotifyClientConnected: %s", gs, client)

	// create a boot entity for the new client and set the client as the OWN CLIENT of the entity
	entity.CreateEntityLocally(gs.config.BootEntity, client)
}
