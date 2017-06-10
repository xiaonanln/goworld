package game

import (
	"fmt"
	"os"

	"time"

	timer "github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/components/dispatcher/dispatcher_client"
	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/netutil"
	"github.com/xiaonanln/goworld/proto"
)

type packetQueueItem struct { // packet queue from dispatcher client
	msgtype proto.MsgType_t
	pkt     *netutil.Packet
}

type GameService struct {
	id                 int
	gameDelegate       IGameDelegate
	registeredServices map[string]entity.EntityIDSet

	packetQueue chan packetQueueItem
}

func newGameService(gameid int, delegate IGameDelegate) *GameService {
	return &GameService{
		id:                 gameid,
		gameDelegate:       delegate,
		registeredServices: map[string]entity.EntityIDSet{},
		packetQueue:        make(chan packetQueueItem, DISPATCHER_CLIENT_PACKET_QUEUE_SIZE),
	}
}

func (gs *GameService) run() {
	cfg := config.GetGame(gameid)
	fmt.Fprintf(os.Stderr, "Read game %d config: \n%s\n", gameid, config.DumpPretty(cfg))

	dispatcher_client.Initialize(gs)
	ticker := time.Tick(TICK_INTERVAL)
	timer.AddCallback(0, func() {
		gs.gameDelegate.OnReady()
	})

	// here begins the main loop of Game
	tickCount := 0
	for {
		select {
		case item := <-gs.packetQueue:
			gwlog.Debug("Game %s recv packet: %v %v", gs, item.msgtype, item.pkt.Payload())
			msgtype, pkt := item.msgtype, item.pkt
			if msgtype == proto.MT_CALL_ENTITY_METHOD {
				eid := pkt.ReadEntityID()
				method := pkt.ReadVarStr()
				argsData := pkt.ReadVarBytes()
				var args []interface{}
				proto.ARGS_PACKER.UnpackMsg(argsData, &args)
				gs.HandleCallEntityMethod(eid, method, args)
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

func (gs *GameService) OnDispatcherClientConnect() {
	gwlog.Debug("%s.OnDispatcherClientConnect ...", gs)
	dispatcher_client.GetDispatcherClientForSend().SendSetGameID(gs.id)
}

func (gs *GameService) HandleDispatcherClientPacket(msgtype proto.MsgType_t, pkt *netutil.Packet) {
	gs.packetQueue <- packetQueueItem{ // may block the dispatcher client routine
		msgtype: msgtype,
		pkt:     pkt,
	}
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
