package dispatcherclient

import (
	"time"

	"net"

	"fmt"

	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/gwioutil"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/gwutils"
	"github.com/xiaonanln/goworld/engine/netutil"
	"github.com/xiaonanln/goworld/engine/proto"
)

const (
	_LOOP_DELAY_ON_DISPATCHER_CLIENT_ERROR = time.Second
)

type DispatcherConnMgr struct {
	dispid                   uint16
	dispatcherClient         *DispatcherClient
	isReconnect              bool
	autoFlush                bool
	dispatcherClientDelegate IDispatcherClientDelegate
}

var (
	errDispatcherNotConnected = errors.New("dispatcher not connected")
)

func NewDispatcherConnMgr(dispid uint16, delegate IDispatcherClientDelegate, autoFlush bool) *DispatcherConnMgr {
	return &DispatcherConnMgr{
		dispid:                   dispid,
		autoFlush:                autoFlush,
		dispatcherClientDelegate: delegate,
	}
}

//func getDispatcherClient() *DispatcherClient { // atomic
//	addr := (*uintptr)(unsafe.Pointer(&_dispatcherClient))
//	return (*DispatcherClient)(unsafe.Pointer(atomic.LoadUintptr(addr)))
//}
//
//func setDispatcherClient(dispatcherClient *DispatcherClient) { // atomic
//	addr := (*uintptr)(unsafe.Pointer(&_dispatcherClient))
//	atomic.StoreUintptr(addr, uintptr(unsafe.Pointer(dispatcherClient)))
//}

func (dcm *DispatcherConnMgr) String() string {
	return fmt.Sprintf("DispatcherConnMgr<%d>", dcm.dispid)
}

func (dcm *DispatcherConnMgr) assureConnectedDispatcherClient() {
	//gwlog.Debugf("assureConnectedDispatcherClient: _dispatcherClient", _dispatcherClient)
	for dcm.dispatcherClient == nil || dcm.dispatcherClient.IsClosed() {
		err := dcm.connectDispatchClient()
		if err != nil {
			gwlog.Errorf("Connect to dispatcher failed: %s", err.Error())
			time.Sleep(_LOOP_DELAY_ON_DISPATCHER_CLIENT_ERROR)
			continue
		}
		dcm.dispatcherClientDelegate.OnDispatcherClientConnect(dcm.isReconnect)
		dcm.isReconnect = true

		gwlog.Infof("dispatcher_client: connected to dispatcher: %s", dcm.dispatcherClient)
	}
}

func (dcm *DispatcherConnMgr) connectDispatchClient() error {
	dispatcherConfig := config.GetDispatcher(dcm.dispid)
	conn, err := netutil.ConnectTCP(dispatcherConfig.Ip, dispatcherConfig.Port)
	if err != nil {
		return err
	}
	tcpConn := conn.(*net.TCPConn)
	tcpConn.SetReadBuffer(consts.DISPATCHER_CLIENT_READ_BUFFER_SIZE)
	tcpConn.SetWriteBuffer(consts.DISPATCHER_CLIENT_WRITE_BUFFER_SIZE)
	dcm.dispatcherClient = newDispatcherClient(conn)
	if dcm.autoFlush {
		dcm.dispatcherClient.StartAutoFlush(dcm.dispatcherClientDelegate.HandleDispatcherClientBeforeFlush)
	}
	return nil
}

// IDispatcherClientDelegate defines functions that should be implemented by dispatcher clients
type IDispatcherClientDelegate interface {
	OnDispatcherClientConnect(isReconnect bool)
	HandleDispatcherClientPacket(msgtype proto.MsgType, packet *netutil.Packet)
	HandleDispatcherClientDisconnect()
	HandleDispatcherClientBeforeFlush()
	//HandleDeclareService(entityID common.EntityID, serviceName string)
	//HandleCallEntityMethod(entityID common.EntityID, method string, args []interface{})
}

// Initialize the dispatcher client, only called by engine
func (dcm *DispatcherConnMgr) Connect() {
	dcm.assureConnectedDispatcherClient()
	go gwutils.RepeatUntilPanicless(dcm.serveDispatcherClient) // start the recv routine
}

//// GetDispatcherClientForSend returns the current dispatcher client for sending messages
//func GetDispatcherClientForSend() *DispatcherClient {
//	dispatcherClient := getDispatcherClient()
//	return dispatcherClient
//}

// serve the dispatcher client, receive RESPs from dispatcher and process
func (dcm *DispatcherConnMgr) serveDispatcherClient() {
	gwlog.Debugf("%s.serveDispatcherClient: start serving dispatcher client ...", dcm)
	for {
		dcm.assureConnectedDispatcherClient()
		var msgtype proto.MsgType
		pkt, err := dcm.dispatcherClient.Recv(&msgtype)

		if err != nil {
			if gwioutil.IsTimeoutError(err) {
				continue
			}

			gwlog.TraceError("serveDispatcherClient: RecvMsgPacket error: %s", err.Error())
			dcm.dispatcherClient.Close()
			dcm.dispatcherClientDelegate.HandleDispatcherClientDisconnect()
			time.Sleep(_LOOP_DELAY_ON_DISPATCHER_CLIENT_ERROR)
			continue
		}

		if consts.DEBUG_PACKETS {
			gwlog.Debugf("%s.RecvPacket: msgtype=%v, payload=%v", dcm.dispatcherClient, msgtype, pkt.Payload())
		}
		dcm.dispatcherClientDelegate.HandleDispatcherClientPacket(msgtype, pkt)
	}
}
