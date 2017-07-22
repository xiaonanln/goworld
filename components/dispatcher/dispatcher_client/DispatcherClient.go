package dispatcher_client

import (
	"net"

	"time"

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

func (dc *DispatcherClient) Close() error {
	return dc.GoWorldConnection.Close()
}
