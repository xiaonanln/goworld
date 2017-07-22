package dispatcher_client

import (
	"time"

	"github.com/pkg/errors"
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
	dispatcherClient          DispatcherClient // DO NOT access it directly
	dispatcherClientDelegate  IDispatcherClientDelegate
	errDispatcherNotConnected = errors.New("dispatcher not connected")
)

type IDispatcherClientDelegate interface {
	OnDispatcherClientConnect(dispatcherClient *DispatcherClient, isReconnect bool)
	HandleDispatcherClientPacket(msgtype proto.MsgType_t, packet *netutil.Packet)
	//HandleDeclareService(entityID common.EntityID, serviceName string)
	//HandleCallEntityMethod(entityID common.EntityID, method string, args []interface{})
}

func Initialize(delegate IDispatcherClientDelegate) {
	dispatcherClientDelegate = delegate

	dispatcherClient.assureConnected()
	go netutil.ServeForever(serveDispatcherClient) // start the recv routine
}

func GetDispatcherClientForSend() *DispatcherClient {
	return &dispatcherClient
}

// serve the dispatcher client, receive RESPs from dispatcher and process
func serveDispatcherClient() {
	gwlog.Debug("serveDispatcherClient: start serving dispatcher client ...")
	for {
		dispatcherClient.assureConnected()

		var msgtype proto.MsgType_t
		pkt, err := dispatcherClient.Recv(&msgtype)

		if err != nil {
			if netutil.IsTemporaryNetError(err) {
				continue
			}

			gwlog.TraceError("serveDispatcherClient: RecvMsgPacket error: %s", err.Error())
			dispatcherClient.Close()
			dispatcherClientDelegate.HandleDispatcherClientDisconnect()
			time.Sleep(LOOP_DELAY_ON_DISPATCHER_CLIENT_ERROR)
			continue
		}

		if consts.DEBUG_PACKETS {
			gwlog.Debug("%s.RecvPacket: msgtype=%v, payload=%v", dispatcherClient, msgtype, pkt.Payload())
		}
		dispatcherClientDelegate.HandleDispatcherClientPacket(msgtype, pkt)
	}
}
