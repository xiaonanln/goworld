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
)

type GameService struct {
	id                 int
	gameDelegate       IGameDelegate
	registeredServices map[string]entity.EntityIDSet
}

func newGameService(gameid int, delegate IGameDelegate) *GameService {
	return &GameService{
		id:                 gameid,
		gameDelegate:       delegate,
		registeredServices: map[string]entity.EntityIDSet{},
	}
}

func (gs *GameService) run() {
	cfg := config.GetGame(gameid)
	fmt.Fprintf(os.Stderr, "Read game %d config: \n%s\n", gameid, config.DumpPretty(cfg))

	dispatcher_client.Initialize(gs)

	timer.AddCallback(0, func() {
		gs.gameDelegate.OnReady()
	})

	tickCounter := 0
	for {
		timer.Tick()
		tickCounter += 1
		os.Stderr.Write([]byte{'.'})
		if tickCounter%100 == 0 {
			os.Stderr.Write([]byte{'\n'})
		}

		time.Sleep(time.Millisecond * 100)
	}
}

func (gs *GameService) String() string {
	return fmt.Sprintf("GameService<%d>", gs.id)
}

func (gs *GameService) OnDispatcherClientConnect() {
	gwlog.Debug("%s.OnDispatcherClientConnect ...", gs)
	dispatcher_client.GetDispatcherClientForSend().SendSetGameID(gs.id)
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
