package netutil

import (
	"io"
	"net"
	"time"
)

type Connection interface {
	io.ReadWriteCloser
	RemoteAddr() net.Addr
	LocalAddr() net.Addr
	SetWriteDeadline(time.Time) error
	SetReadDeadline(time.Time) error
}

type FlushableConnection interface {
	Connection
	Flush() error
}

type _NopFlushable struct {
	Connection
}

func (nf _NopFlushable) Flush() error {
	return nil
}

func NopFlushable(conn Connection) FlushableConnection {
	return &_NopFlushable{
		Connection: conn,
	}
}
