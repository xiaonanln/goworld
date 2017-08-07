package proto

import (
	"net"

	. "github.com/xiaonanln/goworld/engine/common"

	"time"

	"github.com/xiaonanln/go-xnsyncutil/xnsyncutil"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/netutil"
)

// GoWorldConnection is the network protocol implementation of GoWorld components (dispatcher, gate, game)
type GoWorldConnection struct {
	packetConn *netutil.PacketConnection
	closed     xnsyncutil.AtomicBool
}

// NewGoWorldConnection creates a GoWorldConnection using network connection
func NewGoWorldConnection(conn netutil.Connection, compressed bool) *GoWorldConnection {
	return &GoWorldConnection{
		packetConn: netutil.NewPacketConnection(conn, compressed),
	}
}

// SendSetGameID sends MT_SET_GAME_ID message
func (gwc *GoWorldConnection) SendSetGameID(id uint16, isReconnect bool, isRestore bool) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_SET_GAME_ID)
	packet.AppendUint16(id)
	packet.AppendBool(isReconnect)
	packet.AppendBool(isRestore)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

// SendSetGateID sends MT_SET_GATE_ID message
func (gwc *GoWorldConnection) SendSetGateID(id uint16) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_SET_GATE_ID)
	packet.AppendUint16(id)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

// SendNotifyCreateEntity sends MT_NOTIFY_CREATE_ENTITY message
func (gwc *GoWorldConnection) SendNotifyCreateEntity(id EntityID) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_CREATE_ENTITY)
	packet.AppendEntityID(id)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

// SendNotifyDestroyEntity sends MT_NOTIFY_DESTROY_ENTITY message
func (gwc *GoWorldConnection) SendNotifyDestroyEntity(id EntityID) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_DESTROY_ENTITY)
	packet.AppendEntityID(id)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

// SendNotifyClientConnected sends MT_NOTIFY_CLIENT_CONNECTED message
func (gwc *GoWorldConnection) SendNotifyClientConnected(id ClientID) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_CLIENT_CONNECTED)
	packet.AppendClientID(id)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

// SendNotifyClientDisconnected sends MT_NOTIFY_CLIENT_DISCONNECTED message
func (gwc *GoWorldConnection) SendNotifyClientDisconnected(id ClientID) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_CLIENT_DISCONNECTED)
	packet.AppendClientID(id)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

// SendCreateEntityAnywhere sends MT_CREATE_ENTITY_ANYWHERE message
func (gwc *GoWorldConnection) SendCreateEntityAnywhere(typeName string, data map[string]interface{}) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CREATE_ENTITY_ANYWHERE)
	packet.AppendVarStr(typeName)
	packet.AppendData(data)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

// SendLoadEntityAnywhere sends MT_LOAD_ENTITY_ANYWHERE message
func (gwc *GoWorldConnection) SendLoadEntityAnywhere(typeName string, entityID EntityID) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_LOAD_ENTITY_ANYWHERE)
	packet.AppendEntityID(entityID)
	packet.AppendVarStr(typeName)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

// SendDeclareService sends MT_DECLARE_SERVICE message
func (gwc *GoWorldConnection) SendDeclareService(id EntityID, serviceName string) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_DECLARE_SERVICE)
	packet.AppendEntityID(id)
	packet.AppendVarStr(serviceName)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

// SendCallEntityMethod sends MT_CALL_ENTITY_METHOD message
func (gwc *GoWorldConnection) SendCallEntityMethod(id EntityID, method string, args []interface{}) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CALL_ENTITY_METHOD)
	packet.AppendEntityID(id)
	packet.AppendVarStr(method)
	packet.AppendArgs(args)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

// SendCallEntityMethodFromClient sends MT_CALL_ENTITY_METHOD_FROM_CLIENT message
func (gwc *GoWorldConnection) SendCallEntityMethodFromClient(id EntityID, method string, args []interface{}) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CALL_ENTITY_METHOD_FROM_CLIENT)
	packet.AppendEntityID(id)
	packet.AppendVarStr(method)
	packet.AppendArgs(args)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

// SendCreateEntityOnClient sends MT_CREATE_ENTITY_ON_CLIENT message
func (gwc *GoWorldConnection) SendCreateEntityOnClient(gid uint16, clientid ClientID, typeName string, entityid EntityID,
	isPlayer bool, clientData map[string]interface{}, x, y, z float32, yaw float32) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CREATE_ENTITY_ON_CLIENT)
	packet.AppendUint16(gid)
	packet.AppendClientID(clientid)
	packet.AppendBool(isPlayer)
	packet.AppendEntityID(entityid)
	packet.AppendVarStr(typeName)
	packet.AppendFloat32(x)
	packet.AppendFloat32(y)
	packet.AppendFloat32(z)
	packet.AppendFloat32(yaw)
	packet.AppendData(clientData)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

