package dispatcherclient

import (
	"net"

	"time"

	"github.com/xiaonanln/goworld/engine/netutil"
	"github.com/xiaonanln/goworld/engine/proto"
)

// DispatcherClient is a client connection to the dispatcher
type DispatcherClient struct {
	*proto.GoWorldConnection
}

func newDispatcherClient(conn net.Conn) *DispatcherClient {
	gwc := proto.NewGoWorldConnection(netutil.NewBufferedConnection(netutil.NetConnection{conn}), false, "")

	dc := &DispatcherClient{
		GoWorldConnection: gwc,
	}
	return dc
}

func (dc *DispatcherClient) StartAutoFlush(beforeFlushCallback func()) {
	gwc := dc.GoWorldConnection
	go func() {
		//defer gwlog.Debugf("%s: auto flush routine quited", gwc)
		for !gwc.IsClosed() {
			time.Sleep(time.Millisecond * 10)
			beforeFlushCallback()

			err := gwc.Flush("DispatcherClient")
			if err != nil {
				break
			}
		}
	}()
}

// Close the dispatcher client
func (dc *DispatcherClient) Close() error {
	return dc.GoWorldConnection.Close()
}
