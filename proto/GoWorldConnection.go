package proto

import (
	"net"

	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/netutil"
)

const (
	MSG_TYPE_SIZE = 2
)

type GoWorldConnection struct {
	packetConn netutil.PacketConnection
}

func NewGoWorldConnection(conn net.Conn) GoWorldConnection {
	return GoWorldConnection{
		packetConn: netutil.NewPacketConnection(conn),
	}
}

func (gwc *GoWorldConnection) SetGameID(id int) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_SET_GAME_ID)
	packet.AppendUint16(uint16(id))
	return gwc.packetConn.SendPacket(packet)
}

func (gwc *GoWorldConnection) NotifyCreateEntity(id entity.EntityID) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_CREATE_ENTITY)
	packet.AppendBytes([]byte(id))
	return gwc.packetConn.SendPacket(packet)
}

//func (gwc *GoWorldConnection) RecvPacket() (*netutil.Packet, error) {
//	return gwc.packetConn.RecvPacket()
//}

func (gwc *GoWorldConnection) Recv(msgtype *MsgType_t, data *[]byte) (*netutil.Packet, error) {
	pkt, err := gwc.packetConn.RecvPacket()
	if err != nil {
		return nil, err
	}

	payload := pkt.Payload()
	*msgtype = MsgType_t(netutil.PACKET_ENDIAN.Uint16(payload[:MSG_TYPE_SIZE]))
	*data = payload[MSG_TYPE_SIZE:]
	return pkt, nil
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
