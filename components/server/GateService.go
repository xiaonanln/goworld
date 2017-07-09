package server

import (
	"fmt"
	"time"

	"net"

	"sync"

	"os"

	"github.com/xiaonanln/goSyncQueue"
	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/components/dispatcher/dispatcher_client"
	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/netutil"
	"github.com/xiaonanln/goworld/opmon"
	"github.com/xiaonanln/goworld/proto"
)

type GateService struct {
	listenAddr        string
	clientProxies     map[common.ClientID]*ClientProxy
	clientProxiesLock sync.RWMutex
	packetQueue       sync_queue.SyncQueue

	filterTreesLock sync.Mutex
	filterTrees     map[string]*FilterTree
}

func newGateService() *GateService {
	return &GateService{
		//packetQueue: make(chan packetQueueItem, consts.DISPATCHER_CLIENT_PACKET_QUEUE_SIZE),
		clientProxies: map[common.ClientID]*ClientProxy{},
		packetQueue:   sync_queue.NewSyncQueue(),
		filterTrees:   map[string]*FilterTree{},
	}
}

func (gs *GateService) run() {
	cfg := config.GetServer(serverid)
	gwlog.Info("Compress connection: %v", cfg.CompressConnection)
	gs.listenAddr = fmt.Sprintf("%s:%d", cfg.Ip, cfg.Port)
	go netutil.ServeForever(gs.handlePacketRoutine)
	netutil.ServeTCPForever(gs.listenAddr, gs)
}

func (gs *GateService) String() string {
	return fmt.Sprintf("GateService<%s>", gs.listenAddr)
}

func (gs *GateService) ServeTCPConnection(conn net.Conn) {
	cfg := config.GetServer(serverid)
	cp := newClientProxy(conn, cfg)

	gs.clientProxiesLock.Lock()
	gs.clientProxies[cp.clientid] = cp
	gs.clientProxiesLock.Unlock()

	dispatcher_client.GetDispatcherClientForSend().SendNotifyClientConnected(cp.clientid)
	if consts.DEBUG_CLIENTS {
		gwlog.Debug("%s.ServeTCPConnection: client %s connected", gs, cp)
	}
	cp.serve()
}

func (gs *GateService) onClientProxyClose(cp *ClientProxy) {
	gs.clientProxiesLock.Lock()
	delete(gs.clientProxies, cp.clientid)
	gs.clientProxiesLock.Unlock()

	gs.filterTreesLock.Lock()
	for key, val := range cp.filterProps {
		ft := gs.filterTrees[key]
		if ft != nil {
			if consts.DEBUG_FILTER_PROP {
				gwlog.Debug("DROP CLIENT %s FILTER PROP: %s = %s", cp, key, val)
			}
			ft.Remove(cp.clientid, val)
		}
	}
	gs.filterTreesLock.Unlock()

	dispatcher_client.GetDispatcherClientForSend().SendNotifyClientDisconnected(cp.clientid)
	if consts.DEBUG_CLIENTS {
		gwlog.Debug("%s.onClientProxyClose: client %s disconnected", gs, cp)
	}
}

