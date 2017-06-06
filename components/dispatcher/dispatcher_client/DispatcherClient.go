package dispatcher_client

import (
	"net"

	"github.com/xiaonanln/goworld/proto"
)

type DispatcherClient struct {
	proto.GoWorldConnection
}

func newDispatcherClient(conn net.Conn) *DispatcherClient {
	return &DispatcherClient{
		GoWorldConnection: proto.NewGoWorldConnection(conn),
	}
}
