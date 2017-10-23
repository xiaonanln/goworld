package main

import (
	"fmt"
	"time"

	"golang.org/x/net/websocket"

	"net"

	"sync"

	"os"

	"crypto/tls"

	"path"

	"github.com/pkg/errors"
	"github.com/xiaonanln/go-xnsyncutil/xnsyncutil"
	"github.com/xiaonanln/goworld/components/dispatcher/dispatcherclient"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/gwutils"
	"github.com/xiaonanln/goworld/engine/netutil"
	"github.com/xiaonanln/goworld/engine/opmon"
	"github.com/xiaonanln/goworld/engine/proto"
	"github.com/xtaci/kcp-go"
)

// GateService implements the gate service logic
type GateService struct {
	listenAddr        string
	clientProxies     map[common.ClientID]*ClientProxy
	clientProxiesLock sync.RWMutex
	packetQueue       *xnsyncutil.SyncQueue

	filterTreesLock sync.Mutex
	filterTrees     map[string]*_FilterTree

	pendingSyncPackets     []*netutil.Packet
	pendingSyncPacketsLock sync.Mutex

	terminating xnsyncutil.AtomicBool
	terminated  *xnsyncutil.OneTimeCond
	tlsConfig   *tls.Config
}

func newGateService() *GateService {
	return &GateService{
		//packetQueue: make(chan packetQueueItem, consts.DISPATCHER_CLIENT_PACKET_QUEUE_SIZE),
		clientProxies:      map[common.ClientID]*ClientProxy{},
		packetQueue:        xnsyncutil.NewSyncQueue(),
		filterTrees:        map[string]*_FilterTree{},
		pendingSyncPackets: []*netutil.Packet{},
		terminated:         xnsyncutil.NewOneTimeCond(),
	}
}

func (gs *GateService) run() {
	cfg := config.GetGate(gateid)
	gwlog.Infof("Compress connection: %v, encrypt connection: %v", cfg.CompressConnection, cfg.EncryptConnection)

	if cfg.EncryptConnection {
		gs.setupTLSConfig(cfg)
	}

	gs.listenAddr = fmt.Sprintf("%s:%d", cfg.Ip, cfg.Port)
	go netutil.ServeTCPForever(gs.listenAddr, gs)
	go gs.serveKCP(gs.listenAddr)
	gwutils.RepeatUntilPanicless(gs.handlePacketRoutine)
}

func (gs *GateService) setupTLSConfig(cfg *config.GateConfig) {
	cfgdir := config.GetConfigDir()
	rsaCert := path.Join(cfgdir, cfg.RSACertificate)
	rsaKey := path.Join(cfgdir, cfg.RSAKey)
	cert, err := tls.LoadX509KeyPair(rsaCert, rsaKey)
	if err != nil {
		gwlog.Panic(errors.Wrap(err, "load RSA key & certificate failed"))
	}

	gs.tlsConfig = &tls.Config{
		//MinVersion:       tls.VersionTLS12,
		//CurvePreferences: []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		Certificates: []tls.Certificate{cert},
		//CipherSuites: []uint16{
		//	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		//	tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		//	tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		//	tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		//},
		//PreferServerCipherSuites: true,
	}
}

func (gs *GateService) String() string {
	return fmt.Sprintf("GateService<%s>", gs.listenAddr)
}

// ServeTCPConnection handle TCP connections from clients
func (gs *GateService) ServeTCPConnection(conn net.Conn) {
	tcpConn := conn.(*net.TCPConn)
	tcpConn.SetWriteBuffer(consts.CLIENT_PROXY_WRITE_BUFFER_SIZE)
	tcpConn.SetReadBuffer(consts.CLIENT_PROXY_READ_BUFFER_SIZE)
	tcpConn.SetNoDelay(consts.CLIENT_PROXY_SET_TCP_NO_DELAY)

	gs.handleClientConnection(conn, false)
}

func (gs *GateService) serveKCP(addr string) {
	kcpListener, err := kcp.ListenWithOptions(addr, nil, 10, 3)
	if err != nil {
		gwlog.Panic(err)
	}

	gwlog.Infof("Listening on KCP: %s ...", addr)

	gwutils.RepeatUntilPanicless(func() {
		for {
			conn, err := kcpListener.AcceptKCP()
			if err != nil {
				gwlog.Panic(err)
			}
			go gs.handleKCPConn(conn)
		}
	})
}

func (gs *GateService) handleKCPConn(conn *kcp.UDPSession) {
	gwlog.Infof("KCP connection from %s", conn.RemoteAddr())

	conn.SetReadBuffer(consts.CLIENT_PROXY_READ_BUFFER_SIZE)
	conn.SetWriteBuffer(consts.CLIENT_PROXY_WRITE_BUFFER_SIZE)
	// turn on turbo mode according to https://github.com/skywind3000/kcp/blob/master/README.en.md#protocol-configuration
	conn.SetStreamMode(true)
	conn.SetWriteDelay(true)
	conn.SetNoDelay(1, 10, 2, 1)
	gs.handleClientConnection(conn, false)
}

