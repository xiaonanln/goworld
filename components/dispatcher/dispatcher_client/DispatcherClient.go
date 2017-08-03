package dispatcher_client

import (
	"net"

	"time"

	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/netutil"
	"github.com/xiaonanln/goworld/engine/proto"
)

type DispatcherClient struct {
	*proto.GoWorldConnection
}

func newDispatcherClient(conn net.Conn, autoFlush bool) *DispatcherClient {
	gwc := proto.NewGoWorldConnection(netutil.NewBufferedReadConnection(netutil.NetConnection{conn}), false)

	dc := &DispatcherClient{
		GoWorldConnection: gwc,
	}
	if autoFlush {
		go func() {
			defer gwlog.Debug("%s: auto flush routine quited", gwc)
			for !gwc.IsClosed() {
				time.Sleep(time.Millisecond * 10)
				dispatcherClientDelegate.HandleDispatcherClientBeforeFlush()

				err := gwc.Flush()
				if err != nil {
					break
				}
			}
		}()
	}
	return dc
}

func (dc *DispatcherClient) Close() error {
	return dc.GoWorldConnection.Close()
}