func (gs *GateService) HandleDispatcherClientPacket(msgtype proto.MsgType_t, packet *netutil.Packet) {
	if consts.DEBUG_PACKETS {
		gwlog.Debug("%s.HandleDispatcherClientPacket: msgtype=%v, packet(%d)=%v", gs, msgtype, packet.GetPayloadLen(), packet.Payload())
	}

	if msgtype >= proto.MT_REDIRECT_TO_GATEPROXY_MSG_TYPE_START && msgtype <= proto.MT_REDIRECT_TO_GATEPROXY_MSG_TYPE_STOP {
		_ = packet.ReadUint16() // sid
		clientid := packet.ReadClientID()

		gs.clientProxiesLock.RLock()
		clientproxy := gs.clientProxies[clientid]
		gs.clientProxiesLock.RUnlock()

		// message types that should be redirected to client proxy
		if clientproxy != nil {
			clientproxy.SendPacket(packet)
		} else {
			// client already disconnected, but the game service seems not knowing it, so tell it
			dispatcher_client.GetDispatcherClientForSend().SendNotifyClientDisconnected(clientid)
		}
	} else if msgtype == proto.MT_CALL_FILTERED_CLIENTS {
		gs.handleCallFilteredClientProxies(packet)
	} else if msgtype == proto.MT_SET_CLIENTPROXY_FILTER_PROP {
		// set filter property
		_ = packet.ReadUint16() // sid // TODO: bad code style here
		clientid := packet.ReadClientID()

		gs.clientProxiesLock.RLock()
		clientproxy := gs.clientProxies[clientid]
		gs.clientProxiesLock.RUnlock()

		if clientproxy != nil {
			gs.handleSetClientFilterProp(clientproxy, packet)
		}
	} else if msgtype == proto.MT_CLEAR_CLIENTPROXY_FILTER_PROPS {
		_ = packet.ReadUint16() // sid
		clientid := packet.ReadClientID()

		gs.clientProxiesLock.RLock()
		clientproxy := gs.clientProxies[clientid]
		gs.clientProxiesLock.RUnlock()

		if clientproxy != nil {
			gs.handleClearClientFilterProps(clientproxy, packet)
		}
	} else {
		gwlog.Panicf("%s: unknown msg type: %d", gs, msgtype)
		if consts.DEBUG_MODE {
			os.Exit(2)
		}
	}
}

func (gs *GateService) handleSetClientFilterProp(clientproxy *ClientProxy, packet *netutil.Packet) {
	gwlog.Debug("%s.handleSetClientFilterProp: clientproxy=%s", gs, clientproxy)
	key := packet.ReadVarStr()
	val := packet.ReadVarStr()
	clientid := clientproxy.clientid

	gs.filterTreesLock.Lock()
	ft, ok := gs.filterTrees[key]
	if !ok {
		ft = NewFilterTree()
		gs.filterTrees[key] = ft
	}

	oldVal, ok := clientproxy.filterProps[key]
	if ok {
		if consts.DEBUG_FILTER_PROP {
			gwlog.Debug("REMOVE CLIENT %s FILTER PROP: %s = %s", clientproxy, key, val)
		}
		ft.Remove(clientid, oldVal)
	}
	clientproxy.filterProps[key] = val
	ft.Insert(clientid, val)
	gs.filterTreesLock.Unlock()

	if consts.DEBUG_FILTER_PROP {
		gwlog.Debug("SET CLIENT %s FILTER PROP: %s = %s", clientproxy, key, val)
	}
}

func (gs *GateService) handleClearClientFilterProps(clientproxy *ClientProxy, packet *netutil.Packet) {
	gwlog.Debug("%s.handleClearClientFilterProps: clientproxy=%s", gs, clientproxy)
	clientid := clientproxy.clientid

	gs.filterTreesLock.Lock()

	for key, val := range clientproxy.filterProps {
		ft, ok := gs.filterTrees[key]
		if !ok {
			continue
		}
		ft.Remove(clientid, val)
	}
	gs.filterTreesLock.Unlock()

	if consts.DEBUG_FILTER_PROP {
		gwlog.Debug("CLEAR CLIENT %s FILTER PROPS", clientproxy)
	}
}

func (gs *GateService) handleCallFilteredClientProxies(packet *netutil.Packet) {
	key := packet.ReadVarStr()
	val := packet.ReadVarStr()

	gs.filterTreesLock.Lock()
	gs.clientProxiesLock.RLock()

	ft := gs.filterTrees[key]
	if ft != nil {
		ft.Visit(val, func(clientid common.ClientID) {
			//// visit all clientids and
			clientproxy := gs.clientProxies[clientid]
			if clientproxy != nil {
				clientproxy.SendPacket(packet)
			}
		})
	}

	gs.clientProxiesLock.RUnlock()
	gs.filterTreesLock.Unlock()
}

func (gs *GateService) handlePacketRoutine() {
	for {
		item := gs.packetQueue.Pop().(packetQueueItem)
		op := opmon.StartOperation("GateServiceHandlePacket")
		gs.HandleDispatcherClientPacket(item.msgtype, item.packet)
		op.Finish(time.Millisecond * 100)
		item.packet.Release()
	}
}
