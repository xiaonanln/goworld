package netutil

import (
	"io"
	"net"
	"time"
)

type Connection interface {
	io.ReadWriteCloser
	Flush() error

	RemoteAddr() net.Addr
	LocalAddr() net.Addr
	SetWriteDeadline(time.Time) error
	SetReadDeadline(time.Time) error
}

type NetConnection struct {
	net.Conn
}

func (c NetConnection) Flush() error {
	return nil
}
