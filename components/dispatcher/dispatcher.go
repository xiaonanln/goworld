package main

import (
	"fmt"

	"net"

	"flag"

	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/netutil"
)

var (
	configFile = ""
)

func debuglog(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	gwlog.Debug("dispatcher: %s", s)
}

type DispatcherDelegate struct{}

func main() {
	flag.StringVar(&configFile, "c", config.DEFAULT_CONFIG_FILENAME, "config file")
	flag.Parse()

	config.LoadConfig(configFile)
	gwlog.SetLevel(gwlog.DEBUG)

	netutil.ServeTCPForever(config.GetConfig().Dispatcher.Host, &DispatcherDelegate{})
}

func (dd *DispatcherDelegate) ServeTCPConnection(conn net.Conn) {
	client_proxy.NewClientProxy(conn).Serve()
}
