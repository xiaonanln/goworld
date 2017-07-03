package proto

import (
	"net"

	. "github.com/xiaonanln/goworld/common"

	"time"

	"github.com/xiaonanln/goworld/netutil"
	"github.com/xiaonanln/goworld/opmon"
)

type GoWorldConnection struct {
	packetConn netutil.PacketConnection
}

func NewGoWorldConnection(conn netutil.Connection, useSendQueue bool) GoWorldConnection {
	return GoWorldConnection{
		packetConn: netutil.NewPacketConnection(conn, useSendQueue),
	}
}

func (gwc *GoWorldConnection) SendSetServerID(id uint16, isReconnect bool) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_SET_SERVER_ID)
	packet.AppendUint16(id)
	packet.AppendBool(isReconnect)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

func (gwc *GoWorldConnection) SendNotifyCreateEntity(id EntityID) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_CREATE_ENTITY)
	packet.AppendEntityID(id)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}
func (gwc *GoWorldConnection) SendNotifyDestroyEntity(id EntityID) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_DESTROY_ENTITY)
	packet.AppendEntityID(id)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

func (gwc *GoWorldConnection) SendNotifyClientConnected(id ClientID) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_CLIENT_CONNECTED)
	packet.AppendClientID(id)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

func (gwc *GoWorldConnection) SendNotifyClientDisconnected(id ClientID) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_CLIENT_DISCONNECTED)
	packet.AppendClientID(id)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

func (gwc *GoWorldConnection) SendCreateEntityAnywhere(typeName string, data map[string]interface{}) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CREATE_ENTITY_ANYWHERE)
	packet.AppendVarStr(typeName)
	packet.AppendData(data)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

func (gwc *GoWorldConnection) SendLoadEntityAnywhere(typeName string, entityID EntityID) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_LOAD_ENTITY_ANYWHERE)
	packet.AppendEntityID(entityID)
	packet.AppendVarStr(typeName)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

func (gwc *GoWorldConnection) SendDeclareService(id EntityID, serviceName string) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_DECLARE_SERVICE)
	packet.AppendEntityID(id)
	packet.AppendVarStr(serviceName)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

func (gwc *GoWorldConnection) SendCallEntityMethod(id EntityID, method string, args []interface{}) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CALL_ENTITY_METHOD)
	packet.AppendEntityID(id)
	packet.AppendVarStr(method)
	packet.AppendData(args)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

func (gwc *GoWorldConnection) SendCallEntityMethodFromClient(id EntityID, method string, args []interface{}) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CALL_ENTITY_METHOD_FROM_CLIENT)
	packet.AppendEntityID(id)
	packet.AppendVarStr(method)
	packet.AppendData(args)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}
func (gwc *GoWorldConnection) SendCreateEntityOnClient(sid uint16, clientid ClientID, typeName string, entityid EntityID, isPlayer bool, clientData map[string]interface{}) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CREATE_ENTITY_ON_CLIENT)
	packet.AppendUint16(sid)
	packet.AppendClientID(clientid)
	packet.AppendBool(isPlayer)
	packet.AppendEntityID(entityid)
	packet.AppendVarStr(typeName)
	packet.AppendData(clientData)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

func (gwc *GoWorldConnection) SendDestroyEntityOnClient(sid uint16, clientid ClientID, typeName string, entityid EntityID) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_DESTROY_ENTITY_ON_CLIENT)
	packet.AppendUint16(sid)
	packet.AppendClientID(clientid)
	packet.AppendVarStr(typeName)
	packet.AppendEntityID(entityid)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

func (gwc *GoWorldConnection) SendNotifyAttrChangeOnClient(sid uint16, clientid ClientID, entityid EntityID, path []string, key string, val interface{}) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_ATTR_CHANGE_ON_CLIENT)
	packet.AppendUint16(sid)
	packet.AppendClientID(clientid)
	packet.AppendEntityID(entityid)
	packet.AppendStringList(path)
	packet.AppendVarStr(key)
	packet.AppendData(val)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

func (gwc *GoWorldConnection) SendNotifyAttrDelOnClient(sid uint16, clientid ClientID, entityid EntityID, path []string, key string) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_ATTR_DEL_ON_CLIENT)
	packet.AppendUint16(sid)
	packet.AppendClientID(clientid)
	packet.AppendEntityID(entityid)
	packet.AppendStringList(path)
	packet.AppendVarStr(key)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

func (gwc *GoWorldConnection) SendCallEntityMethodOnClient(sid uint16, clientid ClientID, entityID EntityID, method string, args []interface{}) (err error) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CALL_ENTITY_METHOD_ON_CLIENT)
	packet.AppendUint16(sid)
	packet.AppendClientID(clientid)
	packet.AppendEntityID(entityID)
	packet.AppendVarStr(method)
	packet.AppendData(args)
	err = gwc.SendPacket(packet)
	packet.Release()
	return
}

func (gwc *GoWorldConnection) SendSetClientFilterProp(sid uint16, clientid ClientID, key, val string) (err error) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_SET_CLIENTPROXY_FILTER_PROP)
	packet.AppendUint16(sid)
	packet.AppendClientID(clientid)
	packet.AppendVarStr(key)
	packet.AppendVarStr(val)
	err = gwc.SendPacket(packet)
	packet.Release()
	return
}

func (gwc *GoWorldConnection) SendClearClientFilterProp(sid uint16, clientid ClientID) (err error) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CLEAR_CLIENTPROXY_FILTER_PROPS)
	packet.AppendUint16(sid)
	packet.AppendClientID(clientid)
	err = gwc.SendPacket(packet)
	packet.Release()
	return
}

func (gwc *GoWorldConnection) SendCallFilterClientProxies(key string, val string, method string, args []interface{}) (err error) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CALL_FILTERED_CLIENTS)
	packet.AppendVarStr(key)
	packet.AppendVarStr(val)
	packet.AppendVarStr(method)
	packet.AppendData(args)
	err = gwc.SendPacket(packet)
	packet.Release()
	return
}

func (gwc *GoWorldConnection) SendMigrateRequest(spaceID EntityID, entityID EntityID) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_MIGRATE_REQUEST)
	packet.AppendEntityID(entityID)
	packet.AppendEntityID(spaceID)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

func (gwc *GoWorldConnection) SendRealMigrate(eid EntityID, targetServer uint16, targetSpace EntityID, x, y, z float32,
	typeName string, migrateData map[string]interface{}, clientid ClientID, clientsrv uint16) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_REAL_MIGRATE)
	packet.AppendEntityID(eid)
	packet.AppendUint16(targetServer)

	if !clientid.IsNil() {
		packet.AppendBool(true)
		packet.AppendClientID(clientid)
		packet.AppendUint16(clientsrv)
	} else {
		packet.AppendBool(false)
	}

	packet.AppendEntityID(targetSpace)
	packet.AppendFloat32(x)
	packet.AppendFloat32(y)
	packet.AppendFloat32(z)
	packet.AppendVarStr(typeName)
	packet.AppendData(migrateData)

	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

func (gwc *GoWorldConnection) SendPacket(pkt *netutil.Packet) error {
	op := opmon.StartOperation("SendPacket")
	err := gwc.packetConn.SendPacket(pkt)
	op.Finish(time.Millisecond * 10)
	return err
}

//func (gwc *GoWorldConnection) SendPacketRelease(pkt *netutil.Packet) error {
//	err := gwc.packetConn.SendPacket(pkt)
//	pkt.Release()
//	return err
//}

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
