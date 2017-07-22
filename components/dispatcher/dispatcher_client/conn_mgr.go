package dispatcher_client

import (
	"time"

	"sync/atomic"

	"unsafe"

	"net"

	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/netutil"
	"github.com/xiaonanln/goworld/proto"
)

const (
	LOOP_DELAY_ON_DISPATCHER_CLIENT_ERROR = time.Second
)

var (
	isReconnect               = false
	_dispatcherClient         *DispatcherClient // DO NOT access it directly
	dispatcherClientDelegate  IDispatcherClientDelegate
	errDispatcherNotConnected = errors.New("dispatcher not connected")
)

func getDispatcherClient() *DispatcherClient { // atomic
	addr := (*uintptr)(unsafe.Pointer(&_dispatcherClient))
	return (*DispatcherClient)(unsafe.Pointer(atomic.LoadUintptr(addr)))
}

func setDispatcherClient(dc *DispatcherClient) { // atomic
	addr := (*uintptr)(unsafe.Pointer(&_dispatcherClient))
	atomic.StoreUintptr(addr, uintptr(unsafe.Pointer(dc)))
}

func assureConnectedDispatcherClient() *DispatcherClient {
	var err error
	dispatcherClient := getDispatcherClient()
	//gwlog.Debug("assureConnectedDispatcherClient: _dispatcherClient", _dispatcherClient)
	for dispatcherClient == nil || dispatcherClient.IsClosed() {
		dispatcherClient, err = connectDispatchClient()
		if err != nil {
			gwlog.Error("Connect to dispatcher failed: %s", err.Error())
			time.Sleep(LOOP_DELAY_ON_DISPATCHER_CLIENT_ERROR)
			continue
		}
		dispatcherClientDelegate.OnDispatcherClientConnect(dispatcherClient, isReconnect)

		setDispatcherClient(dispatcherClient)
		isReconnect = true

		gwlog.Info("dispatcher_client: connected to dispatcher: %s", dispatcherClient)
	}

	return dispatcherClient
}

func connectDispatchClient() (*DispatcherClient, error) {
	dispatcherConfig := config.GetDispatcher()
	conn, err := netutil.ConnectTCP(dispatcherConfig.Ip, dispatcherConfig.Port)
	if err != nil {
		return nil, err
	}
	tcpConn := conn.(*net.TCPConn)
	tcpConn.SetReadBuffer(consts.DISPATCHER_CLIENT_READ_BUFFER_SIZE)
	tcpConn.SetWriteBuffer(consts.DISPATCHER_CLIENT_WRITE_BUFFER_SIZE)
	return newDispatcherClient(conn), nil
}

type IDispatcherClientDelegate interface {
	OnDispatcherClientConnect(dispatcherClient *DispatcherClient, isReconnect bool)
	HandleDispatcherClientPacket(msgtype proto.MsgType_t, packet *netutil.Packet)
	//HandleDeclareService(entityID common.EntityID, serviceName string)
	//HandleCallEntityMethod(entityID common.EntityID, method string, args []interface{})
}

func Initialize(delegate IDispatcherClientDelegate) {
	dispatcherClientDelegate = delegate

	assureConnectedDispatcherClient()
	go netutil.ServeForever(serveDispatcherClient) // start the recv routine
}

func GetDispatcherClientForSend() *DispatcherClient {
	dispatcherClient := getDispatcherClient()
	return dispatcherClient
}

// serve the dispatcher client, receive RESPs from dispatcher and process
func serveDispatcherClient() {
	gwlog.Debug("serveDispatcherClient: start serving dispatcher client ...")
	for {
		dispatcherClient := assureConnectedDispatcherClient()
		var msgtype proto.MsgType_t
		pkt, err := dispatcherClient.Recv(&msgtype)

		if err != nil {
			if netutil.IsTemporaryNetError(err) {
				continue
			}

			gwlog.TraceError("serveDispatcherClient: RecvMsgPacket error: %s", err.Error())
			dispatcherClient.Close()
			time.Sleep(LOOP_DELAY_ON_DISPATCHER_CLIENT_ERROR)
			continue
		}

		if consts.DEBUG_PACKETS {
			gwlog.Debug("%s.RecvPacket: msgtype=%v, payload=%v", dispatcherClient, msgtype, pkt.Payload())
		}
		dispatcherClientDelegate.HandleDispatcherClientPacket(msgtype, pkt)
	}
}
