package netutil

import (
	"io"
	"net"
	"time"
)

// Connection interface for connections to servers
type Connection interface {
	io.ReadWriteCloser
	Flush() error

	RemoteAddr() net.Addr
	LocalAddr() net.Addr
	SetWriteDeadline(time.Time) error
	SetReadDeadline(time.Time) error
}

// NetConnection converts net.Conn to Connection
type NetConnection struct {
	net.Conn
}

func (c NetConnection) Flush() error {
	return nil
}
