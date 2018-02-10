package main

import (
	"fmt"
	"time"

	"golang.org/x/net/websocket"

	"net"

	"os"

	"crypto/tls"

	"path"

	"github.com/pkg/errors"
	"github.com/xiaonanln/go-xnsyncutil/xnsyncutil"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/dispatchercluster"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/gwutils"
	"github.com/xiaonanln/goworld/engine/netutil"
	"github.com/xiaonanln/goworld/engine/opmon"
	"github.com/xiaonanln/goworld/engine/post"
	"github.com/xiaonanln/goworld/engine/proto"
	"github.com/xtaci/kcp-go"
)

type clientProxyMessage struct {
	cp  *ClientProxy
	msg proto.Message
}

// GateService implements the gate service logic
type GateService struct {
	listenAddr                  string
	clientProxies               map[common.ClientID]*ClientProxy
	dispatcherClientPacketQueue chan proto.Message
	clientPacketQueue           chan clientProxyMessage
	ticker                      <-chan time.Time

	filterTrees             map[string]*_FilterTree
	pendingSyncPackets      []*netutil.Packet
	nextFlushSyncTime       time.Time
	terminating             xnsyncutil.AtomicBool
	terminated              *xnsyncutil.OneTimeCond
	tlsConfig               *tls.Config
	checkHeartbeatsInterval time.Duration
	positionSyncInterval    time.Duration
}

func newGateService() *GateService {
	return &GateService{
		//dispatcherClientPacketQueue: make(chan packetQueueItem, consts.DISPATCHER_CLIENT_PACKET_QUEUE_SIZE),
		clientProxies:               map[common.ClientID]*ClientProxy{},
		dispatcherClientPacketQueue: make(chan proto.Message, consts.GATE_SERVICE_PACKET_QUEUE_SIZE),
		clientPacketQueue:           make(chan clientProxyMessage, consts.GATE_SERVICE_PACKET_QUEUE_SIZE),
		ticker:                      time.Tick(consts.GATE_SERVICE_TICK_INTERVAL),
		filterTrees:                 map[string]*_FilterTree{},
		pendingSyncPackets:          []*netutil.Packet{},
		terminated:                  xnsyncutil.NewOneTimeCond(),
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

	if cfg.HeartbeatCheckInterval > 0 {
		gs.checkHeartbeatsInterval = time.Second * time.Duration(cfg.HeartbeatCheckInterval)
		gwlog.Infof("%s: checkHeartbeatsInterval = %s", gs, gs.checkHeartbeatsInterval)
	}
	gs.positionSyncInterval = time.Millisecond * time.Duration(cfg.PositionSyncIntervalMS)
	gwlog.Infof("%s: positionSyncInterval = %s", gs, gs.positionSyncInterval)

	fmt.Fprintf(gwlog.GetOutput(), "%s\n", consts.GATE_STARTED_TAG)
	gwutils.RepeatUntilPanicless(gs.mainRoutine)
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
			gs.handleKCPConn(conn)
		}
	})
}

func (gs *GateService) handleKCPConn(conn *kcp.UDPSession) {
	gwlog.Infof("KCP connection from %s", conn.RemoteAddr())

	conn.SetReadBuffer(consts.CLIENT_PROXY_READ_BUFFER_SIZE)
	conn.SetWriteBuffer(consts.CLIENT_PROXY_WRITE_BUFFER_SIZE)
	// turn on turbo mode according to https://github.com/skywind3000/kcp/blob/master/README.en.md#protocol-configuration
	conn.SetNoDelay(consts.KCP_NO_DELAY, consts.KCP_INTERNAL_UPDATE_TIMER_INTERVAL, consts.KCP_ENABLE_FAST_RESEND, consts.KCP_DISABLE_CONGESTION_CONTROL)
	conn.SetStreamMode(consts.KCP_SET_STREAM_MODE)
	conn.SetWriteDelay(consts.KCP_SET_WRITE_DELAY)
	conn.SetACKNoDelay(consts.KCP_SET_ACK_NO_DELAY)

	gs.handleClientConnection(conn, false)
}

func (gs *GateService) handleWebSocketConn(wsConn *websocket.Conn) {
	gwlog.Debugf("WebSocket Connection: %s", wsConn.RemoteAddr())
	//var conn netutil.Connection = NewWebSocketConn(wsConn)
	wsConn.PayloadType = websocket.BinaryFrame
	gs.handleClientConnection(wsConn, true)
}

