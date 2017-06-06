package proto

import (
	"net"

	"github.com/xiaonanln/goworld/netutil"
)

type GoWorldConnection struct {
	packetConn netutil.PacketConnection
}

func NewGoWorldConnection(conn net.Conn) GoWorldConnection {
	return GoWorldConnection{
		packetConn: netutil.NewPacketConnection(conn),
	}
}

func (gwc *GoWorldConnection) SetGameID(id int) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_SET_GAME_ID)
	packet.AppendUint16(uint16(id))
	gwc.packetConn.SendPacket(packet)
}

func (gwc *GoWorldConnection) RecvPacket() (*netutil.Packet, error) {
	return gwc.packetConn.RecvPacket()
}

func (gwc *GoWorldConnection) Close() {
	gwc.packetConn.Close()
}

func (gwc *GoWorldConnection) RemoteAddr() net.Addr {
	return gwc.packetConn.RemoteAddr()
}

func (gwc *GoWorldConnection) LocalAddr() net.Addr {
	return gwc.packetConn.LocalAddr()
}

func (gwc *GoWorldConnection) String() string {
	return gwc.packetConn.String()
}
