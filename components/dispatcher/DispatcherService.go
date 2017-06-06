package main

import (
	"fmt"

	"net"

	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/netutil"
)

type DispatcherService struct {
	config *config.DispatcherConfig
}

func newDispatcherService(cfg *config.DispatcherConfig) *DispatcherService {
	return &DispatcherService{
		config: cfg,
	}
}

func (ds *DispatcherService) run() {
	host := fmt.Sprintf("%s:%d", ds.config.Ip, ds.config.Port)
	netutil.ServeTCPForever(host, ds)
}

func (ds *DispatcherService) ServeTCPConnection(conn net.Conn) {
	client := newDispatcherClientProxy(conn)
	client.serve()
}
