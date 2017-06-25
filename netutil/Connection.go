package netutil

import (
	"io"
	"net"
)

type Connection interface {
	io.ReadWriteCloser
	RemoteAddr() net.Addr
	LocalAddr() net.Addr
}
