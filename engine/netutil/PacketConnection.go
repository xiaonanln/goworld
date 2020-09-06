package netutil

import (
	"context"
	"fmt"
	"github.com/xiaonanln/pktconn"
	"net"
)

// PacketConnection is a connection that send and receive data packets upon a network stream connection
type PacketConnection pktconn.PacketConn

// NewPacketConnection creates a packet connection based on network connection
func NewPacketConnection(conn Connection, tag interface{}) *PacketConnection {
	config := pktconn.DefaultConfig()
	config.Tag = tag
	return (*PacketConnection)(pktconn.NewPacketConnWithConfig(context.TODO(), conn, config))
}

// NewPacket allocates a new packet (usually for sending)
func (pc *PacketConnection) NewPacket() *Packet {
	return NewPacket()
}

// SendPacket send packets to remote
func (pc *PacketConnection) SendPacket(packet *Packet) {
	(*pktconn.PacketConn)(pc).Send((*pktconn.Packet)(packet))
}

// RecvPacket receives the next packet
func (pc *PacketConnection) RecvChan(recvChan chan *pktconn.Packet) error {
	return (*pktconn.PacketConn)(pc).RecvChan(recvChan)
}

// Close the connection
func (pc *PacketConnection) Close() error {
	return (*pktconn.PacketConn)(pc).Close()
}

// RemoteAddr return the remote address
func (pc *PacketConnection) RemoteAddr() net.Addr {
	return (*pktconn.PacketConn)(pc).RemoteAddr()
}

// LocalAddr returns the local address
func (pc *PacketConnection) LocalAddr() net.Addr {
	return (*pktconn.PacketConn)(pc).LocalAddr()
}

func (pc *PacketConnection) String() string {
	return fmt.Sprintf("[%s >>> %s]", pc.LocalAddr(), pc.RemoteAddr())
}
