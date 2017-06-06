package dispatcher_client

import "net"

type DispatcherClient struct {
}

func newDispatcherClient(conn net.Conn) *DispatcherClient {
	return &DispatcherClient{}
}