func (gs *GateService) handleClientConnection(netconn net.Conn, isWebSocket bool) {
	// this function might run in multiple threads
	if gs.terminating.Load() {
		// server terminating, not accepting more connections
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
	if consts.DEBUG_CLIENTS {
		gwlog.Debugf("%s.ServeTCPConnection: client %s connected", gs, cp)
	}

	// pass the client proxy to GateService ...
	post.Post(func() {
		gs.clientProxies[cp.clientid] = cp
		dispatchercluster.SelectByGateID(gateid).SendNotifyClientConnected(cp.clientid)
		go cp.serve()
	})
}

func (gs *GateService) checkClientHeartbeats() {
	now := time.Now()

	for _, cp := range gs.clientProxies { // close all connected clients when terminating
		if cp.heartbeatTime.Add(gs.checkHeartbeatsInterval).Before(now) {
			// 10 seconds no heartbeat, close it...
			gwlog.Infof("Connection %s timeout ...", cp)
			cp.Close()
		}
	}
}

func (gs *GateService) onClientProxyClose(cp *ClientProxy) {
	delete(gs.clientProxies, cp.clientid)

	for key, val := range cp.filterProps {
		ft := gs.filterTrees[key]
		if ft != nil {
			if consts.DEBUG_FILTER_PROP {
				gwlog.Debugf("DROP CLIENT %s FILTER PROP: %s = %s", cp, key, val)
			}
			ft.Remove(cp.clientid, val)
		}
	}

	dispatchercluster.SelectByGateID(gateid).SendNotifyClientDisconnected(cp.clientid)
	if consts.DEBUG_CLIENTS {
		gwlog.Debugf("%s.onClientProxyClose: client %s disconnected", gs, cp)
	}
}

// HandleDispatcherClientPacket handles packets received by dispatcher client
func (gs *GateService) handleClientProxyPacket(cp *ClientProxy, msgtype proto.MsgType, pkt *netutil.Packet) {
	cp.heartbeatTime = time.Now()

	if msgtype == proto.MT_SYNC_POSITION_YAW_FROM_CLIENT {
		cp.handleSyncPositionYawFromClient(pkt)
	} else if msgtype == proto.MT_CALL_ENTITY_METHOD_FROM_CLIENT {
		cp.handleCallEntityMethodFromClient(pkt)
	} else if msgtype == proto.MT_HEARTBEAT_FROM_CLIENT {
		// kcp connected from client, need to do nothing here

	} else {
		if consts.DEBUG_MODE {
			gwlog.TraceError("unknown message type from client: %d", msgtype)
			os.Exit(2)
		} else {
			gwlog.Panicf("unknown message type from client: %d", msgtype)
		}
	}

}

func (gs *GateService) handleDispatcherClientPacket(msgtype proto.MsgType, packet *netutil.Packet) {
	if consts.DEBUG_PACKETS {
		gwlog.Debugf("%s.handleDispatcherClientPacket: msgtype=%v, packet(%d)=%v", gs, msgtype, packet.GetPayloadLen(), packet.Payload())
	}

	if msgtype >= proto.MT_REDIRECT_TO_GATEPROXY_MSG_TYPE_START && msgtype <= proto.MT_REDIRECT_TO_GATEPROXY_MSG_TYPE_STOP {
		_ = packet.ReadUint16() // gid
		clientid := packet.ReadClientID()

		clientproxy := gs.clientProxies[clientid]

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
			dispatchercluster.SelectByGateID(gateid).SendNotifyClientDisconnected(clientid)
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

	if consts.DEBUG_FILTER_PROP {
		gwlog.Debugf("SET CLIENT %s FILTER PROP: %s = %s", clientproxy, key, val)
	}
}

func (gs *GateService) handleClearClientFilterProps(clientproxy *ClientProxy, packet *netutil.Packet) {
	gwlog.Debugf("%s.handleClearClientFilterProps: clientproxy=%s", gs, clientproxy)
	clientid := clientproxy.clientid

	for key, val := range clientproxy.filterProps {
		ft, ok := gs.filterTrees[key]
		if !ok {
			continue
		}
		ft.Remove(clientid, val)
	}

	if consts.DEBUG_FILTER_PROP {
		gwlog.Debugf("CLEAR CLIENT %s FILTER PROPS", clientproxy)
	}
}

func (gs *GateService) handleSyncPositionYawOnClients(packet *netutil.Packet) {
	_ = packet.ReadUint16() // read useless gateid
	payload := packet.UnreadPayload()
	payloadLen := len(payload)
	//gwlog.Infof("handleSyncPositionYawOnClients payloadLen=%v", payloadLen)
	dispatch := map[common.ClientID][]byte{}
	for i := 0; i < payloadLen; i += common.CLIENTID_LENGTH + common.ENTITYID_LENGTH + proto.SYNC_INFO_SIZE_PER_ENTITY {
		clientid := common.ClientID(payload[i : i+common.CLIENTID_LENGTH])
		data := payload[i+common.CLIENTID_LENGTH : i+common.CLIENTID_LENGTH+common.ENTITYID_LENGTH+proto.SYNC_INFO_SIZE_PER_ENTITY]
		dispatch[clientid] = append(dispatch[clientid], data...)
	}
	//fmt.Fprintf(os.Stderr, "(%d,%d)", payloadLen, len(dispatch))

	// multiple entity sync infos are received from game->dispatcher, gate need to dispatcher these infos to different clients

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
}

func (gs *GateService) handleCallFilteredClientProxies(packet *netutil.Packet) {
	key := packet.ReadVarStr()
	val := packet.ReadVarStr()

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

}

func (gs *GateService) handleSyncPositionYawFromClient(packet *netutil.Packet) {
	packet.AddRefCount(1)
	gs.pendingSyncPackets = append(gs.pendingSyncPackets, packet)
}

func (gs *GateService) tryFlushPendingSyncPackets() {
	now := time.Now()
	if now.Before(gs.nextFlushSyncTime) {
		return
	}

	gs.nextFlushSyncTime = now.Add(gs.positionSyncInterval)
	if len(gs.pendingSyncPackets) == 0 {
		return
	}

	pendingSyncPackets := gs.pendingSyncPackets
	// merge all client sync packets, and send in one packet (to reduce dispatcher overhead)

	packet := pendingSyncPackets[0] // use the first packet for sending
	if len(packet.UnreadPayload()) != common.ENTITYID_LENGTH+proto.SYNC_INFO_SIZE_PER_ENTITY {
		gwlog.Panicf("%s.tryFlushPendingSyncPackets: entity sync info size should be %d, but received %d", gs, proto.SYNC_INFO_SIZE_PER_ENTITY, len(packet.UnreadPayload())-common.ENTITYID_LENGTH)
	}

	//gwlog.Infof("sycn packet payload len %d, unread %d", packet.GetPayloadLen(), len(packet.UnreadPayload()))
	for i, syncPkt := range pendingSyncPackets[1:] { // merge other packets to the first packet
		//gwlog.Infof("sycn packet unread %d", len(syncPkt.UnreadPayload()))
		packet.AppendBytes(syncPkt.UnreadPayload())
		syncPkt.Release()
		pendingSyncPackets[i+1] = nil
	}
	dispatchercluster.SelectByGateID(gateid).SendPacket(packet)
	packet.Release()
	pendingSyncPackets[0] = nil

	gs.pendingSyncPackets = gs.pendingSyncPackets[:0] // reuse pendingSyncPackets capacity, but clear all packets
}

func (gs *GateService) mainRoutine() {
	for {
		select {
		case item := <-gs.clientPacketQueue:
			op := opmon.StartOperation("GateServiceHandlePacket")
			gs.handleClientProxyPacket(item.cp, item.msg.MsgType, item.msg.Packet)
			op.Finish(time.Millisecond * 100)
			item.msg.Packet.Release()
		case item := <-gs.dispatcherClientPacketQueue:
			op := opmon.StartOperation("GateServiceHandlePacket")
			gs.handleDispatcherClientPacket(item.MsgType, item.Packet)
			op.Finish(time.Millisecond * 100)
			item.Packet.Release()
			break
		case <-gs.ticker:
			gs.tryFlushPendingSyncPackets()
			break
		}

		post.Tick()
	}
}

func (gs *GateService) terminate() {
	gs.terminating.Store(true)

	for _, cp := range gs.clientProxies { // close all connected clients when terminating
		cp.Close()
	}

	gs.terminated.Signal()
}
