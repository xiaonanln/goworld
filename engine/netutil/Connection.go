package netutil

import "net"

// Connection interface for connections to servers
type Connection interface {
	net.Conn // Connection is more than net.Conn
	Flush() error
}

// NetConnection converts net.Conn to Connection
type NetConnection struct {
	net.Conn
}

// Flush flushes network connection
func (c NetConnection) Flush() error {
	return nil
}