func (gs *GateService) handleWebSocketConn(wsConn *websocket.Conn) {
	gwlog.Debugf("WebSocket Connection: %s", wsConn.RemoteAddr())
	//var conn netutil.Connection = NewWebSocketConn(wsConn)
	wsConn.PayloadType = websocket.BinaryFrame
	gs.handleClientConnection(wsConn, true)
}

func (gs *GateService) handleClientConnection(netconn net.Conn, isWebSocket bool) {
	if gs.terminating.Load() {
		// server terminating, not accepting more connectionsF
		netconn.Close()
		return
	}

	cfg := config.GetGate(gateid)

	if cfg.EncryptConnection && !isWebSocket {
		tlsConn := tls.Server(netconn, gs.tlsConfig)
		netconn = net.Conn(tlsConn)
	}

	conn := netutil.NetConnection{netconn}
	cp := newClientProxy(conn, cfg)

	gs.clientProxiesLock.Lock()
	gs.clientProxies[cp.clientid] = cp
	gs.clientProxiesLock.Unlock()

	dispatcherclient.GetDispatcherClientForSend().SendNotifyClientConnected(cp.clientid)
	if consts.DEBUG_CLIENTS {
		gwlog.Debugf("%s.ServeTCPConnection: client %s connected", gs, cp)
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
				gwlog.Debugf("DROP CLIENT %s FILTER PROP: %s = %s", cp, key, val)
			}
			ft.Remove(cp.clientid, val)
		}
	}
	gs.filterTreesLock.Unlock()

	dispatcherclient.GetDispatcherClientForSend().SendNotifyClientDisconnected(cp.clientid)
	if consts.DEBUG_CLIENTS {
		gwlog.Debugf("%s.onClientProxyClose: client %s disconnected", gs, cp)
	}
}

// HandleDispatcherClientPacket handles packets received by dispatcher client
func (gs *GateService) HandleDispatcherClientPacket(msgtype proto.MsgType, packet *netutil.Packet) {
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.HandleDispatcherClientPacket: msgtype=%v, packet(%d)=%v", gs, msgtype, packet.GetPayloadLen(), packet.Payload())
	}

	if msgtype >= proto.MT_REDIRECT_TO_GATEPROXY_MSG_TYPE_START && msgtype <= proto.MT_REDIRECT_TO_GATEPROXY_MSG_TYPE_STOP {
		_ = packet.ReadUint16() // gid
		clientid := packet.ReadClientID()

		gs.clientProxiesLock.RLock()
		clientproxy := gs.clientProxies[clientid]
		gs.clientProxiesLock.RUnlock()

		if clientproxy != nil {
			if msgtype == proto.MT_SET_CLIENTPROXY_FILTER_PROP {
				gs.handleSetClientFilterProp(clientproxy, packet)
			} else if msgtype == proto.MT_CLEAR_CLIENTPROXY_FILTER_PROPS {
				gs.handleClearClientFilterProps(clientproxy, packet)
			} else {
				// message types that should be redirected to client proxy
				clientproxy.SendPacket(packet)
			}
		} else {
			// client already disconnected, but the game service seems not knowing it, so tell it
			dispatcherclient.GetDispatcherClientForSend().SendNotifyClientDisconnected(clientid)
		}
	} else if msgtype == proto.MT_SYNC_POSITION_YAW_ON_CLIENTS {
		gs.handleSyncPositionYawOnClients(packet)
	} else if msgtype == proto.MT_CALL_FILTERED_CLIENTS {
		gs.handleCallFilteredClientProxies(packet)
	} else {
		gwlog.Panicf("%s: unknown msg type: %d", gs, msgtype)
		if consts.DEBUG_MODE {
			os.Exit(2)
		}
	}
}

func (gs *GateService) handleSetClientFilterProp(clientproxy *ClientProxy, packet *netutil.Packet) {
	gwlog.Debugf("%s.handleSetClientFilterProp: clientproxy=%s", gs, clientproxy)
	key := packet.ReadVarStr()
	val := packet.ReadVarStr()
	clientid := clientproxy.clientid

	gs.filterTreesLock.Lock()
	ft, ok := gs.filterTrees[key]
	if !ok {
		ft = newFilterTree()
		gs.filterTrees[key] = ft
	}

	oldVal, ok := clientproxy.filterProps[key]
	if ok {
		if consts.DEBUG_FILTER_PROP {
			gwlog.Debugf("REMOVE CLIENT %s FILTER PROP: %s = %s", clientproxy, key, val)
		}
		ft.Remove(clientid, oldVal)
	}
	clientproxy.filterProps[key] = val
	ft.Insert(clientid, val)
	gs.filterTreesLock.Unlock()

	if consts.DEBUG_FILTER_PROP {
		gwlog.Debugf("SET CLIENT %s FILTER PROP: %s = %s", clientproxy, key, val)
	}
}

