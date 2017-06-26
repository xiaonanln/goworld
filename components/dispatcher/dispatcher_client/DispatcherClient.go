package dispatcher_client

import (
	"net"

	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/netutil"
	"github.com/xiaonanln/goworld/proto"
)

type DispatcherClient struct {
	proto.GoWorldConnection
}

func newDispatcherClient(_conn net.Conn) *DispatcherClient {
	var conn netutil.Connection
	conn = _conn
	if consts.DISPATCHER_CLIENT_BUFFERED_DELAY > 0 {
		conn = netutil.NewBufferedConnection(conn, consts.DISPATCHER_CLIENT_BUFFERED_DELAY)
	}
	return &DispatcherClient{
		GoWorldConnection: proto.NewGoWorldConnection(conn, false),
	}
}