// SendSyncPositionYawFromClient sends MT_SYNC_POSITION_YAW_FROM_CLIENT message
func (gwc *GoWorldConnection) SendSyncPositionYawFromClient(entityID EntityID, x, y, z float32, yaw float32) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_SYNC_POSITION_YAW_FROM_CLIENT)
	packet.AppendEntityID(entityID)
	packet.AppendFloat32(x)
	packet.AppendFloat32(y)
	packet.AppendFloat32(z)
	packet.AppendFloat32(yaw)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

// SendSyncPositionOnClient sends MT_UPDATE_POSITION_ON_CLIENT message
func (gwc *GoWorldConnection) SendSyncPositionOnClient(gid uint16, clientid ClientID, entityID EntityID, x, y, z float32) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_UPDATE_POSITION_ON_CLIENT)
	packet.AppendUint16(gid)
	packet.AppendClientID(clientid)
	packet.AppendEntityID(entityID)
	packet.AppendFloat32(x)
	packet.AppendFloat32(y)
	packet.AppendFloat32(z)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

// SendUpdateYawOnClient sends MT_UPDATE_YAW_ON_CLIENT message
func (gwc *GoWorldConnection) SendUpdateYawOnClient(gid uint16, clientid ClientID, entityID EntityID, yaw float32) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_UPDATE_YAW_ON_CLIENT)
	packet.AppendUint16(gid)
	packet.AppendClientID(clientid)
	packet.AppendEntityID(entityID)
	packet.AppendFloat32(yaw)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

// SendDestroyEntityOnClient sends MT_DESTROY_ENTITY_ON_CLIENT message
func (gwc *GoWorldConnection) SendDestroyEntityOnClient(gid uint16, clientid ClientID, typeName string, entityid EntityID) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_DESTROY_ENTITY_ON_CLIENT)
	packet.AppendUint16(gid)
	packet.AppendClientID(clientid)
	packet.AppendVarStr(typeName)
	packet.AppendEntityID(entityid)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

// SendNotifyMapAttrChangeOnClient sends MT_NOTIFY_MAP_ATTR_CHANGE_ON_CLIENT message
func (gwc *GoWorldConnection) SendNotifyMapAttrChangeOnClient(gid uint16, clientid ClientID, entityid EntityID, path []interface{}, key string, val interface{}) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_MAP_ATTR_CHANGE_ON_CLIENT)
	packet.AppendUint16(gid)
	packet.AppendClientID(clientid)
	packet.AppendEntityID(entityid)
	packet.AppendData(path)
	packet.AppendVarStr(key)
	packet.AppendData(val)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

// SendNotifyMapAttrDelOnClient sends MT_NOTIFY_MAP_ATTR_DEL_ON_CLIENT message
func (gwc *GoWorldConnection) SendNotifyMapAttrDelOnClient(gid uint16, clientid ClientID, entityid EntityID, path []interface{}, key string) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_MAP_ATTR_DEL_ON_CLIENT)
	packet.AppendUint16(gid)
	packet.AppendClientID(clientid)
	packet.AppendEntityID(entityid)
	packet.AppendData(path)
	packet.AppendVarStr(key)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

// SendNotifyListAttrChangeOnClient sends MT_NOTIFY_LIST_ATTR_CHANGE_ON_CLIENT message
func (gwc *GoWorldConnection) SendNotifyListAttrChangeOnClient(gid uint16, clientid ClientID, entityid EntityID, path []interface{}, index uint32, val interface{}) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_LIST_ATTR_CHANGE_ON_CLIENT)
	packet.AppendUint16(gid)
	packet.AppendClientID(clientid)
	packet.AppendEntityID(entityid)
	packet.AppendData(path)
	packet.AppendUint32(index)
	packet.AppendData(val)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

// SendNotifyListAttrPopOnClient sends MT_NOTIFY_LIST_ATTR_POP_ON_CLIENT message
func (gwc *GoWorldConnection) SendNotifyListAttrPopOnClient(gid uint16, clientid ClientID, entityid EntityID, path []interface{}) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_LIST_ATTR_POP_ON_CLIENT)
	packet.AppendUint16(gid)
	packet.AppendClientID(clientid)
	packet.AppendEntityID(entityid)
	packet.AppendData(path)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

// SendNotifyListAttrAppendOnClient sends MT_NOTIFY_LIST_ATTR_APPEND_ON_CLIENT message
func (gwc *GoWorldConnection) SendNotifyListAttrAppendOnClient(gid uint16, clientid ClientID, entityid EntityID, path []interface{}, val interface{}) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_NOTIFY_LIST_ATTR_APPEND_ON_CLIENT)
	packet.AppendUint16(gid)
	packet.AppendClientID(clientid)
	packet.AppendEntityID(entityid)
	packet.AppendData(path)
	packet.AppendData(val)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

