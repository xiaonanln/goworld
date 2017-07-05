package server

import (
	"net"
	"time"

	"fmt"

	"os"

	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/components/dispatcher/dispatcher_client"
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/netutil"
	"github.com/xiaonanln/goworld/proto"
)

type ClientProxy struct {
	proto.GoWorldConnection
	clientid    common.ClientID
	filterProps map[string]string
}

func newClientProxy(conn net.Conn) *ClientProxy {
	tcpConn := conn.(*net.TCPConn)
	tcpConn.SetWriteBuffer(consts.CLIENT_PROXY_WRITE_BUFFER_SIZE)
	tcpConn.SetReadBuffer(consts.CLIENT_PROXY_READ_BUFFER_SIZE)
	return &ClientProxy{
		GoWorldConnection: proto.NewGoWorldConnection(netutil.NewBufferedConnection(conn, time.Millisecond*50)), // using buffered connection for client proxy
		//GoWorldConnection: proto.NewGoWorldConnection(conn, false),
		clientid:    common.GenClientID(), // each client has its unique clientid
		filterProps: map[string]string{},
	}
}

func (cp *ClientProxy) String() string {
	return fmt.Sprintf("ClientProxy<%s@%s>", cp.clientid, cp.RemoteAddr())
}

func (cp *ClientProxy) serve() {
	defer func() {
		cp.Close()
		// tell the gate service that this client is down
		gateService.onClientProxyClose(cp)
		if err := recover(); err != nil && !netutil.IsConnectionClosed(err) {
			gwlog.Error("%s error: %s", cp, err)
		} else {
			gwlog.Info("%s disconnected", cp)
		}
	}()

	for {
		var msgtype proto.MsgType_t
		pkt, err := cp.Recv(&msgtype)
		if err != nil {
			panic(err)
		}

		if msgtype == proto.MT_CALL_ENTITY_METHOD_FROM_CLIENT {
			cp.handleCallEntityMethodFromClient(pkt)
		} else {
			gwlog.Panicf("unknown message type from client: %d", msgtype)
			if consts.DEBUG_MODE {
				os.Exit(2)
			}
		}

		pkt.Release()
	}
}

func (cp *ClientProxy) handleCallEntityMethodFromClient(pkt *netutil.Packet) {
	pkt.AppendClientID(cp.clientid) // append clientid to the packet
	dispatcher_client.GetDispatcherClientForSend().SendPacket(pkt)
}
