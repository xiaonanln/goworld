package main

import (
	"fmt"
	"github.com/xiaonanln/netconnutil"
	"net"
	"time"

	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/netutil"
	"github.com/xiaonanln/goworld/engine/post"
	"github.com/xiaonanln/goworld/engine/proto"
)

type clientSyncInfo struct {
	EntityID common.EntityID
	X, Y, Z  float32
	Yaw      float32
}

func (info *clientSyncInfo) IsEmpty() bool {
	return info.EntityID == ""
}

// ClientProxy is a game client connections managed by gate
type ClientProxy struct {
	*proto.GoWorldConnection
	clientid       common.ClientID
	filterProps    map[string]string
	clientSyncInfo clientSyncInfo
	heartbeatTime  time.Time
	ownerEntityID  common.EntityID // owner entity's ID
}

func newClientProxy(_conn net.Conn, cfg *config.GateConfig) *ClientProxy {
	_conn = netconnutil.NewNoTempErrorConn(_conn)
	var conn netutil.Connection = netutil.NetConn{_conn}
	if cfg.CompressConnection {
		conn = netconnutil.NewSnappyConn(conn)
	}
	conn = netconnutil.NewBufferedConn(conn, consts.BUFFERED_READ_BUFFSIZE, consts.BUFFERED_WRITE_BUFFSIZE)
	clientProxy := &ClientProxy{
		clientid:    common.GenClientID(), // each client has its unique clientid
		filterProps: map[string]string{},
	}
	clientProxy.GoWorldConnection = proto.NewGoWorldConnection(conn, clientProxy)
	return clientProxy
}

func (cp *ClientProxy) String() string {
	return fmt.Sprintf("ClientProxy<%s@%s>", cp.clientid, cp.RemoteAddr())
}

func (cp *ClientProxy) serve() {
	defer func() {
		cp.Close()
		// tell the gate service that this client is down
		post.Post(func() {
			gateService.onClientProxyClose(cp)
		})

		if err := recover(); err != nil && !netutil.IsConnectionError(err.(error)) {
			gwlog.TraceError("%s error: %s", cp, err.(error))
		} else {
			gwlog.Debugf("%s disconnected", cp)
		}
	}()

	err := cp.RecvChan(gateService.clientPacketQueue)
	if err != nil {
		gwlog.Panic(err)
	}
}
