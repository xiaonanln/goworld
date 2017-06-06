package main

import (
	"net"
	"time"

	"github.com/xiaonanln/goTimer"
)

type DispatcherClientProxy struct {
	net.Conn
}

func newDispatcherClientProxy(conn net.Conn) *DispatcherClientProxy {
	return &DispatcherClientProxy{conn}
}

func (dcp *DispatcherClientProxy) serve() {
	timer.AddCallback(time.Second, func() {
		dcp.Close()
	})
}
