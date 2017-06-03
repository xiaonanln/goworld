package goworld_startup

import (
	"net"

	"github.com/xiaonanln/goworld/netutil"
)

func Startup() {
	netutil.ServeTCPForever("127.0.0.1:4000", &serverDelegate{})
}

type serverDelegate struct {
}

func (sd *serverDelegate) ServeTCPConnection(net.Conn) {

}