// SendCallEntityMethodOnClient sends MT_CALL_ENTITY_METHOD_ON_CLIENT message
func (gwc *GoWorldConnection) SendCallEntityMethodOnClient(gid uint16, clientid ClientID, entityID EntityID, method string, args []interface{}) (err error) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CALL_ENTITY_METHOD_ON_CLIENT)
	packet.AppendUint16(gid)
	packet.AppendClientID(clientid)
	packet.AppendEntityID(entityID)
	packet.AppendVarStr(method)
	packet.AppendArgs(args)
	err = gwc.SendPacket(packet)
	packet.Release()
	return
}

// SendSetClientFilterProp sends MT_SET_CLIENTPROXY_FILTER_PROP message
func (gwc *GoWorldConnection) SendSetClientFilterProp(gid uint16, clientid ClientID, key, val string) (err error) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_SET_CLIENTPROXY_FILTER_PROP)
	packet.AppendUint16(gid)
	packet.AppendClientID(clientid)
	packet.AppendVarStr(key)
	packet.AppendVarStr(val)
	err = gwc.SendPacket(packet)
	packet.Release()
	return
}

// SendClearClientFilterProp sends MT_CLEAR_CLIENTPROXY_FILTER_PROPS message
func (gwc *GoWorldConnection) SendClearClientFilterProp(gid uint16, clientid ClientID) (err error) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CLEAR_CLIENTPROXY_FILTER_PROPS)
	packet.AppendUint16(gid)
	packet.AppendClientID(clientid)
	err = gwc.SendPacket(packet)
	packet.Release()
	return
}

// SendCallFilterClientProxies sends MT_CALL_FILTERED_CLIENTS message
func (gwc *GoWorldConnection) SendCallFilterClientProxies(key string, val string, method string, args []interface{}) (err error) {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_CALL_FILTERED_CLIENTS)
	packet.AppendVarStr(key)
	packet.AppendVarStr(val)
	packet.AppendVarStr(method)
	packet.AppendArgs(args)
	err = gwc.SendPacket(packet)
	packet.Release()
	return
}

// SendMigrateRequest sends MT_MIGRATE_REQUEST message
func (gwc *GoWorldConnection) SendMigrateRequest(spaceID EntityID, entityID EntityID) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_MIGRATE_REQUEST)
	packet.AppendEntityID(entityID)
	packet.AppendEntityID(spaceID)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

// SendRealMigrate sends MT_REAL_MIGRATE message
func (gwc *GoWorldConnection) SendRealMigrate(eid EntityID, targetGame uint16, targetSpace EntityID, x, y, z float32,
	typeName string, migrateData map[string]interface{}, timerData []byte, clientid ClientID, clientsrv uint16) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_REAL_MIGRATE)
	packet.AppendEntityID(eid)
	packet.AppendUint16(targetGame)

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
	packet.AppendVarBytes(timerData)

	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

// SendStartFreezeGame sends MT_START_FREEZE_GAME message
func (gwc *GoWorldConnection) SendStartFreezeGame(gameid uint16) error {
	packet := gwc.packetConn.NewPacket()
	packet.AppendUint16(MT_START_FREEZE_GAME)
	err := gwc.SendPacket(packet)
	packet.Release()
	return err
}

// SendPacket send a packet to remote
func (gwc *GoWorldConnection) SendPacket(packet *netutil.Packet) error {
	return gwc.packetConn.SendPacket(packet)
}

// Flush connection writes
func (gwc *GoWorldConnection) Flush() error {
	return gwc.packetConn.Flush()
}

// SetAutoFlush starts a goroutine to flush connection writes at some specified interval
func (gwc *GoWorldConnection) SetAutoFlush(interval time.Duration) {
	go func() {
		defer gwlog.Debug("%s: auto flush routine quited", gwc)
		for !gwc.IsClosed() {
			time.Sleep(interval)
			err := gwc.Flush()
			if err != nil {
				break
			}
		}
	}()
}

// Recv receives the next packet and retrive the message type
func (gwc *GoWorldConnection) Recv(msgtype *MsgType) (*netutil.Packet, error) {
	pkt, err := gwc.packetConn.RecvPacket()
	if err != nil {
		return nil, err
	}

	*msgtype = MsgType(pkt.ReadUint16())
	return pkt, nil
}

// SetRecvDeadline set receive deadline
func (gwc *GoWorldConnection) SetRecvDeadline(deadline time.Time) error {
	return gwc.packetConn.SetRecvDeadline(deadline)
}

// Close this connection
func (gwc *GoWorldConnection) Close() error {
	gwc.closed.Store(true)
	return gwc.packetConn.Close()
}

// IsClosed returns if the connection is closed
func (gwc *GoWorldConnection) IsClosed() bool {
	return gwc.closed.Load()
}

// RemoteAddr returns the remote address
func (gwc *GoWorldConnection) RemoteAddr() net.Addr {
	return gwc.packetConn.RemoteAddr()
}

// LocalAddr returns the local addres
func (gwc *GoWorldConnection) LocalAddr() net.Addr {
	return gwc.packetConn.LocalAddr()
}

func (gwc *GoWorldConnection) String() string {
	return gwc.packetConn.String()
}
