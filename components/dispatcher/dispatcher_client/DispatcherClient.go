package dispatcher_client

import (
	"net"

	"time"

	"sync/atomic"
	"unsafe"

	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/netutil"
	"github.com/xiaonanln/goworld/proto"
)

type DispatcherClient struct {
	*proto.GoWorldConnection
}

func newDispatcherClient(conn net.Conn) *DispatcherClient {
	gwc := proto.NewGoWorldConnection(netutil.NewBufferedReadConnection(netutil.NetConnection{conn}), false)
	gwc.SetAutoFlush(time.Millisecond * 10)
	return &DispatcherClient{
		GoWorldConnection: gwc,
	}
}

func (dc *DispatcherClient) connect() error {
	dispatcherConfig := config.GetDispatcher()
	conn, err := netutil.ConnectTCP(dispatcherConfig.Ip, dispatcherConfig.Port)
	if err != nil {
		return err
	}

	tcpConn := conn.(*net.TCPConn)
	tcpConn.SetReadBuffer(consts.DISPATCHER_CLIENT_READ_BUFFER_SIZE)
	tcpConn.SetWriteBuffer(consts.DISPATCHER_CLIENT_WRITE_BUFFER_SIZE)

	gwc := proto.NewGoWorldConnection(netutil.NewBufferedReadConnection(netutil.NetConnection{conn}), false)
	gwc.SetAutoFlush(time.Millisecond * 10)
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&dc.GoWorldConnection)), unsafe.Pointer(gwc)) // set underlying GWC atomically
	return nil
}

func (dc *DispatcherClient) assureConnected() {
	err = dc.connect()
	if err != nil {
		gwlog.Error("Connect to dispatcher failed: %s", err.Error())
		time.Sleep(LOOP_DELAY_ON_DISPATCHER_CLIENT_ERROR)
		return
	}
	dispatcherClientDelegate.OnDispatcherClientConnect(dispatcherClient, isReconnect)

	setDispatcherClient(dispatcherClient)
	isReconnect = true

	gwlog.Info("dispatcher_client: connected to dispatcher: %s", dispatcherClient)

}
