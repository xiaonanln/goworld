package main

import "net"

type DispatcherClientProxy struct {
	net.Conn
}

func newDispatcherClientProxy(conn net.Conn) *DispatcherClientProxy {
	return &DispatcherClientProxy{conn}
}

func (dcp *DispatcherClientProxy) serve() {
	dcp.Close()
}
