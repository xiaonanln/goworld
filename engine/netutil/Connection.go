package netutil

import (
	"github.com/xiaonanln/netconnutil"
	"net"
)

type Connection interface {
	netconnutil.FlushableConn
}

type NetConn struct {
	net.Conn
}

func (n NetConn) Flush() error {
	return nil
}
