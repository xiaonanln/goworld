package dispatcherclient

import (
	"net"

	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/netutil"
	"github.com/xiaonanln/goworld/engine/proto"
)

type DispatcherClientType int

const (
	GameDispatcherClientType DispatcherClientType = 1 + iota
	GateDispatcherClientType
)

// DispatcherClient is a client connection to the dispatcher
type DispatcherClient struct {
	*proto.GoWorldConnection
	dctype        DispatcherClientType
	isReconnect   bool
	isRestoreGame bool
}

func newDispatcherClient(dctype DispatcherClientType, conn net.Conn, isReconnect bool, isRestoreGame bool) *DispatcherClient {
	gwc := proto.NewGoWorldConnection(netutil.NewBufferedConnection(netutil.NetConnection{conn}), false, "")
	if dctype != GameDispatcherClientType && dctype != GateDispatcherClientType {
		gwlog.Fatalf("invalid dispatcher client type: %v", dctype)
	}

	dc := &DispatcherClient{
		GoWorldConnection: gwc,
		dctype:            dctype,
		isReconnect:       isReconnect,
		isRestoreGame:     isRestoreGame,
	}
	return dc
}

//func (dc *DispatcherClient) StartAutoFlush(interval time.Duration, beforeFlushCallback func()) {
//	gwc := dc.GoWorldConnection
//	go func() {
//		//defer gwlog.Debugf("%s: auto flush routine quited", gwc)
//		for !gwc.IsClosed() {
//			time.Sleep(interval)
//			if beforeFlushCallback != nil {
//				beforeFlushCallback()
//			}
//
//			err := gwc.Flush("DispatcherClient")
//			if err != nil {
//				break
//			}
//		}
//	}()
//}

// Close the dispatcher client
func (dc *DispatcherClient) Close() error {
	return dc.GoWorldConnection.Close()
}
