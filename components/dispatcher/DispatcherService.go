package main

import (
	"fmt"

	"net"

	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/netutil"
)

type DispatcherService struct {
	config     *config.DispatcherConfig
	clients    map[int]*DispatcherClientProxy
	entityLocs map[common.EntityID]int
}

func newDispatcherService(cfg *config.DispatcherConfig) *DispatcherService {
	return &DispatcherService{
		config:     cfg,
		entityLocs: map[common.EntityID]int{},
	}
}

func (ds *DispatcherService) String() string {
	return fmt.Sprintf("DispatcherService<C%d|E%d>", len(ds.clients), len(ds.entityLocs))
}

func (ds *DispatcherService) run() {
	host := fmt.Sprintf("%s:%d", ds.config.Ip, ds.config.Port)
	netutil.ServeTCPForever(host, ds)
}

func (ds *DispatcherService) ServeTCPConnection(conn net.Conn) {
	client := newDispatcherClientProxy(ds, conn)
	client.serve()
}

func (ds *DispatcherService) HandleSetGameID(dcp *DispatcherClientProxy, gameid int) {
	gwlog.Debug("%s.HandleSetGameID: dcp=%s, gameid=%d", ds, dcp, gameid)
	return
}

// Entity is create on the target game
func (ds *DispatcherService) HandleNotifyCreateEntity(dcp *DispatcherClientProxy, entityID common.EntityID) {
	gwlog.Debug("%s.HandleNotifyCreateEntity: dcp=%s, entityID=%s", ds, dcp, entityID)
	ds.entityLocs[entityID] = dcp.gameid
}
