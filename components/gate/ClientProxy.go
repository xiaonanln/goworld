package main

import (
	"fmt"

	"time"

	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/gwioutil"
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
}

func newClientProxy(conn netutil.Connection, cfg *config.GateConfig) *ClientProxy {
	gwc := proto.NewGoWorldConnection(netutil.NewBufferedConnection(conn), cfg.CompressConnection, cfg.CompressFormat)
	return &ClientProxy{
		GoWorldConnection: gwc,
		clientid:          common.GenClientID(), // each client has its unique clientid
		filterProps:       map[string]string{},
	}
}

func (cp *ClientProxy) String() string {
	return fmt.Sprintf("ClientProxy<%s@%s>", cp.clientid, cp.RemoteAddr())
}

//func (cp *ClientProxy) SendPacket(packet *netutil.Packet) error {
//	err := cp.GoWorldConnection.SendPacket(packet)
//	if err != nil {
//		return err
//	}
//	return cp.Flush("ClientProxy")
//}

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

	cp.SetAutoFlush(consts.CLIENT_PROXY_WRITE_FLUSH_INTERVAL)
	//cp.SendSetClientClientID(cp.cp) // set the cp on the client side

	for {
		var msgtype proto.MsgType
		pkt, err := cp.Recv(&msgtype)
		if pkt != nil {
			gateService.clientPacketQueue <- clientProxyMessage{cp, proto.Message{msgtype, pkt}}
		} else if err != nil && !gwioutil.IsTimeoutError(err) {
			if netutil.IsConnectionError(err) {
				break
			} else {
				panic(err)
			}
		}
	}
}
