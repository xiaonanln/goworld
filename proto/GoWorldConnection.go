package proto

import (
	"net"

	. "github.com/xiaonanln/goworld/common"

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

func (gwc *GoWorldConnection) SendSetServerID(id int) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_SET_SERVER_ID)
	packet.AppendUint16(uint16(id))
	return gwc.SendPacketRelease(packet)
}

func (gwc *GoWorldConnection) SendNotifyCreateEntity(id EntityID) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_CREATE_ENTITY)
	packet.AppendEntityID(id)
	return gwc.SendPacketRelease(packet)
}

func (gwc *GoWorldConnection) SendNotifyNewClient(id ClientID) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_CLIENT_CONNECTED)
	packet.AppendClientID(id)
	return gwc.SendPacketRelease(packet)
}

func (gwc *GoWorldConnection) SendCreateEntityAnywhere(typeName string) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CREATE_ENTITY_ANYWHERE)
	packet.AppendVarStr(typeName)
	return gwc.SendPacketRelease(packet)
}

func (gwc *GoWorldConnection) SendLoadEntityAnywhere(typeName string, entityID EntityID) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_LOAD_ENTITY_ANYWHERE)
	packet.AppendVarStr(typeName)
	packet.AppendEntityID(entityID)
	return gwc.SendPacketRelease(packet)
}

func (gwc *GoWorldConnection) SendDeclareService(id EntityID, serviceName string) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_DECLARE_SERVICE)
	packet.AppendEntityID(id)
	packet.AppendVarStr(serviceName)
	return gwc.SendPacketRelease(packet)
}

func (gwc *GoWorldConnection) SendCallEntityMethod(id EntityID, method string, args []interface{}) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CALL_ENTITY_METHOD)
	packet.AppendEntityID(id)
	packet.AppendVarStr(method)
	packet.AppendMessage(args)
	return gwc.SendPacketRelease(packet)
}

func (gwc *GoWorldConnection) SendCallEntityMethodFromClient(id EntityID, method string, args []interface{}) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CALL_ENTITY_METHOD_FROM_CLIENT)
	packet.AppendEntityID(id)
	packet.AppendVarStr(method)
	packet.AppendMessage(args)
	return gwc.SendPacketRelease(packet)
}

func (gwc *GoWorldConnection) SendCreateEntityOnClient(sid uint16, clientid ClientID, typeName string, entityid EntityID) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CREATE_ENTITY_ON_CLIENT)
	packet.AppendUint16(sid)
	packet.AppendClientID(clientid)
	packet.AppendVarStr(typeName)
	packet.AppendEntityID(entityid)
	return gwc.SendPacketRelease(packet)
}

func (gwc *GoWorldConnection) SendDestroyEntityOnClient(sid uint16, clientid ClientID, typeName string, entityid EntityID) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_DESTROY_ENTITY_ON_CLIENT)
	packet.AppendUint16(sid)
	packet.AppendClientID(clientid)
	packet.AppendVarStr(typeName)
	packet.AppendEntityID(entityid)
	return gwc.SendPacketRelease(packet)
}

//func (gwc *GoWorldConnection) SendDeclareServiceReply(id EntityID, serviceName string, success bool) error {
//	packet := gwc.packetConn.NewPacket()
//	packet.AppendUint16(MT_DECLARE_SERVICE_REPLY)
//	packet.AppendEntityID(id)
//	packet.AppendVarStr(serviceName)
//	packet.AppendBool(success)
//	return gwc.packetConn.SendPacket(packet)
//}

func (gwc *GoWorldConnection) SendPacket(pkt *netutil.Packet) error {
	return gwc.packetConn.SendPacket(pkt)
}

func (gwc *GoWorldConnection) SendPacketRelease(pkt *netutil.Packet) error {
	err := gwc.packetConn.SendPacket(pkt)
	pkt.Release()
	return err
}

//func (gwc *GoWorldConnection) RecvPacket() (*netutil.Packet, error) {
//	return gwc.packetConn.RecvPacket()
//}

func (gwc *GoWorldConnection) Recv(msgtype *MsgType_t) (*netutil.Packet, error) {
	pkt, err := gwc.packetConn.RecvPacket()
	if err != nil {
		return nil, err
	}

	*msgtype = MsgType_t(pkt.ReadUint16())
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