func (gs *GateService) handleClearClientFilterProps(clientproxy *ClientProxy, packet *netutil.Packet) {
	gwlog.Debugf("%s.handleClearClientFilterProps: clientproxy=%s", gs, clientproxy)
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
		gwlog.Debugf("CLEAR CLIENT %s FILTER PROPS", clientproxy)
	}
}

func (gs *GateService) handleSyncPositionYawOnClients(packet *netutil.Packet) {
	_ = packet.ReadUint16() // read useless gateid
	payload := packet.UnreadPayload()
	payloadLen := len(payload)
	dispatch := map[common.ClientID][]byte{}
	for i := 0; i < payloadLen; i += common.CLIENTID_LENGTH + common.ENTITYID_LENGTH + proto.SYNC_INFO_SIZE_PER_ENTITY {
		clientid := common.ClientID(payload[i : i+common.CLIENTID_LENGTH])
		data := payload[i+common.CLIENTID_LENGTH : i+common.CLIENTID_LENGTH+common.ENTITYID_LENGTH+proto.SYNC_INFO_SIZE_PER_ENTITY]
		dispatch[clientid] = append(dispatch[clientid], data...)
	}
	//fmt.Fprintf(os.Stderr, "(%d,%d)", payloadLen, len(dispatch))

	// multiple entity sync infos are received from game->dispatcher, gate need to dispatcher these infos to different clients
	gs.clientProxiesLock.RLock()

	for clientid, data := range dispatch {
		clientproxy := gs.clientProxies[clientid]
		if clientproxy != nil {
			packet := netutil.NewPacket()
			packet.AppendUint16(proto.MT_SYNC_POSITION_YAW_ON_CLIENTS)
			packet.AppendBytes(data)
			packet.SetNotCompress() // too many these packets, giveup compress to save time
			clientproxy.SendPacket(packet)
			packet.Release()
		}
	}

	gs.clientProxiesLock.RUnlock()
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

func (gs *GateService) handleSyncPositionYawFromClient(packet *netutil.Packet) {
	packet.AddRefCount(1)
	gs.pendingSyncPacketsLock.Lock()
	gs.pendingSyncPackets = append(gs.pendingSyncPackets, packet)
	gs.pendingSyncPacketsLock.Unlock()
	//eid := packet.ReadEntityID()
	//x := packet.ReadFloat32()
	//y := packet.ReadFloat32()
	//z := packet.ReadFloat32()
	//yaw := packet.ReadFloat32()
}

func (gs *GateService) handleDispatcherClientBeforeFlush() {
	gs.pendingSyncPacketsLock.Lock()
	pendingSyncPackets := gs.pendingSyncPackets
	gs.pendingSyncPackets = make([]*netutil.Packet, 0, len(pendingSyncPackets))
	gs.pendingSyncPacketsLock.Unlock()
	// merge all client sync packets, and send in one packet (to reduce dispatcher overhead)

	if len(pendingSyncPackets) == 0 {
		return
	}

	packet := pendingSyncPackets[0] // use the first packet for sending
	if len(packet.UnreadPayload()) != common.ENTITYID_LENGTH+proto.SYNC_INFO_SIZE_PER_ENTITY {
		gwlog.Panicf("%s.handleDispatcherClientBeforeFlush: entity sync info size should be %d, but received %d", gs, proto.SYNC_INFO_SIZE_PER_ENTITY, len(packet.UnreadPayload())-common.ENTITYID_LENGTH)
	}

	//gwlog.Infof("sycn packet payload len %d, unread %d", packet.GetPayloadLen(), len(packet.UnreadPayload()))
	for _, syncPkt := range pendingSyncPackets[1:] { // merge other packets to the first packet
		//gwlog.Infof("sycn packet unread %d", len(syncPkt.UnreadPayload()))
		packet.AppendBytes(syncPkt.UnreadPayload())
		syncPkt.Release()
	}
	dispatcherclient.GetDispatcherClientForSend().SendPacket(packet)
	packet.Release()
}

type packetQueueItem struct { // packet queue from dispatcher client
	msgtype proto.MsgType
	packet  *netutil.Packet
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

func (gs *GateService) terminate() {
	gs.terminating.Store(true)

	gs.clientProxiesLock.RLock()

	for _, cp := range gs.clientProxies { // close all connected clients when terminating
		cp.Close()
	}

	gs.clientProxiesLock.RUnlock()

	gs.terminated.Signal()
}
